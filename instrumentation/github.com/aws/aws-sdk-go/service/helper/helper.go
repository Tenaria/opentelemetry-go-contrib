// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helper

import (
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/label"
)

func AppendSpanAndTraceIDFromSpan(attrs []label.KeyValue, span trace.Span) []label.KeyValue {
	linkSpanAttr := []label.KeyValue{
		label.String("span.id", span.SpanContext().SpanID.String()),
		label.String("trace.id", span.SpanContext().TraceID.String()),
	}

	return append(linkSpanAttr, attrs...)
}