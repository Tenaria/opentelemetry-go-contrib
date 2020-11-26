module go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go/service/s3/example

go 1.15

require (
	github.com/aws/aws-sdk-go v1.35.3
	go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go v0.0.0
	go.opentelemetry.io/otel v0.14.0
	go.opentelemetry.io/otel/exporters/stdout v0.14.0
	go.opentelemetry.io/otel/sdk v0.14.0
)

replace go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go v0.0.0 => ./../../../..
