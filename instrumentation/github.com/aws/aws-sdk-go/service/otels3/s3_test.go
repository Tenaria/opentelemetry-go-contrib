package otels3

import (
	"bytes"
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"

	"bitbucket.org/observability/obsvs-go/instrumentation/github.com/aws/aws-sdk-go/service/config"
	mockmeter "go.opentelemetry.io/contrib/internal/metric"
	mocktrace "go.opentelemetry.io/contrib/internal/trace"
	"go.opentelemetry.io/otel/api/global"
)

var (
	tracerName = "mock-tracer"
)

func setupClient(spanCorrelationInMetrics bool) s3iface.S3API {
	_, mockedTracer := mocktrace.NewTracerProviderAndTracer(tracerName)
	_, mockedMeter := mockmeter.NewMeter()

	mockedPropagators = global.TextMapPropagator()

	client := &instrumentedS3(
		&mockS3Client{},
		mockedMeter,
		mockedTracer,
		mockedPropagators,
		createCounters(),
		createRecorders(),
		global.TextMapPropagator(),
		spanCorrelationInMetrics,
	)

	return client
}

func TestS3_basicFunctionality_putObject(t *testing.T) {
	type args struct {
		ctx   aws.Context
		input *s3.PutObjectInput
		opts  []request.Option
	}

	testcases := []struct {
		name               string
		args               args
		config             *config.Config
		expectedStatusCode int
		expectedError      error
		client             *instrumentedS3
	}{
		{
			testName:       "hi",
			expectedStatus: 200,
			expectedError:  nil,
			args: args{
				context: context.Background(),
				input: &s3.PutObjectInput{
					Bucket: aws.String("test-bucket"),
					Key:    aws.String("010101"),
					Body:   bytes.NewReader([]byte("foo")),
				},
				opts: [],
			},
			client: setupClient(true),
		},
		{
			testName:       "hi2",
			expectedStatus: 200,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.testName, func(t *testing.T) {

			output, err := client.PutObjectWithContext(tc.args.context, tc.args.input, tc.args.options)

		})
	}
}
