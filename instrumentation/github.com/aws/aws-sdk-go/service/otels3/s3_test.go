package otels3

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go/service/otels3/mocks"
	mockmetric "go.opentelemetry.io/contrib/internal/metric"
	mocktrace "go.opentelemetry.io/contrib/internal/trace"
	"go.opentelemetry.io/otel/label"

	"reflect"
	"testing"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/trace"
)

var (
	s3bucket = "s3bucket"
)

func getLabelValFromMeasurementBatch(key label.Key, batch mockmetric.Batch) *label.Value {
	for _, label := range batch.Labels {
		if label.Key == key {
			return &label.Value
		}
	}
	return nil
}

func getLabelValFromSpan(key label.Key, span mocktrace.Span) *label.Value {
	if value, ok := span.Attributes[key]; ok {
		return &value
	}
	return nil
}

func Test_instrumentedS3_PutObjectWithContext(t *testing.T) {
	type fields struct {
		spanCorrelationInMetrics bool
		mockSetup                func(s3Client *mock.Mock) (expectedReturn interface{})
	}
	type args struct {
		ctx   aws.Context
		input *s3.PutObjectInput
		opts  []request.Option
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "instrumentedS3.PutObjectWithContext should be delegated to S3.PutObjectWithContext while metrics and spans are linked",
			fields: fields{
				spanCorrelationInMetrics: true,
				mockSetup: func(m *mock.Mock) (expectedReturn interface{}) {
					expectedReturn = &s3.PutObjectOutput{}
					m.On("PutObjectWithContext", mock.Anything, mock.Anything).Return(expectedReturn, nil)
					return
				},
			},
			args: args{
				ctx: context.Background(),
				input: &s3.PutObjectInput{
					Bucket: &s3bucket,
				},
				opts: nil,
			},
			wantErr: false,
		},
		{
			name: "instrumentedS3.PutObjectWithContext should be delegated to S3.PutObjectWithContext while metrics and spans are NOT linked",
			fields: fields{
				spanCorrelationInMetrics: false,
				mockSetup: func(m *mock.Mock) (expectedReturn interface{}) {
					expectedReturn = &s3.PutObjectOutput{}
					m.On("PutObjectWithContext", mock.Anything, mock.Anything).Return(expectedReturn, nil)
					return
				},
			},
			args: args{
				ctx: context.Background(),
				input: &s3.PutObjectInput{
					Bucket: &s3bucket,
				},
				opts: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, mockedTracer := mocktrace.NewTracerProviderAndTracer("github.com/aws/aws-sdk-go/aws/service/s3")
			mockedMeterImp, mockedMeter := mockmetric.NewMeter()
			mockedCounters := createCounters(mockedMeter)
			mockedRecorders := createRecorders(mockedMeter)
			mockedPropagators := global.TextMapPropagator()

			s3Mock := &mocks.S3Client{}
			s := &instrumentedS3{
				S3API:                    s3Mock,
				tracer:                   mockedTracer,
				meter:                    mockedMeter,
				propagators:              mockedPropagators,
				counters:                 mockedCounters,
				recorders:                mockedRecorders,
				spanCorrelationInMetrics: tt.fields.spanCorrelationInMetrics,
			}
			expectedReturn := tt.fields.mockSetup(&s3Mock.S3API.Mock)
			got, err := s.PutObjectWithContext(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("PutObjectWithContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, expectedReturn) {
				t.Errorf("PutObjectWithContext() got = %v, want %v", got, expectedReturn)
			}
			spans := mockedTracer.EndedSpans()
			assert.Equal(t, 1, len(spans))
			assert.Equal(t, trace.SpanKindClient, spans[0].Kind)
			assert.Equal(t, s3StorageSystemValue, getLabelValFromSpan(storageSystemKey, *spans[0]).AsString())
			assert.Equal(t, *tt.args.input.Bucket, getLabelValFromSpan(storageDestinationKey, *spans[0]).AsString())
			assert.Equal(t, operationPutObject, getLabelValFromSpan(storageOperationKey, *spans[0]).AsString())

			// In Meter we have one duration recorder, one operation counter
			assert.Equal(t, 2, len(mockedMeterImp.MeasurementBatches))
			assert.Equal(t, "storage.operation.duration_μs", mockedMeterImp.MeasurementBatches[0].Measurements[0].Instrument.Descriptor().Name())
			assert.Equal(t, "storage.s3.operation", mockedMeterImp.MeasurementBatches[1].Measurements[0].Instrument.Descriptor().Name())

			if tt.fields.spanCorrelationInMetrics {
				traceID := spans[0].SpanContext().TraceID.String()
				spanID := spans[0].SpanContext().SpanID.String()

				assert.Equal(t, traceID, getLabelValFromMeasurementBatch("trace.id", mockedMeterImp.MeasurementBatches[0]).AsString())
				assert.Equal(t, spanID, getLabelValFromMeasurementBatch("span.id", mockedMeterImp.MeasurementBatches[0]).AsString())

				assert.Equal(t, traceID, getLabelValFromMeasurementBatch("trace.id", mockedMeterImp.MeasurementBatches[1]).AsString())
				assert.Equal(t, spanID, getLabelValFromMeasurementBatch("span.id", mockedMeterImp.MeasurementBatches[1]).AsString())
			} else {
				assert.Nil(t, getLabelValFromMeasurementBatch("trace.id", mockedMeterImp.MeasurementBatches[0]))
				assert.Nil(t, getLabelValFromMeasurementBatch("span.id", mockedMeterImp.MeasurementBatches[0]))

				assert.Nil(t, getLabelValFromMeasurementBatch("trace.id", mockedMeterImp.MeasurementBatches[1]))
				assert.Nil(t, getLabelValFromMeasurementBatch("span.id", mockedMeterImp.MeasurementBatches[1]))
			}
		})
	}
}
