// Copyright 2019-2021 Charles Korn.
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

package propagators

import (
	"context"
	"encoding/binary"
	"fmt"
	"regexp"
	"strconv"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const headerName = "X-Cloud-Trace-Context"

type GCPPropagator struct{}

func (c GCPPropagator) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	headerValue := carrier.Get(headerName)

	if ok, sc := c.extractSpanContext(headerValue); ok {
		return trace.ContextWithRemoteSpanContext(ctx, sc)
	}

	return ctx
}

// See https://cloud.google.com/trace/docs/setup#force-trace for a description of the X-Cloud-Trace-Context header,
// and https://github.com/census-ecosystem/opencensus-go-exporter-stackdriver/blob/master/propagation/http.go for the OpenCensus implementation.
func (c GCPPropagator) extractSpanContext(headerValue string) (bool, trace.SpanContext) {
	regex := regexp.MustCompile(`^([\da-fA-F]{32})/(\d+)(?:;o=([01]))?$`)
	segments := regex.FindStringSubmatch(headerValue)

	if segments == nil {
		return false, trace.SpanContext{}
	}

	tid, err := trace.TraceIDFromHex(segments[1])

	if err != nil {
		return false, trace.SpanContext{}
	}

	sid, err := strconv.ParseUint(segments[2], 10, 64)

	if err != nil {
		return false, trace.SpanContext{}
	}

	sidBytes := trace.SpanID{}
	binary.BigEndian.PutUint64(sidBytes[:], sid)

	flags := trace.TraceFlags(0)

	if segments[3] == "1" {
		flags = flags.WithSampled(true)
	}

	return true, trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    tid,
		SpanID:     sidBytes,
		TraceFlags: flags,
	})
}

func (c GCPPropagator) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	sc := trace.SpanFromContext(ctx).SpanContext()
	spanID := sc.SpanID()
	sid := binary.BigEndian.Uint64(spanID[:])
	headerValue := fmt.Sprintf("%v/%v", sc.TraceID().String(), sid)

	if sc.IsSampled() {
		headerValue += ";o=1"
	}

	carrier.Set(headerName, headerValue)
}

func (c GCPPropagator) Fields() []string {
	return []string{headerName}
}
