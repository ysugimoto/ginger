package request

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"

	"github.com/ysugimoto/ginger/config"
	"github.com/ysugimoto/ginger/entity"
	"github.com/ysugimoto/ginger/logger"
)

// CloudWatchRequest is the struct which wrap AWS cloud watch logs service.
type CloudWatchRequest struct {
	svc    *cloudwatchlogs.CloudWatchLogs
	events *cloudwatchevents.CloudWatchEvents
	log    *logger.Logger
	config *config.Config
}

func NewCloudWatch(c *config.Config) *CloudWatchRequest {
	sess := createAWSSession(c)
	return &CloudWatchRequest{
		config: c,
		svc:    cloudwatchlogs.New(sess),
		events: cloudwatchevents.New(sess),
		log:    logger.WithNamespace("ginger.request.cloudwatch"),
	}
}

func (c *CloudWatchRequest) errorLog(err error) {
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {

		// cloudwatchlogs error codes
		case cloudwatchlogs.ErrCodeInvalidParameterException:
			c.log.Error(cloudwatchlogs.ErrCodeInvalidParameterException, aerr.Error())
		case cloudwatchlogs.ErrCodeResourceNotFoundException:
			c.log.Error(cloudwatchlogs.ErrCodeResourceNotFoundException, aerr.Error())
		case cloudwatchlogs.ErrCodeServiceUnavailableException:
			c.log.Error(cloudwatchlogs.ErrCodeServiceUnavailableException, aerr.Error())

		// cloudwatchevents error codes
		case cloudwatchevents.ErrCodeInvalidEventPatternException:
			c.log.Error(cloudwatchlogs.ErrCodeServiceUnavailableException, aerr.Error())
		case cloudwatchevents.ErrCodeLimitExceededException:
			c.log.Error(cloudwatchevents.ErrCodeLimitExceededException, aerr.Error())
		case cloudwatchevents.ErrCodeConcurrentModificationException:
			c.log.Error(cloudwatchevents.ErrCodeConcurrentModificationException, aerr.Error())
		case cloudwatchevents.ErrCodeInternalException:
			c.log.Error(cloudwatchevents.ErrCodeInternalException, aerr.Error())
		default:
			c.log.Error(aerr.Error())
		}
	} else {
		c.log.Error(err.Error())
	}
}

func (c *CloudWatchRequest) TailLogs(ctx context.Context, groupName, filter string) {
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

func (c *CloudWatchRequest) requestLog(groupName, filter string, startTime int64) (int64, error) {
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

func (c *CloudWatchRequest) CreateOrUpdateSchedule(sc *entity.Scheduler) (string, error) {
	c.log.Printf("Create schedule for cloudwatch, name: %s, cron: %s...\n", sc.Name, sc.Expression)
	state := "DISABLED"
	if sc.Enable {
		state = "ENABLED"
	}
	input := &cloudwatchevents.PutRuleInput{
		Description:        aws.String(fmt.Sprintf("Created by ginger for %s", c.config.ProjectName)),
		Name:               aws.String(sc.Name),
		ScheduleExpression: aws.String(sc.Expression),
		State:              aws.String(state),
	}
	debugRequest(input)
	result, err := c.events.PutRule(input)
	if err != nil {
		c.errorLog(err)
		return "", err
	}
	debugRequest(result)
	c.log.Info("Schedule event created successfully")
	return *result.RuleArn, nil
}

func (c *CloudWatchRequest) GetScheduleArn(name string) (string, error) {
	input := &cloudwatchevents.ListRulesInput{
		Limit:      aws.Int64(100),
		NamePrefix: aws.String(name),
	}
	debugRequest(input)
	result, err := c.events.ListRules(input)
	if err != nil {
		c.errorLog(err)
		return "", err
	}
	debugRequest(result)
	for _, r := range result.Rules {
		if *r.Name == name {
			return *r.Arn, nil
		}
	}
	return "", nil
}

func (c *CloudWatchRequest) DeleteSchedule(name string) error {
	c.log.Printf("Delete schedule from cloudwatch, name %s...\n", name)
	input := &cloudwatchevents.DeleteRuleInput{
		Name: aws.String(name),
	}
	debugRequest(input)
	result, err := c.events.DeleteRule(input)
	if err != nil {
		c.errorLog(err)
		return err
	}
	debugRequest(result)
	c.log.Info("Schedule event deleted successfully")
	return nil
}

func (c *CloudWatchRequest) PutTarget(scheduleName string, functionArn *string) error {
	c.log.Printf("Putting schedule target of lambda function to %s...\n", scheduleName)
	input := &cloudwatchevents.PutTargetsInput{
		Rule: aws.String(scheduleName),
		Targets: []*cloudwatchevents.Target{
			&cloudwatchevents.Target{
				Arn: functionArn,
				Id:  aws.String("1"),
			},
		},
	}
	debugRequest(input)
	result, err := c.events.PutTargets(input)
	if err != nil {
		c.errorLog(err)
		return err
	}
	debugRequest(result)
	c.log.Info("Put schedule target successfully")
	return nil
}
