package entity

type Integration struct {
	Id             string `toml:"resource_id"`
	LambdaFunction string `toml:"lambda_function"`
}
