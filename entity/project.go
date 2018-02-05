package entity

type Project struct {
	Region              string `toml:"region"`
	Profile             string `toml:"profile"`
	LambdaExecutionRole string `toml:"lambda_execution_role"`
}
