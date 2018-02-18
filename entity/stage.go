package entity

type Stage struct {
	Name      string            `toml:"name"`
	Variables map[string]string `toml:"variables"`
}
