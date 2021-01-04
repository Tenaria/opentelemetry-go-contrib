module go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go/service/s3/otels3

go 1.15

require (
	github.com/aws/aws-sdk-go v1.35.3
	github.com/opentracing/opentracing-go v1.2.0
	github.com/stretchr/testify v1.6.1
	go.opencensus.io v0.22.5
	go.opentelemetry.io/otel v0.15.0
)

replace (
	go.opentelemetry.io/contrib => ../../../..
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp => ../../../net/http/otelhttp
	go.opentelemetry.io/contrib/propagators => ../../../../propagators
)
