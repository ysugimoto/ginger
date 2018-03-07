package entity

// Scheduler is the entity struct which maps from configuration.
type Scheduler struct {
	Name       string   `toml:"name"`
	Arn        string   `toml:"arn"`
	Enable     bool     `toml:"enable"`
	Expression string   `toml:"expression"`
	Functions  []string `toml:"functions"`
}
