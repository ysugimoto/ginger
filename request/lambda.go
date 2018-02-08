package request

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/ysugimoto/ginger/config"
	"github.com/ysugimoto/ginger/entity"
	"github.com/ysugimoto/ginger/logger"
)

type LambdaRequest struct {
	svc    *lambda.Lambda
	log    *logger.Logger
	config *config.Config
}

func NewLambda(c *config.Config) *LambdaRequest {
	return &LambdaRequest{
		config: c,
		svc:    lambda.New(createAWSSession(c)),
		log:    logger.WithNamespace("ginger.request.lambda"),
	}
}

func (l *LambdaRequest) errorLog(err error) {
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case lambda.ErrCodeServiceException:
			l.log.Error(lambda.ErrCodeServiceException, aerr.Error())
		case lambda.ErrCodeResourceNotFoundException:
			l.log.Error(lambda.ErrCodeResourceNotFoundException, aerr.Error())
		case lambda.ErrCodeInvalidRequestContentException:
			l.log.Error(lambda.ErrCodeInvalidRequestContentException, aerr.Error())
		case lambda.ErrCodeRequestTooLargeException:
			l.log.Error(lambda.ErrCodeRequestTooLargeException, aerr.Error())
		case lambda.ErrCodeUnsupportedMediaTypeException:
			l.log.Error(lambda.ErrCodeUnsupportedMediaTypeException, aerr.Error())
		case lambda.ErrCodeTooManyRequestsException:
			l.log.Error(lambda.ErrCodeTooManyRequestsException, aerr.Error())
		case lambda.ErrCodeInvalidParameterValueException:
			l.log.Error(lambda.ErrCodeInvalidParameterValueException, aerr.Error())
		case lambda.ErrCodeEC2UnexpectedException:
			l.log.Error(lambda.ErrCodeEC2UnexpectedException, aerr.Error())
		case lambda.ErrCodeSubnetIPAddressLimitReachedException:
			l.log.Error(lambda.ErrCodeSubnetIPAddressLimitReachedException, aerr.Error())
		case lambda.ErrCodeENILimitReachedException:
			l.log.Error(lambda.ErrCodeENILimitReachedException, aerr.Error())
		case lambda.ErrCodeEC2ThrottledException:
			l.log.Error(lambda.ErrCodeEC2ThrottledException, aerr.Error())
		case lambda.ErrCodeEC2AccessDeniedException:
			l.log.Error(lambda.ErrCodeEC2AccessDeniedException, aerr.Error())
		case lambda.ErrCodeInvalidSubnetIDException:
			l.log.Error(lambda.ErrCodeInvalidSubnetIDException, aerr.Error())
		case lambda.ErrCodeInvalidSecurityGroupIDException:
			l.log.Error(lambda.ErrCodeInvalidSecurityGroupIDException, aerr.Error())
		case lambda.ErrCodeInvalidZipFileException:
			l.log.Error(lambda.ErrCodeInvalidZipFileException, aerr.Error())
		case lambda.ErrCodeKMSDisabledException:
			l.log.Error(lambda.ErrCodeKMSDisabledException, aerr.Error())
		case lambda.ErrCodeKMSInvalidStateException:
			l.log.Error(lambda.ErrCodeKMSInvalidStateException, aerr.Error())
		case lambda.ErrCodeKMSAccessDeniedException:
			l.log.Error(lambda.ErrCodeKMSAccessDeniedException, aerr.Error())
		case lambda.ErrCodeKMSNotFoundException:
			l.log.Error(lambda.ErrCodeKMSNotFoundException, aerr.Error())
		case lambda.ErrCodeInvalidRuntimeException:
			l.log.Error(lambda.ErrCodeInvalidRuntimeException, aerr.Error())
		default:
			l.log.Error(aerr.Error())
		}
	} else {
		l.log.Error(err.Error())
	}
}

func (l *LambdaRequest) FunctionExists(name string) bool {
	if _, err := l.GetFunction(name); err != nil {
		return false
	}
	return true
}

func (l *LambdaRequest) DeleteFunction(name string) error {
	input := &lambda.DeleteFunctionInput{
		FunctionName: aws.String(name),
	}
	if _, err := l.svc.DeleteFunction(input); err != nil {
		l.errorLog(err)
		return err
	}
	l.log.Infof("Function %s deleted from AWS", name)
	return nil
}

func (l *LambdaRequest) DeployFunction(fn *entity.Function, zipBytes []byte) (string, error) {
	if l.FunctionExists(fn.Name) {
		l.log.Printf("%s already exists, update fucntion\n", fn.Name)
		return l.UpdateFunction(fn, zipBytes)
	} else {
		l.log.Printf("Creating new function %s\n", name)
		return l.CreateFunction(fn, zipBytes)
	}
}

