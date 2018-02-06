package request

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/ysugimoto/ginger/config"
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
		case lambda.ErrCodeResourceConflictException:
			l.log.Error(lambda.ErrCodeResourceConflictException, aerr.Error())
		case lambda.ErrCodeInvalidParameterValueException:
			l.log.Error(lambda.ErrCodeInvalidParameterValueException, aerr.Error())
		case lambda.ErrCodePolicyLengthExceededException:
			l.log.Error(lambda.ErrCodePolicyLengthExceededException, aerr.Error())
		case lambda.ErrCodeTooManyRequestsException:
			l.log.Error(lambda.ErrCodeTooManyRequestsException, aerr.Error())
		case lambda.ErrCodePreconditionFailedException:
			l.log.Error(lambda.ErrCodePreconditionFailedException, aerr.Error())
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

func (l *LambdaRequest) DeployFunction(name string, zipBytes []byte) error {
	if l.FunctionExists(name) {
		l.log.Printf("%s already exists, update fucntion\n", name)
		return l.UpdateFunction(name, zipBytes)
	} else {
		l.log.Printf("Creating new function %s\n", name)
		return l.CreateFunction(name, zipBytes)
	}
}

func (l *LambdaRequest) CreateFunction(name string, zipBytes []byte) error {
	input := &lambda.CreateFunctionInput{
		Code: &lambda.FunctionCode{
			ZipFile: zipBytes,
		},
		FunctionName: aws.String(name),
		Handler:      aws.String(name),
		Role:         aws.String(l.config.Project.LambdaExecutionRole),
		MemorySize:   aws.Int64(256),
		Publish:      aws.Bool(true),
		Runtime:      aws.String("go1.x"),
		Timeout:      aws.Int64(10),
	}
	if _, err := l.svc.CreateFunction(input); err != nil {
		l.errorLog(err)
		return err
	}
	return nil
}

func (l *LambdaRequest) UpdateFunction(name string, zipBytes []byte) error {
	input := &lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String(name),
		Publish:      aws.Bool(true),
		ZipFile:      zipBytes,
	}

	if _, err := l.svc.UpdateFunctionCode(input); err != nil {
		l.errorLog(err)
		return err
	}
	return nil
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
