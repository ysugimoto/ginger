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
	debugRequest(input)
	result, err := a.svc.CreateRestApi(input)
	if err != nil {
		a.errorLog(err)
		return "", err
	}
	debugRequest(result)
	a.log.Infof("REST API created successfully. Id is %s\n", *result.Id)
	return *result.Id, nil
}

func (a *APIGatewayRequest) ResourceExists(restId, resourceId string) bool {
	input := &apigateway.GetResourcesInput{
		RestApiId: aws.String(restId),
	}
	debugRequest(input)
	result, err := a.svc.GetResources(input)
	if err != nil {
		a.errorLog(err, apigateway.ErrCodeNotFoundException)
		return false
	}
	debugRequest(result)
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
	debugRequest(input)
	result, err := a.svc.GetResources(input)
	if err != nil {
		a.errorLog(err)
		return "", err
	}
	debugRequest(result)
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
	debugRequest(input)
	result, err := a.svc.CreateResource(input)
	if err != nil {
		a.errorLog(err)
		return "", err
	}
	debugRequest(result)
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
		if rs, err = a.config.LoadResource(parts); err != nil {
			rs = entity.NewResource("", parts)
			a.config.Resources = append(a.config.Resources, rs)
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

func (a *APIGatewayRequest) PutIntegration(restId, resourceId, method string, i *entity.Integration) (err error) {
	switch i.IntegrationType {
	case "lambda":
		fn, err := a.config.LoadFunction(*i.LambdaFunction)
		if err != nil {
			err = fmt.Errorf("Function %s couldn't find in your project.\n", *i.LambdaFunction)
			a.errorLog(err)
			return err
		}
		a.PutMethod(restId, resourceId, method)
		return a.putLambdaIntegration(restId, resourceId, method, i.Path+"/{proxy+}", fn)
	case "s3":
		a.PutMethod(restId, resourceId, method)
		return a.putS3Integration(restId, resourceId, method, *i.Bucket)
	default:
		a.log.Errorf("Unexpected integration type %s\n", i.IntegrationType)
		return nil
	}
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
	debugRequest(input)
	result, err := a.svc.PutMethod(input)
	if err != nil {
		a.errorLog(err, apigateway.ErrCodeConflictException)
		return err
	}
	debugRequest(result)
	a.log.Info("Put method successfully.")
	return nil
}

func (a *APIGatewayRequest) generateIntegrationUri(lambdaArn *string) string {
	return fmt.Sprintf(
		"arn:aws:apigateway:%s:lambda:path/2015-03-31/functions/%s/invocations",
		a.config.Region,
		*lambdaArn,
	)
}

func (a *APIGatewayRequest) generateSourceArn(account, restId, httpMethod, path string) string {
	return fmt.Sprintf(
		"arn:aws:execute-api:%s:%s:%s/*/%s%s/{proxy+}",
		a.config.Region,
		account,
		restId,
		httpMethod,
		formatProxyPath(path),
	)
}

func (a *APIGatewayRequest) putS3Integration(restId, resourceId, httpMethod, bucketName string) error {
	a.log.Print("Putting S3 integration...")

	input := &apigateway.PutIntegrationInput{
		HttpMethod:            aws.String(httpMethod),
		Type:                  aws.String("HTTP_PROXY"),
		Uri:                   aws.String(fmt.Sprintf("https://s3.amazonaws.com/%s/{proxy}", bucketName)),
		ResourceId:            aws.String(resourceId),
		RestApiId:             aws.String(restId),
		IntegrationHttpMethod: aws.String("GET"),
		ContentHandling:       aws.String("CONVERT_TO_BINAY"),
		RequestParameters: map[string]*string{
			"integration.request.path.proxy": aws.String("method.request.path.proxy"),
		},
		CacheKeyParameters: []*string{
			aws.String("method.request.path.proxy"),
		},
	}
	debugRequest(input)
	result, err := a.svc.PutIntegration(input)
	if err != nil {
		a.errorLog(err)
		return err
	}
	debugRequest(result)
	a.log.Print("Put integration successfully.")
	return nil
}

func (a *APIGatewayRequest) putLambdaIntegration(restId, resourceId, httpMethod, path string, fn *entity.Function) error {
	a.log.Printf("Putting Lambda integration for %s...\n", path)

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
	debugRequest(input)
	result, err := a.svc.PutIntegration(input)
	if err != nil {
		a.errorLog(err)
		return err
	}
	debugRequest(result)
	// Add permision to lambda
	account, err := NewSts(a.config).GetAccount()
	if err != nil {
		return err
	}
	sourceArn := a.generateSourceArn(account, restId, httpMethod, path)
	if err := l.AddAPIGatewayPermission(fn.Name, sourceArn); err == nil {
		a.log.Info("Put integration successfully.")
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
	debugRequest(input)
	result, err := a.svc.CreateDeployment(input)
	if err != nil {
		a.errorLog(err)
		return err
	}
	debugRequest(result)
	a.log.Infof("Stage %s deployed successfully.\n", stage)
	return nil
}

func (a *APIGatewayRequest) DeleteRestApi(restId string) error {
	a.log.Print("Deleting REST API...")
	input := &apigateway.DeleteRestApiInput{
		RestApiId: aws.String(restId),
	}
	debugRequest(input)
	result, err := a.svc.DeleteRestApi(input)
	if err != nil {
		a.errorLog(err)
		return err
	}
	debugRequest(result)
	a.log.Info("REST API deleted successfully.")
	return nil
}

func (a *APIGatewayRequest) DeleteResource(restId, resourceId string) error {
	a.log.Print("Deleting resource...")
	input := &apigateway.DeleteResourceInput{
		RestApiId:  aws.String(restId),
		ResourceId: aws.String(resourceId),
	}
	debugRequest(input)
	result, err := a.svc.DeleteResource(input)
	if err != nil {
		a.errorLog(err)
		return err
	}
	debugRequest(result)
	a.log.Info("Resource deleted successfully.")
	return nil
}

func (a *APIGatewayRequest) DeleteMethod(restId, resourceId, method string) error {
	a.log.Print("Deleting method...")
	input := &apigateway.DeleteMethodInput{
		HttpMethod: aws.String(strings.ToUpper(method)),
		RestApiId:  aws.String(restId),
		ResourceId: aws.String(resourceId),
	}
	debugRequest(input)
	result, err := a.svc.DeleteMethod(input)
	if err != nil {
		a.errorLog(err)
		return err
	}
	debugRequest(result)
	a.log.Info("Method deleted successfully.")
	return nil
}

func (a *APIGatewayRequest) DeleteIntegration(restId, resourceId, method string) error {
	a.log.Print("Deleting integration...")
	input := &apigateway.DeleteIntegrationInput{
		HttpMethod: aws.String(strings.ToUpper(method)),
		RestApiId:  aws.String(restId),
		ResourceId: aws.String(resourceId),
	}
	debugRequest(input)
	result, err := a.svc.DeleteIntegration(input)
	if err != nil {
		a.errorLog(err)
		return err
	}
	debugRequest(result)
	a.log.Info("Integration deleted successfully.")
	return nil
}

func (a *APIGatewayRequest) CreateStage(restId, stageName string) error {
	a.log.Printf("Creating new stage \"%s\"...\n", stageName)
	input := &apigateway.CreateStageInput{
		StageName:   aws.String(stageName),
		Description: aws.String("Created via ginger project"),
		RestApiId:   aws.String(restId),
	}
	debugRequest(input)
	result, err := a.svc.CreateStage(input)
	if err != nil {
		a.errorLog(err)
		return err
	}
	debugRequest(result)
	a.log.Info("Stage created successfully.")
	return nil
}

func (a *APIGatewayRequest) DeleteStage(restId, stageName string) error {
	a.log.Printf("Deleting stage \"%s\"...\n", stageName)
	input := &apigateway.DeleteStageInput{
		StageName: aws.String(stageName),
		RestApiId: aws.String(restId),
	}
	debugRequest(input)
	result, err := a.svc.DeleteStage(input)
	if err != nil {
		a.errorLog(err)
		return err
	}
	debugRequest(result)
	a.log.Info("Stage deleted successfully.")
	return nil
}

func (a *APIGatewayRequest) StageExists(restId, stageName string) bool {
	input := &apigateway.GetStageInput{
		StageName: aws.String(stageName),
		RestApiId: aws.String(restId),
	}
	debugRequest(input)
	result, err := a.svc.GetStage(input)
	if err != nil {
		a.errorLog(err, apigateway.ErrCodeNotFoundException)
		return false
	}
	debugRequest(result)
	return true
}

func (a *APIGatewayRequest) GetStages(restId string) []*apigateway.Stage {
	input := &apigateway.GetStagesInput{
		RestApiId: aws.String(restId),
	}
	debugRequest(input)
	result, err := a.svc.GetStages(input)
	if err != nil {
		a.errorLog(err, apigateway.ErrCodeNotFoundException)
		return make([]*apigateway.Stage, 0)
	}
	debugRequest(result)
	return result.Item
}
