package request

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/sts"

	"github.com/ysugimoto/ginger/internal/config"
	"github.com/ysugimoto/ginger/internal/logger"
)

type StsRequest struct {
	svc    *sts.STS
	log    *logger.Logger
	config *config.Config
}

func NewSts(c *config.Config) *StsRequest {
	return &StsRequest{
		config: c,
		svc:    sts.New(createAWSSession(c)),
		log:    logger.WithNamespace("ginger.request.sts"),
	}
}

func (s *StsRequest) GetAccount() (string, error) {
	input := &sts.GetCallerIdentityInput{}
	result, err := s.svc.GetCallerIdentity(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			s.log.Error(aerr.Error())
		} else {
			s.log.Error(err)
		}
		return "", err
	}
	return *result.Account, nil
}
