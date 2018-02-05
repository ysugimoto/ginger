package request

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
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
	conf := aws.NewConfig().WithRegion(c.Project.Region)
	if c.Project.Profile != "" {
		conf = conf.WithCredentials(
			credentials.NewSharedCredentials("", c.Project.Profile),
		)
	}
	return &LambdaRequest{
		config: c,
		svc:    lambda.New(session.New(), conf),
		log:    logger.WithNamespace("ginger.request.lambda"),
	}
}

func (l *LambdaRequest) FunctionExists(name string) bool {
	_, err := l.svc.GetFunction(&lambda.GetFunctionInput{
		FunctionName: aws.String(name),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case lambda.ErrCodeServiceException:
				l.log.Error(lambda.ErrCodeServiceException, aerr.Error())
			case lambda.ErrCodeResourceNotFoundException:
				return false
			case lambda.ErrCodeTooManyRequestsException:
				l.log.Error(lambda.ErrCodeTooManyRequestsException, aerr.Error())
			case lambda.ErrCodeInvalidParameterValueException:
				l.log.Error(lambda.ErrCodeInvalidParameterValueException, aerr.Error())
			default:
				l.log.Error(aerr.Error())
			}
		} else {
			l.log.Error(err.Error())
		}
		return false
	}
	return true
}

func (l *LambdaRequest) DeleteFunction(name string) error {
	_, err := l.svc.DeleteFunction(&lambda.DeleteFunctionInput{
		FunctionName: aws.String(name),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case lambda.ErrCodeServiceException:
				l.log.Error(lambda.ErrCodeServiceException, aerr.Error())
			case lambda.ErrCodeResourceNotFoundException:
				l.log.Error(lambda.ErrCodeResourceNotFoundException, aerr.Error())
			case lambda.ErrCodeTooManyRequestsException:
				l.log.Error(lambda.ErrCodeTooManyRequestsException, aerr.Error())
			case lambda.ErrCodeInvalidParameterValueException:
				l.log.Error(lambda.ErrCodeInvalidParameterValueException, aerr.Error())
			case lambda.ErrCodeResourceConflictException:
				l.log.Error(lambda.ErrCodeResourceConflictException, aerr.Error())
			default:
				l.log.Error(aerr.Error())
			}
			return aerr
		} else {
			l.log.Error(err.Error())
			return err
		}
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
	_, err := l.svc.CreateFunction(&lambda.CreateFunctionInput{
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
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case lambda.ErrCodeServiceException:
				l.log.Error(lambda.ErrCodeServiceException, aerr.Error())
			case lambda.ErrCodeInvalidParameterValueException:
				l.log.Error(lambda.ErrCodeInvalidParameterValueException, aerr.Error())
			case lambda.ErrCodeResourceNotFoundException:
				l.log.Error(lambda.ErrCodeResourceNotFoundException, aerr.Error())
			case lambda.ErrCodeResourceConflictException:
				l.log.Error(lambda.ErrCodeResourceConflictException, aerr.Error())
			case lambda.ErrCodeTooManyRequestsException:
				l.log.Error(lambda.ErrCodeTooManyRequestsException, aerr.Error())
			case lambda.ErrCodeCodeStorageExceededException:
				l.log.Error(lambda.ErrCodeCodeStorageExceededException, aerr.Error())
			default:
				l.log.Error(aerr.Error())
			}
		} else {
			l.log.Error(err.Error())
		}
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

	_, err := l.svc.UpdateFunctionCode(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case lambda.ErrCodeServiceException:
				l.log.Error(lambda.ErrCodeServiceException, aerr.Error())
			case lambda.ErrCodeResourceNotFoundException:
				l.log.Error(lambda.ErrCodeResourceNotFoundException, aerr.Error())
			case lambda.ErrCodeInvalidParameterValueException:
				l.log.Error(lambda.ErrCodeInvalidParameterValueException, aerr.Error())
			case lambda.ErrCodeTooManyRequestsException:
				l.log.Error(lambda.ErrCodeTooManyRequestsException, aerr.Error())
			case lambda.ErrCodeCodeStorageExceededException:
				l.log.Error(lambda.ErrCodeCodeStorageExceededException, aerr.Error())
			case lambda.ErrCodePreconditionFailedException:
				l.log.Error(lambda.ErrCodePreconditionFailedException, aerr.Error())
			default:
				l.log.Error(aerr.Error())
			}
		} else {
			l.log.Error(err.Error())
		}
		return err
	}
	return nil
}
