package entity

type APIGateway struct {
	Path   string `toml:"path"`
	Method string `toml:"method"`
}
