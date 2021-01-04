package otels3

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// TODO: CLAIRE: SHOULD I MOVE THESE INTO CONFIG.GO?
func createCounters(meter metric.Meter) *counters {
	operationCounter, err := meter.NewInt64Counter("aws.s3.operation")
	if err != nil {
		otel.Handle(err)
	}
	return &counters{operation: operationCounter}
}

func createRecorders(meter metric.Meter) *recorders {
	execTimeRecorder, err := meter.NewFloat64ValueRecorder("aws.s3.operation.duration", metric.WithUnit("μs"))
	if err != nil {
		otel.Handle(err)
	}
	return &recorders{operationDuration: execTimeRecorder}
}

// appendSpanAndTraceIDFromSpan extracts the trace id and span id from a span using the context field.
// It returns a list of attributes with the span id and trace id appended to the original slice.
// Any modification on the original slice will modify the returned slice.
func appendSpanAndTraceIDFromSpan(attrs []label.KeyValue, span trace.Span) []label.KeyValue {
	return append(attrs,
		label.String("event.spanId", span.SpanContext().SpanID.String()),
		label.String("event.traceId", span.SpanContext().TraceID.String()),
	)
}

func AddOtelS3Handlers(s3 s3.S3, opts ...Option) {
	cfg := config{
		Tracer:            otel.GetTracerProvider().Tracer(instrumentationName),
		Meter:             otel.GetMeterProvider().Meter(instrumentationName),
		TextMapPropagator: otel.GetTextMapPropagator(),
	}
	for _, opt := range opts {
		opt.apply(&cfg)
	}

	cfg.Counters = createCounters(cfg.Meter)
	cfg.Recorders = createRecorders(cfg.Meter)

	handler := otelS3Handler(&cfg)
	s3.Handlers.Build.PushFront(handler)
}

func otelS3Handler(c *config) func(*request.Request) {
	tracer := c.Tracer

	return func(r *request.Request) {
		startTime := time.Now()
		ctx := r.Context()
		//dest := r.GetBody() //TODO: CLAIRE
		//destination := aws.StringValue(dest)
		dest := r.Params.(*s3.DeleteObjectInput).Bucket
		destination := aws.StringValue(dest)
		attrs := createAttributes(destination, r.Operation.Name)

		spanCtx, span := tracer.Start(
			ctx,
			fmt.Sprintf("%s.%s", destination, r.Operation.Name),
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(attrs...),
		)

		r.Handlers.Complete.PushBack(func(req *request.Request) {
			callReturnTime := time.Now()

			if req.Error != nil {
				attrs = append(attrs, labelStatusFailure)
				span.SetAttributes(labelStatusFailure)
				span.SetStatus(codes.Error, req.Error.Error())
			} else {
				attrs = append(attrs, labelStatusSuccess)
				span.SetAttributes(labelStatusSuccess)
				span.SetStatus(codes.Ok, "")
			}

			if c.SpanCorrelation {
				attrs = appendSpanAndTraceIDFromSpan(attrs, span)
			}

			c.Recorders.operationDuration.Record(
				spanCtx,
				float64(callReturnTime.Sub(startTime).Microseconds()),
				attrs...,
			)
			c.Counters.operation.Add(ctx, 1, attrs...)

			span.End()
		})
	}
}
