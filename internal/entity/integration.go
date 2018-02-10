package entity

// Integration is the struct which maps api.resources.integration field.
type Integration struct {
	Id             string `toml:"resource_id"`
	LambdaFunction string `toml:"lambda_function"`
}
