package entity

// Project is the struct that maps 'project' field in configuration.
type Project struct {
	Name                string `toml:"name"`
	Region              string `toml:"region"`
	Profile             string `toml:"profile"`
	LambdaExecutionRole string `toml:"lambda_execution_role"`
}
