package request

import (
	"bytes"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/ysugimoto/ginger/internal/config"
	"github.com/ysugimoto/ginger/internal/entity"
	"github.com/ysugimoto/ginger/internal/logger"
)

// S3Request is the struct which manages AWS s3 service.
type S3Request struct {
	svc    *s3.S3
	log    *logger.Logger
	config *config.Config
}

func NewS3(c *config.Config) *S3Request {
	return &S3Request{
		config: c,
		svc:    s3.New(createAWSSession(c)),
		log:    logger.WithNamespace("ginger.request.s3"),
	}
}

func (s *S3Request) errorLog(err error) {
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case s3.ErrCodeBucketAlreadyExists:
			s.log.Error(s3.ErrCodeBucketAlreadyExists, aerr.Error())
		case s3.ErrCodeBucketAlreadyOwnedByYou:
			s.log.Error(s3.ErrCodeBucketAlreadyOwnedByYou, aerr.Error())
		case s3.ErrCodeNoSuchBucket:
			s.log.Error(s3.ErrCodeNoSuchBucket, aerr.Error())
		case s3.ErrCodeNoSuchKey:
			s.log.Error(s3.ErrCodeNoSuchKey, aerr.Error())
		case s3.ErrCodeNoSuchUpload:
			s.log.Error(s3.ErrCodeNoSuchUpload, aerr.Error())
		case s3.ErrCodeObjectAlreadyInActiveTierError:
			s.log.Error(s3.ErrCodeObjectAlreadyInActiveTierError, aerr.Error())
		case s3.ErrCodeObjectNotInActiveTierError:
			s.log.Error(s3.ErrCodeObjectNotInActiveTierError, aerr.Error())
		default:
			s.log.Error(aerr.Error())
		}
	} else {
		s.log.Error(err.Error())
	}
}

func (s *S3Request) EnsureBucketExists(bucket string) error {
	input := &s3.CreateBucketInput{
		ACL:    aws.String("public-read"),
		Bucket: aws.String(bucket),
	}
	debugRequest(input)
	result, err := s.svc.CreateBucket(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeBucketAlreadyOwnedByYou:
				// ErrCodeBucketAlreadyOwnedByYou means bucket has already been created by you,
				// So this error code should deal with success.
				s.log.Infof("Bucket %s has already been created.\n", bucket)
				return nil
			}
			return aerr
		}
		return err
	}
	debugRequest(result)
	s.log.Infof("Bucket %s set up successfully.\n", bucket)
	return nil
}

func (s *S3Request) PutObject(bucket string, so *entity.StorageObject) error {
	input := &s3.PutObjectInput{
		Body:          aws.ReadSeekCloser(bytes.NewReader(so.Data)),
		Bucket:        aws.String(bucket),
		Key:           aws.String(so.Key),
		ContentLength: aws.Int64(so.Info.Size()),
		ACL:           aws.String("public-read"),
	}
	debugRequest(input)
	result, err := s.svc.PutObject(input)
	if err != nil {
		s.errorLog(err)
		return err
	}
	debugRequest(result)
	return nil
}
