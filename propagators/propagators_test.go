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

package propagators_test

import (
	"context"
	"net/http"

	"github.com/batect/service-observability/propagators"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Based on test cases from https://github.com/census-ecosystem/opencensus-go-exporter-stackdriver/blob/master/propagation/http_test.go
var _ = Describe("A GCP tracing propagator", func() {
	propagator := propagators.GCPPropagator{}

	Context("when processing incoming requests", func() {
		originalSpanContext := trace.NewSpanContext(trace.SpanContextConfig{
			TraceID: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			SpanID:  [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
			Remote:  true,
		})
		originalContext := trace.ContextWithRemoteSpanContext(context.Background(), originalSpanContext)

		Context("given no X-Cloud-Trace-Context header", func() {
			var spanContext trace.SpanContext

			BeforeEach(func() {
				headers := http.Header{}
				ctx := propagator.Extract(originalContext, propagation.HeaderCarrier(headers))
				spanContext = trace.SpanContextFromContext(ctx)
			})

			It("does not modify the span context on the context", func() {
				Expect(spanContext).To(Equal(originalSpanContext))
			})
		})

		Context("given the X-Cloud-Trace-Context header contains a valid trace and span ID", func() {
			var spanContext trace.SpanContext

			BeforeEach(func() {
				headers := http.Header{
					"X-Cloud-Trace-Context": {"105445aa7843bc8bf206b12000100000/18374686479671623803"},
				}

				ctx := propagator.Extract(originalContext, propagation.HeaderCarrier(headers))
				spanContext = trace.SpanContextFromContext(ctx)
			})

			It("returns a span context with the trace and span ID extracted from the header", func() {
				Expect(spanContext).To(Equal(trace.NewSpanContext(trace.SpanContextConfig{
					TraceID: [16]byte{16, 84, 69, 170, 120, 67, 188, 139, 242, 6, 177, 32, 0, 16, 0, 0},
					SpanID:  [8]byte{255, 0, 0, 0, 0, 0, 0, 123},
					Remote:  true,
				})))
			})
		})

		Context("given the X-Cloud-Trace-Context header contains a valid trace and short span ID", func() {
			var spanContext trace.SpanContext

			BeforeEach(func() {
				headers := http.Header{
					"X-Cloud-Trace-Context": {"105445aa7843bc8bf206b12000100000/123"},
				}

				ctx := propagator.Extract(originalContext, propagation.HeaderCarrier(headers))
				spanContext = trace.SpanContextFromContext(ctx)
			})

			It("returns a span context with the trace and span ID extracted from the header", func() {
				Expect(spanContext).To(Equal(trace.NewSpanContext(trace.SpanContextConfig{
					TraceID: [16]byte{16, 84, 69, 170, 120, 67, 188, 139, 242, 6, 177, 32, 0, 16, 0, 0},
					SpanID:  [8]byte{0, 0, 0, 0, 0, 0, 0, 123},
					Remote:  true,
				})))
			})
		})

		Context("given the X-Cloud-Trace-Context header contains a valid trace and span ID and explicitly disables tracing", func() {
			var spanContext trace.SpanContext

			BeforeEach(func() {
				headers := http.Header{
					"X-Cloud-Trace-Context": {"105445aa7843bc8bf206b12000100000/18374686479671623803;o=0"},
				}

				ctx := propagator.Extract(originalContext, propagation.HeaderCarrier(headers))
				spanContext = trace.SpanContextFromContext(ctx)
			})

			It("returns a span context with the trace and span ID extracted from the header and no trace flags", func() {
				Expect(spanContext).To(Equal(trace.NewSpanContext(trace.SpanContextConfig{
					TraceID: [16]byte{16, 84, 69, 170, 120, 67, 188, 139, 242, 6, 177, 32, 0, 16, 0, 0},
					SpanID:  [8]byte{255, 0, 0, 0, 0, 0, 0, 123},
					Remote:  true,
				})))
			})
		})

		Context("given the X-Cloud-Trace-Context header contains a valid trace and span ID and explicitly enables tracing", func() {
			var spanContext trace.SpanContext

			BeforeEach(func() {
				headers := http.Header{
					"X-Cloud-Trace-Context": {"105445aa7843bc8bf206b12000100000/18374686479671623803;o=1"},
				}

				ctx := propagator.Extract(originalContext, propagation.HeaderCarrier(headers))
				spanContext = trace.SpanContextFromContext(ctx)
			})

			It("returns a span context with the trace and span ID extracted from the header and the appropriate trace flag to enable tracing", func() {
				Expect(spanContext).To(Equal(trace.NewSpanContext(trace.SpanContextConfig{
					TraceID:    [16]byte{16, 84, 69, 170, 120, 67, 188, 139, 242, 6, 177, 32, 0, 16, 0, 0},
					SpanID:     [8]byte{255, 0, 0, 0, 0, 0, 0, 123},
					TraceFlags: trace.FlagsSampled,
					Remote:     true,
				})))
			})
		})

		for _, v := range []string{
			"",
			"/",
			"c1e9153fb27f8ac9f2edac765023676e",
			"c1e9153fb27f8ac9f2edac765023676e/",
			"/13102258660371621412",
			"13102258660371621412",
			"c1e9153fb27f8ac9f2edac765023676e/;",
			"c1e9153fb27f8ac9f2edac765023676e/;o=1",
			"c1e9153fb27f8ac9f2edac765023676e/13102258660371621412;",
			"c1e9153fb27f8ac9f2edac765023676e/13102258660371621412;o",
			"c1e9153fb27f8ac9f2edac765023676e/13102258660371621412;o=",
			"c1e9153fb27f8ac9f2edac765023676e/13102258660371621412;o=2",
		} {
			headerValue := v

			Context("given the X-Cloud-Trace-Context header has the invalid value '"+headerValue+"'", func() {
				var spanContext trace.SpanContext

				BeforeEach(func() {
					headers := http.Header{
						"X-Cloud-Trace-Context": {headerValue},
					}

					ctx := propagator.Extract(originalContext, propagation.HeaderCarrier(headers))
					spanContext = trace.SpanContextFromContext(ctx)
				})

				It("does not modify the span context on the context", func() {
					Expect(spanContext).To(Equal(originalSpanContext))
				})
			})
		}
	})

	Context("when processing outgoing requests", func() {
		var headers http.Header

		BeforeEach(func() {
			headers = http.Header{}

			idGenerator := dummyIDGenerator{
				traceID: [16]byte{16, 84, 69, 170, 120, 67, 188, 139, 242, 6, 177, 32, 0, 16, 0, 0},
				spanID:  [8]byte{255, 0, 0, 0, 0, 0, 0, 123},
			}

			provider := sdktrace.NewTracerProvider(sdktrace.WithIDGenerator(idGenerator))
			tracer := provider.Tracer("Tracer")
			ctx, _ := tracer.Start(context.Background(), "Test trace")
			propagator.Inject(ctx, propagation.HeaderCarrier(headers))
		})

		It("adds a X-Cloud-Trace-Context header with the trace ID and span ID", func() {
			Expect(headers).To(HaveKeyWithValue("X-Cloud-Trace-Context", []string{"105445aa7843bc8bf206b12000100000/18374686479671623803"}))
		})
	})
})

type dummyIDGenerator struct {
	traceID [16]byte
	spanID [8]byte
}

func (d dummyIDGenerator) NewIDs(ctx context.Context) (trace.TraceID, trace.SpanID) {
	return d.traceID, d.spanID
}

func (d dummyIDGenerator) NewSpanID(ctx context.Context, traceID trace.TraceID) trace.SpanID {
	panic("not supported")
}


