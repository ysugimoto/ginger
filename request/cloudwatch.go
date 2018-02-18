package request

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"

	"github.com/ysugimoto/ginger/config"
	"github.com/ysugimoto/ginger/logger"
)

// CloudWatchLogsRequest is the struct which wrap AWS cloud watch logs service.
type CloudWatchLogsRequest struct {
	svc    *cloudwatchlogs.CloudWatchLogs
	log    *logger.Logger
	config *config.Config
}

func NewCloudWatchLogsRequest(c *config.Config) *CloudWatchLogsRequest {
	return &CloudWatchLogsRequest{
		config: c,
		svc:    cloudwatchlogs.New(createAWSSession(c)),
		log:    logger.WithNamespace("ginger.request.logs"),
	}
}

func (c *CloudWatchLogsRequest) errorLog(err error) {
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case cloudwatchlogs.ErrCodeInvalidParameterException:
			c.log.Error(cloudwatchlogs.ErrCodeInvalidParameterException, aerr.Error())
		case cloudwatchlogs.ErrCodeResourceNotFoundException:
			c.log.Error(cloudwatchlogs.ErrCodeResourceNotFoundException, aerr.Error())
		case cloudwatchlogs.ErrCodeServiceUnavailableException:
			c.log.Error(cloudwatchlogs.ErrCodeServiceUnavailableException, aerr.Error())
		default:
			c.log.Error(aerr.Error())
		}
	} else {
		c.log.Error(err.Error())
	}
}

func (c *CloudWatchLogsRequest) TailLogs(ctx context.Context, groupName, filter string) {
	// First request immediately
	updatedTime, err := c.requestLog(groupName, filter, time.Now().UnixNano()/int64(time.Millisecond))
	if err != nil {
		c.errorLog(err)
		return
	}
	// Send request will call on tick by each 500 ms
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if updatedTime, err = c.requestLog(groupName, filter, updatedTime); err != nil {
				c.errorLog(err)
				return
			}
		}
	}
}

func (c *CloudWatchLogsRequest) requestLog(groupName, filter string, startTime int64) (int64, error) {
	input := &cloudwatchlogs.FilterLogEventsInput{
		Limit:        aws.Int64(100),
		LogGroupName: aws.String(groupName),
		StartTime:    aws.Int64(startTime),
	}
	if filter != "" {
		input.FilterPattern = aws.String(filter)
	}
	result, err := c.svc.FilterLogEvents(input)
	if err != nil {
		c.errorLog(err)
		return 0, err
	}
	events := result.Events
	for _, e := range events {
		ts := *e.Timestamp
		t := time.Unix(ts/1000, (ts%1000)*1000000)
		fmt.Printf("[%s] %s", t.Format(time.RFC3339), *e.Message)
	}
	if len(events) > 0 {
		startTime = *events[len(events)-1].Timestamp
		startTime++
	}
	return startTime, nil
}
