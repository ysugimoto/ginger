package request

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/ysugimoto/ginger/config"
	"github.com/ysugimoto/ginger/entity"
	"github.com/ysugimoto/ginger/logger"
)

func formatProxyPath(path string) string {
	path = strings.TrimLeft(path, "/")
	if n := strings.Index(path, "/{proxy+}"); n != -1 {
		path = path[0:n]
	}
	return "/" + path
}

type APIGatewayRequest struct {
	svc    *apigateway.APIGateway
	log    *logger.Logger
	config *config.Config
}

func NewAPIGateway(c *config.Config) *APIGatewayRequest {
	return &APIGatewayRequest{
		config: c,
		svc:    apigateway.New(createAWSSession(c)),
		log:    logger.WithNamespace("ginger.request.apigateway"),
	}
}

func (a *APIGatewayRequest) errorLog(err error, skipCodes ...string) {
	if aerr, ok := err.(awserr.Error); ok {
		code := aerr.Code()
		for _, c := range skipCodes {
			if c == code {
				return
			}
		}
		switch code {
		case apigateway.ErrCodeUnauthorizedException:
			a.log.Error(apigateway.ErrCodeUnauthorizedException, aerr.Error())
		case apigateway.ErrCodeLimitExceededException:
			a.log.Error(apigateway.ErrCodeLimitExceededException, aerr.Error())
		case apigateway.ErrCodeConflictException:
			a.log.Error(apigateway.ErrCodeConflictException, aerr.Error())
		case apigateway.ErrCodeBadRequestException:
			a.log.Error(apigateway.ErrCodeBadRequestException, aerr.Error())
		case apigateway.ErrCodeNotFoundException:
			a.log.Error(apigateway.ErrCodeNotFoundException, aerr.Error())
		case apigateway.ErrCodeTooManyRequestsException:
			a.log.Error(apigateway.ErrCodeTooManyRequestsException, aerr.Error())
		case apigateway.ErrCodeServiceUnavailableException:
			a.log.Error(apigateway.ErrCodeServiceUnavailableException, aerr.Error())
		default:
			a.log.Error(aerr.Error())
		}
	} else {
		a.log.Error(err.Error())
	}
}

func (a *APIGatewayRequest) CreateRestAPI(name string) (string, error) {
	a.log.Printf("Creating REST API %s...\n", name)
	input := &apigateway.CreateRestApiInput{
		Name: aws.String(name),
		Description: aws.String(
			fmt.Sprintf("Managed by ginger, created at %s", time.Now().Format("2006-01-02: 15:04:05")),
		),
	}
	result, err := a.svc.CreateRestApi(input)
	if err != nil {
		a.errorLog(err)
		return "", err
	}
	a.log.Infof("REST API created successfully. Id is %s\n", *result.Id)
	return *result.Id, nil
}

func (a *APIGatewayRequest) ResourceExists(restId, resourceId string) bool {
	input := &apigateway.GetResourcesInput{
		RestApiId: aws.String(restId),
	}
	result, err := a.svc.GetResources(input)
	if err != nil {
		a.errorLog(err, apigateway.ErrCodeNotFoundException)
		return false
	}
	for _, item := range result.Items {
		if *item.Id == resourceId {
			return true
		}
	}
	return false
}

func (a *APIGatewayRequest) GetResourceIdByPath(restId, path string) (string, error) {
	a.log.Printf("Getting resource Id by path %s...\n", path)
	input := &apigateway.GetResourcesInput{
		RestApiId: aws.String(restId),
	}
	result, err := a.svc.GetResources(input)
	if err != nil {
		a.errorLog(err)
		return "", err
	}
	for _, item := range result.Items {
		if formatProxyPath(*item.Path) == path {
			return *item.Id, nil
		}
	}
	return "", fmt.Errorf("%s path not found in resources", path)
}

func (a *APIGatewayRequest) CreateResource(restId, parentId, pathPart string) (string, error) {
	a.log.Printf("Creating resource for path part \"%s\"...\n", pathPart)
	input := &apigateway.CreateResourceInput{
		ParentId:  aws.String(parentId),
		PathPart:  aws.String(pathPart),
		RestApiId: aws.String(restId),
	}
	result, err := a.svc.CreateResource(input)
	if err != nil {
		a.errorLog(err)
		return "", err
	}
	a.log.Infof("Resource created successfully. Id is %s\n", *result.Id)
	return *result.Id, nil
}

func (a *APIGatewayRequest) CreateResourceRecursive(restId, path string) (err error) {
	var (
		parentId, parts string
		rs              *entity.Resource
	)
	for _, part := range strings.Split(path, "/") {
		parts += "/" + part
		if rs = a.config.API.Find(parts); rs == nil {
			rs = entity.NewResource("", parts)
			a.config.API.Resources = append(a.config.API.Resources, rs)
		}
		if rs.Id == "" {
			if rs.Id, err = a.CreateResource(restId, parentId, part); err != nil {
				return err
			}
		}
		parentId = rs.Id
	}
	return nil
}

func (a *APIGatewayRequest) PutIntegration(restId string, r *entity.Resource) (err error) {
	fn := a.config.Functions.Find(r.Integration.LambdaFunction)
	if fn == nil {
		err = fmt.Errorf("Function %s couldn't find in your project.\n", r.Integration.LambdaFunction)
		a.errorLog(err)
		return err
	}
	// Need to add extra "/{proxy+}" resource for lambda integration
	if r.Integration.Id == "" {
		r.Integration.Id, err = a.CreateResource(restId, r.Id, "{proxy+}")
		if err != nil {
			return err
		}
	}
	a.PutMethod(restId, r.Integration.Id, "ANY")
	return a.putLambdaIntegration(restId, r.Integration.Id, "ANY", r.Path+"/{proxy+}", fn)
}

