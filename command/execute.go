package command

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"net/rpc"
	"os/exec"
	"path/filepath"

	"github.com/aws/aws-lambda-go/lambda/messages"
	"github.com/ysugimoto/ginger/config"
)

// Execute `go xxx` command with our context
func execGoCommand(ctx context.Context, c *config.Config, name, subcommand string, arguments []string) error {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = c.LibPath
	} else {
		gopath += ":" + c.LibPath
	}
	args := []string{subcommand}
	if arguments != nil {
		args = append(args, arguments...)
	}
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = filepath.Join(c.FunctionPath, name)
	cmd.Env = buildEnv(map[string]string{
		"GOOS":                runtime.GOOS,
		"GOARCH":              "amd64",
		"GOPATH":              gopath,
		"_LAMBDA_SERVER_PORT": LAMBDARPCPORT,
	})
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Call Local Lambda RPC server like actual AWS's way
func execLambdaRPC(timeout int64, source, clientContext []byte) (*messages.InvokeResponse, error) {
	client, err := rpc.Dial("tcp", "127.0.0.1:"+LAMBDARPCPORT)
	if err != nil {
		return nil, exception("Failed to connect local lambda RPC: %s", err.Error())
	}
	defer client.Close()
	req := &messages.InvokeRequest{
		Payload:   source,
		RequestId: fmt.Sprintf("ginger-invoke-%d", time.Now().Unix()),
		Deadline: messages.InvokeRequest_Timestamp{
			Seconds: time.Now().UTC().Unix() + timeout,
			Nanos:   0,
		},
		ClientContext: clientContext,
	}
	res := &messages.InvokeResponse{}
	err = client.Call("Function.Invoke", req, res)
	return res, err
}
