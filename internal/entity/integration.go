package entity

// Integration is the struct which maps api.resources.integration field.
// Integration type accepts on "lambda" or "s3":
//    If IntegrationType is "s3", this field is empty.
//   If IntegrationType is "apigateway", this field is empty.
type Integration struct {
	Id              string `toml:"resource_id"`
	IntegrationType string `toml:"type"`
	LambdaFunction  string `toml:"lambda_function"`
	Path            string `toml:"path"`
	Bucket          string `toml:"bucket"`
}

func NewIntegration(iType, value, path string) *Integration {
	i := &Integration{
		IntegrationType: iType,
		Path:            path,
	}
	switch iType {
	case "lambda":
		i.LambdaFunction = value
	case "s3":
		i.Bucket = value
	}
	return i
}

func (i *Integration) String() string {
	switch i.IntegrationType {
	case "lambda":
		return i.IntegrationType + ":" + i.LambdaFunction
	case "s3":
		return i.IntegrationType + ":" + i.Bucket
	default:
		return ""
	}
}
