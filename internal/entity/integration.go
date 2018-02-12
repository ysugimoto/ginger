package entity

// Integration is the struct which maps api.resources.integration field.
// Integration type accepts on "lambda" or "s3":
//    If IntegrationType is "s3", this field is empty.
//   If IntegrationType is "apigateway", this field is empty.
type Integration struct {
	Id              string `toml:"resource_id"`
	IntegrationType string `toml:"type"`
	LambdaFunction  string `toml:"lambda_function"`
	Bucket          string `toml:"bucket"`
}
