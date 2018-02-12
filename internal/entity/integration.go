package entity

// Integration is the struct which maps api.resources.integration field.
type Integration struct {
	// Integration type accepts on "lambda" or "s3"
	IntegrationType string `toml:"type"`
	Id              string `toml:"resource_id"`
	// If IntegrationType is "s3", this field is empty
	LambdaFunction string `toml:"lambda_function"`
	// If IntegrationType is "apigateway", this field is empty
	Bucket string `toml:"bucket"`
}

func (i *Integration) Method() string {
	switch i.IntegrationType {
	case "lambda":
		return "ANY"
	case "s3":
		return "GET"
	default:
		return "GET"
	}
}

func (i *Integration) ProxyPathPart() string {
	switch i.IntegrationType {
	case "lambda":
		return "{proxy+}"
	case "s3":
		return "{object}"
	default:
		return "{object}"
	}
}
