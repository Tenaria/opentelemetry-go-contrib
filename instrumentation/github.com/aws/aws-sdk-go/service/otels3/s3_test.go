package otels3

import (
	"bytes"
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"

	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go/service/config"
	mockmeter "go.opentelemetry.io/contrib/internal/metric"
	mocktrace "go.opentelemetry.io/contrib/internal/trace"
	"go.opentelemetry.io/otel/api/global"
)

var (
	tracerName = "mock-tracer"
)

type mockS3Client struct {
	s3iface.S3API
}

func setupClient(spanCorrelationInMetrics bool) s3iface.S3API {
	_, mockedTracer := mocktrace.NewTracerProviderAndTracer(tracerName)
	_, mockedMeter := mockmeter.NewMeter()

	mockedPropagators := global.TextMapPropagator()

	client := &instrumentedS3{
		&mockS3Client{},
		mockedTracer,
		mockedMeter,
		mockedPropagators,
		createCounters(mockedMeter),
		createRecorders(mockedMeter),
		spanCorrelationInMetrics,
	}

	return client
}

func TestS3_basicFunctionality_putObject(t *testing.T) {
	type args struct {
		context aws.Context
		input   *s3.PutObjectInput
		options []request.Option
	}

	testcases := []struct {
		name           string
		args           args
		config         *config.Config
		expectedStatus int
		expectedError  error
		client         s3iface.S3API
	}{
		{
			name:           "hi",
			expectedStatus: 200,
			expectedError:  nil,
			args: args{
				context: context.Background(),
				input: &s3.PutObjectInput{
					Bucket: aws.String("test-bucket"),
					Key:    aws.String("010101"),
					Body:   bytes.NewReader([]byte("foo")),
				},
			},
			client: setupClient(true),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {

			_, _ = tc.client.PutObjectWithContext(tc.args.context, tc.args.input, tc.args.options...)

		})
	}
}
