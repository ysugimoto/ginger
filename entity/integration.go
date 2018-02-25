package entity

// Integration is the struct which maps api.resources.integration field.
// Integration type accepts on "lambda" or "s3":
//   If IntegrationType is "s3", Bucket field must not be empty.
//   If IntegrationType is "lambda", LambdaFunction must not be empty.
type Integration struct {
	Id              string  `toml:"resource_id"`
	IntegrationType string  `toml:"type"`
	LambdaFunction  *string `toml:"lambda_function"`
	Path            string  `toml:"path"`
	BucketPath      *string `toml:"bucket_path"`
	ProxyResourceId *string `toml:"proxy_resource_id"`
}

func NewIntegration(iType, value, path string) *Integration {
	i := &Integration{
		IntegrationType: iType,
		Path:            path,
	}
	switch iType {
	case "lambda":
		i.LambdaFunction = &value
	case "s3":
		i.BucketPath = &value
	}
	return i
}

func (i *Integration) String() string {
	switch i.IntegrationType {
	case "lambda":
		return i.IntegrationType + ":" + *i.LambdaFunction
	case "s3":
		return i.IntegrationType + ":" + *i.BucketPath
	default:
		return ""
	}
}
