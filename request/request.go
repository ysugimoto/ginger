package request

import (
	"fmt"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/ysugimoto/ginger/config"
)

var debug string = ""

// generateStatementId() generated unique statement string.
func generateStatementId(sType string) string {
	return fmt.Sprintf("ginger-statement-%s-%d", sType, time.Now().UnixNano())
}

// Create common AWS session.
func createAWSSession(c *config.Config) *session.Session {
	conf := aws.NewConfig().WithRegion(c.Project.Region)
	if c.Project.Profile != "" {
		conf = conf.WithCredentials(
			credentials.NewSharedCredentials("", c.Project.Profile),
		)
	}
	return session.New(conf)
}

// debug print if enables.
func debugRequest(obj fmt.Stringer) {
	if debug != "enable" {
		return
	}
	var name string
	if t := reflect.TypeOf(obj); t.Kind() == reflect.Ptr {
		name = "*" + t.Elem().Name()
	} else {
		name = t.Name()
	}
	fmt.Printf("[DEBUG] %s\n", name)
	fmt.Println(obj)
}
