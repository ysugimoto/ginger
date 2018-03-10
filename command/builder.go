package command

import (
	"bytes"
	"errors"
	"os"
	"sync"

	"os/exec"
	"path/filepath"

	"github.com/ysugimoto/ginger/entity"
	"github.com/ysugimoto/ginger/logger"
)

const parallelBuildNum = 5

// builder builds go application dynamically.
// It's funny go application executes `go build` command :-)
type builder struct {
	src     string
	dest    string
	libPath string
}

func newBuilder(src, dest, libPath string) *builder {
	return &builder{
		src:     src,
		dest:    dest,
		libPath: libPath,
	}
}

// build builds go application by each functions
func (b *builder) build(targets []*entity.Function) error {
	log := logger.WithNamespace("ginger.build")

	// Parallel build by each functions
	index := 0
	errorBuilds := []error{}
	for {
		var end int
		if len(targets) < index+parallelBuildNum {
			end = len(targets)
		} else {
			end = index + 5
		}
		var wg sync.WaitGroup
		// var mu sync.Mutex
		for _, fn := range targets[index:end] {
			wg.Add(1)
			log.Printf("Building function: %s...\n", fn.Name)
			success := make(chan struct{})
			err := make(chan error)
			go b.compile(fn.Name, success, err)
			go func() {
				defer func() {
					wg.Done()
					close(success)
					close(err)
				}()
				select {
				case e := <-err:
					log.Errorf("Failed to build function: %s\n", e.Error())
					errorBuilds = append(errorBuilds, e)
					return
				case <-success:
					log.Infof("Function %s built successfully\n", fn.Name)
					return
				}
			}()
		}
		wg.Wait()
		if end == len(targets) {
			break
		}
		index += parallelBuildNum
	}

	if len(errorBuilds) > 0 {
		return errors.New("Either function build failed")
	}
	return nil
}

// compile compiles go application by `go build` command.
// Note that runtime in AWS Lambda is linux, so we have to build as linux amd64 target.
func (b *builder) compile(name string, successChan chan struct{}, errChan chan error) {
	buffer := new(bytes.Buffer)
	out := filepath.Join(b.dest, name)
	src := filepath.Join(b.src, name)

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = b.libPath
	} else {
		gopath += ":" + b.libPath
	}
	cmd := exec.Command("go", "build", "-o", out)
	cmd.Dir = src
	cmd.Env = buildEnv(map[string]string{
		"GOOS":   "linux",
		"GOARCH": "amd64",
		"GOPATH": gopath,
	})
	cmd.Stdout = os.Stdout
	cmd.Stderr = buffer
	if err := cmd.Run(); err != nil {
		errChan <- errors.New(string(buffer.Bytes()))
	} else {
		successChan <- struct{}{}
	}
}
