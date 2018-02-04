package command

import (
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/ysugimoto/ginger/entity"
)

var buildEnv []string

func init() {
	cwd, _ := os.Getwd()
	vendorPath := filepath.Join(cwd, "vendor")

	var found bool
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "GOPATH=") {
			found = true
			buildEnv = append(buildEnv, fmt.Sprintf("%s:%s", e, vendorPath))
		} else {
			buildEnv = append(buildEnv, e)
		}
	}
	if !found {
		buildEnv = append(buildEnv, fmt.Sprintf("GOPATH=%s", vendorPath))
	}
}

type builder struct {
	src  string
	dest string
}

func newBuilder(src, dest string) {
	return &builder{
		src:  src,
		dest: dest,
	}
}

func (b *builder) build(targets []entity.Functions) map[*entity.Function]string {
	binaries := make(map[*entity.Function]string)
	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, fn := range targets {
		wg.Add(1)
		bin := make(chan string)
		err := make(chan error)
		go builder.compile(fn.Name, bin, err, &wg)
		go func() {
			defer func() {
				close(bin)
				close(err)
			}()
			select {
			case <-err:
				return
			case binary := <-bin:
				mu.Lock()
				binaries[fn] = bin
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	return binaries
}

func (b *builder) compile(name string, bin chan string, err chan error, wg *sync.WaitGroup) {
	defer wg.Done()
	out := filepath.Join(b.dest, name)
	src := filepath.Join(b.src, name)

	cmd := exec.Command("go", []string{"build", "-o", out})
	cmd.Dir = src
	cmd.Env = buildEnv
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		err <- err
	} else {
		bin <- src
	}
}