func (a *APIGatewayRequest) PutMethod(restId, resourceId, httpMethod string) error {
	a.log.Printf("Putting %s method for resource %s...\n", httpMethod, resourceId)
	input := &apigateway.PutMethodInput{
		// TODO: avoid hard code
		ApiKeyRequired:    aws.Bool(false),
		AuthorizationType: aws.String("NONE"),
		HttpMethod:        aws.String(httpMethod),
		ResourceId:        aws.String(resourceId),
		RestApiId:         aws.String(restId),
	}
	fmt.Println(input)
	if _, err := a.svc.PutMethod(input); err != nil {
		a.errorLog(err, apigateway.ErrCodeConflictException)
		return err
	}
	a.log.Info("Put method successfully.")
	return nil
}

func (a *APIGatewayRequest) generateIntegrationUri(lambdaArn *string) string {
	return fmt.Sprintf(
		"arn:aws:apigateway:%s:lambda:path/2015-03-31/functions/%s/invocations",
		a.config.Project.Region,
		*lambdaArn,
	)
}

func (a *APIGatewayRequest) generateSourceArn(account, restId, httpMethod, path string) string {
	return fmt.Sprintf(
		"arn:aws:execute-api:%s:%s:%s/*/%s%s/{proxy+}",
		a.config.Project.Region,
		account,
		restId,
		httpMethod,
		formatProxyPath(path),
	)
}

func (a *APIGatewayRequest) putLambdaIntegration(restId, resourceId, httpMethod, path string, fn *entity.Function) error {
	a.log.Printf("Putting integration for lambda %s...\n", fn.Name)

	l := NewLambda(a.config)
	fnConfig, err := l.GetFunction(fn.Name)
	if err != nil {
		return err
	}
	input := &apigateway.PutIntegrationInput{
		HttpMethod:            aws.String(httpMethod),
		Type:                  aws.String("AWS_PROXY"),
		Uri:                   aws.String(a.generateIntegrationUri(fnConfig.FunctionArn)),
		ResourceId:            aws.String(resourceId),
		RestApiId:             aws.String(restId),
		IntegrationHttpMethod: aws.String("POST"),
	}
	fmt.Println(input)
	if r, err := a.svc.PutIntegration(input); err != nil {
		a.errorLog(err)
		return err
	} else {
		fmt.Println(r)
	}
	// Add permision to lambda
	account, err := NewSts(a.config).GetAccount()
	if err != nil {
		return err
	}
	sourceArn := a.generateSourceArn(account, restId, httpMethod, path)
	if err := l.AddAPIGatewayPermission(fn.Name, sourceArn); err == nil {
		a.log.Info("Put integration successfully")
	}
	return nil
}

func (a *APIGatewayRequest) Deploy(restId, stage string) error {
	a.log.Printf("Deploy rest APIs for stage: %s...\n", stage)
	input := &apigateway.CreateDeploymentInput{
		StageName:        aws.String(stage),
		StageDescription: aws.String("This stage is managed by ginger"),
		RestApiId:        aws.String(restId),
	}
	if _, err := a.svc.CreateDeployment(input); err != nil {
		a.errorLog(err)
		return err
	}
	a.log.Infof("Stage %s deployed successfully.\n", stage)
	return nil
}

func (a *APIGatewayRequest) DeleteRestApi(restId string) error {
	a.log.Print("Deleting REST API...")
	input := &apigateway.DeleteRestApiInput{
		RestApiId: aws.String(restId),
	}
	if _, err := a.svc.DeleteRestApi(input); err != nil {
		a.errorLog(err)
		return err
	}
	a.log.Info("REST API deleted successfully.")
	return nil
}

func (a *APIGatewayRequest) DeleteResource(restId, resourceId string) error {
	a.log.Print("Deleting resource...")
	input := &apigateway.DeleteResourceInput{
		RestApiId:  aws.String(restId),
		ResourceId: aws.String(resourceId),
	}
	if _, err := a.svc.DeleteResource(input); err != nil {
		a.errorLog(err)
		return err
	}
	a.log.Info("Resource deleted successfully.")
	return nil
}

func (a *APIGatewayRequest) DeleteMethod(restId, resourceId string) error {
	a.log.Print("Deleting method...")
	input := &apigateway.DeleteMethodInput{
		HttpMethod: aws.String("ANY"),
		RestApiId:  aws.String(restId),
		ResourceId: aws.String(resourceId),
	}
	if _, err := a.svc.DeleteMethod(input); err != nil {
		a.errorLog(err)
		return err
	}
	a.log.Info("Method deleted successfully.")
	return nil
}

func (a *APIGatewayRequest) DeleteIntegration(restId, resourceId string) error {
	a.log.Print("Deleting integration...")
	input := &apigateway.DeleteIntegrationInput{
		HttpMethod: aws.String("ANY"),
		RestApiId:  aws.String(restId),
		ResourceId: aws.String(resourceId),
	}
	if _, err := a.svc.DeleteIntegration(input); err != nil {
		a.errorLog(err)
		return err
	}
	a.log.Info("Integration deleted successfully.")
	return nil
}
