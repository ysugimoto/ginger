package entity

type APIs []*APIGateway

type APIGateway struct {
	Path   string `toml:"path"`
	Method string `toml:"method"`
}
