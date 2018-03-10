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
func (b *builder) build(targets []*entity.Function) map[*entity.Function]string {
	log := logger.WithNamespace("ginger.build")
	binaries := make(map[*entity.Function]string)

	// Parallel build by each functions
	index := 0
	for {
		var end int
		if len(targets) < index+parallelBuildNum {
			end = len(targets)
		} else {
			end = index + 5
		}
		var wg sync.WaitGroup
		var mu sync.Mutex
		for _, fn := range targets[index:end] {
			wg.Add(1)
			log.Printf("Building function: %s...\n", fn.Name)
			bin := make(chan string)
			err := make(chan error)
			go b.compile(fn.Name, bin, err)
			go func() {
				defer func() {
					wg.Done()
					close(bin)
					close(err)
				}()
				select {
				case e := <-err:
					log.Errorf("Failed to build function: %s\n", e.Error())
					return
				case binary := <-bin:
					log.Infof("Function built successfully: %s:%s\n", fn.Name, binary)
					mu.Lock()
					binaries[fn] = binary
					mu.Unlock()
				}
			}()
		}
		wg.Wait()
		if end == len(targets) {
			break
		}
		index += parallelBuildNum
	}
	return binaries
}

// compile compiles go application by `go build` command.
// Note that runtime in AWS Lambda is linux, so we have to build as linux amd64 target.
func (b *builder) compile(name string, binChan chan string, errChan chan error) {
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
		binChan <- out
	}
}
