package command

import (
	"bytes"
	"errors"
	"os"
	"sync"

	"os/exec"
	"path/filepath"

	"github.com/ysugimoto/ginger/internal/entity"
	"github.com/ysugimoto/ginger/internal/logger"
)

type builder struct {
	src  string
	dest string
}

func newBuilder(src, dest string) *builder {
	return &builder{
		src:  src,
		dest: dest,
	}
}

func (b *builder) build(targets entity.Functions) map[*entity.Function]string {
	log := logger.WithNamespace("ginger.build")
	binaries := make(map[*entity.Function]string)
	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, fn := range targets {
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
	return binaries
}

func (b *builder) compile(name string, binChan chan string, errChan chan error) {
	buffer := new(bytes.Buffer)
	out := filepath.Join(b.dest, name)
	src := filepath.Join(b.src, name)

	cmd := exec.Command("go", "build", "-o", out)
	cmd.Dir = src
	cmd.Env = buildEnv(map[string]string{
		"GOOS":   "linux",
		"GOARCH": "amd64",
	})
	cmd.Stdout = os.Stdout
	cmd.Stderr = buffer
	if err := cmd.Run(); err != nil {
		errChan <- errors.New(string(buffer.Bytes()))
	} else {
		binChan <- out
	}
}
