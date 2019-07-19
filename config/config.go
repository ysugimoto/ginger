package config

import (
	"os"
	"sort"
	"strings"
	"sync"

	"path/filepath"

	"github.com/BurntSushi/toml"

	"github.com/ysugimoto/ginger/entity"
	"github.com/ysugimoto/ginger/logger"
)

// Config is the struct which maps configuration file into this.
// Ensure call Write() to update configuration.
type Config struct {
	exists        bool   `toml:"-"`
	Root          string `toml:"-"`
	Path          string `toml:"-"`
	FunctionPath  string `toml:"-"`
	LibPath       string `toml:"-"`
	StoragePath   string `toml:"-"`
	StagePath     string `toml:"-"`
	SchedulerPath string `toml:"-"`

	RestApiId         string             `toml:"rest_api_id"`
	ProjectName       string             `toml:"project_name"`
	Profile           string             `toml:"profile"`
	Region            string             `toml:"region"`
	DefaultLambdaRole string             `toml:"default_lambda_role"`
	S3BucketName      string             `toml:"s3_bucket_name"`
	DeployHookCommand string             `toml:"deploy_hook_command"`
	Resources         []*entity.Resource `toml:"resources"`
	LocalPackages     []string           `toml:"local_packages"`

	Queue map[string]*entity.Function `toml:"-"`
	log   *logger.Logger              `toml:"-"`
}

// Exists() returns bool which config file exists or not.
func (c *Config) Exists() bool {
	return c.exists
}

// Sort resources by shorter path to access by shorter resource
func (c *Config) SortResources() {
	sort.Slice(c.Resources, func(i, j int) bool {
		return len(strings.Split(c.Resources[i].Path, "/")) < len(strings.Split(c.Resources[j].Path, "/"))
	})
}

// Mutex for file I/O
var mu sync.Mutex

// Write() writes configuration to file.
func (c *Config) Write() {
	mu.Lock()
	defer mu.Unlock()
	fp, _ := os.OpenFile(c.Path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	defer fp.Close()
	enc := toml.NewEncoder(fp)
	enc.Encode(c)
	for name, fn := range c.Queue {
		p := filepath.Join(c.FunctionPath, name, "Function.toml")
		fp, _ := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		defer fp.Close()
		enc := toml.NewEncoder(fp)
		enc.Encode(fn)
	}
}
