package entity

// Function is the entity struct which maps from configuration.
type Function struct {
	Name       string  `toml:"name"`
	Arn        string  `toml:"arn"`
	MemorySize int64   `toml:"memory_size"`
	Timeout    int64   `toml:"timeout"`
	Role       string  `toml:"role"`
	Schedule   *string `toml:"schedule"`
}
