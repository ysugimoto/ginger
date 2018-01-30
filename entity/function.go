package entity

import (
	"errors"
	"fmt"
	"os"

	"io/ioutil"
	"os/exec"
	"path/filepath"

	"github.com/iancoleman/strcase"
)

const MAIN_FUNCTION_TEMPLATE = `package main

import (
	"github.com/aws/aws-lambda-go/lambda"%s
	"github.com/aws/aws-sdk-go"
)

func %sHandler(%s) (%s, error) {
	return %s
}

func main() {
	lambda.Start(%sHandler)
}
`

func bindTemplate(name string, apiGateway *APIGateway) []byte {
	binds := []interface{}{}
	if apiGateway != nil {
		binds = append(binds, `\n\t"github.com/aws/aws-lambda/go/events"`)
	} else {
		binds = append(binds, "")
	}
	binds = append(binds, strcase.ToCamel(name))
	if apiGateway != nil {
		binds = append(binds,
			"request events.APIGatewayProxyRequest",
			"events.APIGatewayProxyResponse",
			"events.APIGatewayProxyResponse{}",
		)
	} else {
		binds = append(binds, "request string", "string", `"Response here"`)
	}
	binds = append(binds, strcase.ToCamel(name))
	return []byte(fmt.Sprintf(MAIN_FUNCTION_TEMPLATE, binds...))
}

type Function struct {
	Name       string      `toml:"name"`
	APIGateway *APIGateway `toml:"api_gateway"`
}

func (f Function) Invoke(message []byte) ([]byte, error) {
	return nil, errors.New("Not implemented")
}

func (f Function) Save(root string) error {
	fnPath := filepath.Join(root, f.Name)
	if err := os.Mkdir(fnPath, 0755); err != nil {
		return fmt.Errorf("[Function] Directory create error: %s", err.Error())
	}
	if err := ioutil.WriteFile(filepath.Join(fnPath, "main.go"), bindTemplate(f.Name, f.APIGateway), 0644); err != nil {
		return fmt.Errorf("[Function] Create lambda function template error: %s", err.Error())
	}
	return nil
}

func (f Function) Build(root string) error {
	cmdArgs := []string{
		"build",
		"-o",
		f.Name,
	}
	cmd := exec.Command("go", cmdArgs...)
	cmd.Dir = filepath.Join(root, f.Name)
	cmd.Env = append(
		os.Environ(),
		fmt.Sprintf("GOPATH=%s:%s", os.Getenv("GOPATH"), filepath.Join(root, "vendors")),
	)
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