func (l *LambdaRequest) CreateFunction(fn *entity.Function, zipBytes []byte) (string, error) {
	input := &lambda.CreateFunctionInput{
		Code: &lambda.FunctionCode{
			ZipFile: zipBytes,
		},
		FunctionName: aws.String(fn.Name),
		Handler:      aws.String(fn.Name),
		Role:         aws.String(l.config.Project.LambdaExecutionRole),
		MemorySize:   aws.Int64(fn.MemorySize),
		Publish:      aws.Bool(true),
		Runtime:      aws.String("go1.x"),
		Timeout:      aws.Int64(fn.Timeout),
	}
	result, err := l.svc.CreateFunction(input)
	if err != nil {
		l.errorLog(err)
		return "", err
	}
	return *result.FunctionArn, nil
}

func (l *LambdaRequest) UpdateFunction(fn *entity.Function, zipBytes []byte) (string, error) {
	if err := l.UpdateFunctionConfiguration(fn); err != nil {
		return "", err
	}
	input := &lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String(name),
		Publish:      aws.Bool(true),
		ZipFile:      zipBytes,
	}

	result, err := l.svc.UpdateFunctionCode(input)
	if err != nil {
		l.errorLog(err)
		return "", err
	}
	return *result.FunctionArn, nil
}

func (l *LambdaRequest) AddS3Permission(name, bucketName string) error {
	sts := NewSts(l.config)
	account, err := sts.GetAccount()
	if err != nil {
		return err
	}
	input := &lambda.AddPermissionInput{
		Action:        aws.String("lambda.InvokeFunction"),
		Principal:     aws.String("s3.amazonaws.com"),
		SourceArn:     aws.String(fmt.Sprintf("arn:aws:s3:::%s", bucketName)),
		SourceAccount: aws.String(account),
		FunctionName:  aws.String(name),
		StatementId:   aws.String(generateStatementId("s3")),
	}
	if _, err := l.svc.AddPermission(input); err != nil {
		l.errorLog(err)
		return err
	}
	return nil
}

func (l *LambdaRequest) AddAPIGatewayPermission(name, apiArn string) error {
	sts := NewSts(l.config)
	account, err := sts.GetAccount()
	if err != nil {
		return err
	}
	input := &lambda.AddPermissionInput{
		Action:        aws.String("lambda.InvokeFunction"),
		Principal:     aws.String("apigateway.amazonaws.com"),
		SourceArn:     aws.String(apiArn),
		SourceAccount: aws.String(account),
		FunctionName:  aws.String(name),
		StatementId:   aws.String(generateStatementId("apigateway")),
	}
	if _, err := l.svc.AddPermission(input); err != nil {
		l.errorLog(err)
		return err
	}
	return nil
}

func (l *LambdaRequest) GetFunction(name string) (*lambda.FunctionConfiguration, error) {
	l.log.Printf("Getting lambda function for %s...\n", name)
	input := &lambda.GetFunctionInput{
		FunctionName: aws.String(name),
	}
	result, err := l.svc.GetFunction(input)
	if err != nil {
		l.errorLog(err)
		return nil, err
	}
	return result.Configuration, nil
}

func (l *LambdaRequest) UpdateFunctionConfiguration(fn *entity.Function) error {
	l.log.Printf("Updating function configuration for %s...\n", fn.Name)
	input := &lambda.UpdateFunctionConfigurationInput{
		FunctionName: aws.String(fn.Name),
		MemorySize:   aws.Int64(fn.MemorySize),
		TImeout:      aws.Int64(fn.Timeout),
	}
	if _, err := lambda.UpdateFunctionConfiguration(input); err != nil {
		l.errorLog(err)
		return err
	}
	l.log.Infof("Function configuration has been updated.")
	return nil
}

func (l *LambdaRequest) InvokeFunction(name string, payload []byte) error {
	input := &lambda.InvokeFunctionInput{
		FunctionName: name,
		Payload:      payload,
	}
	result, err := l.svc.Invoke(input)
	if err != nil {
		l.errorLog(err)
		return err
	}
	if result.FunctionError != nil {
		l.log.Warnf("Function invoked on version: %s and handed error\n", *result.ExecutedVersion)
	} else {
		l.log.Infof("Function invoked on version: %s and succeeded\n", *result.ExecutedVersion)
	}
	l.log.Print(string(Payload))
	return nil
}
