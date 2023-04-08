// Copyright 2019-2023 Charles Korn.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package startup

import (
	"context"
	"fmt"
	"net/http"

	"cloud.google.com/go/profiler"
	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	gcppropagator "github.com/GoogleCloudPlatform/opentelemetry-operations-go/propagator"
	"github.com/batect/services-common/tracing"
	stackdriver "github.com/charleskorn/logrus-stackdriver-formatter"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"google.golang.org/grpc/credentials"
)

func InitialiseObservability(serviceName string, serviceVersion string, gcpProjectID string, honeycombAPIKey string) (func(), error) {
	initLogging(serviceName, serviceVersion)
	otel.SetErrorHandler(&errorHandler{})

	if err := initProfiling(serviceName, serviceVersion, gcpProjectID); err != nil {
		return nil, err
	}

	resources := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(serviceName),
		semconv.ServiceVersionKey.String(serviceVersion),
	)

	flushTraces, err := initTracing(gcpProjectID, honeycombAPIKey, resources)

	if err != nil {
		return nil, err
	}

	return func() {
		flushTraces()
	}, nil
}

func initLogging(serviceName string, serviceVersion string) {
	logrus.SetFormatter(stackdriver.NewFormatter(
		stackdriver.WithService(serviceName),
		stackdriver.WithVersion(serviceVersion),
	))
}

func initProfiling(serviceName string, serviceVersion string, gcpProjectID string) error {
	err := profiler.Start(profiler.Config{
		Service:        serviceName,
		ServiceVersion: serviceVersion,
		ProjectID:      gcpProjectID,
		MutexProfiling: true,
	})

	if err != nil {
		return fmt.Errorf("could not create profiler: %w", err)
	}

	return nil
}

func createHoneycombExporter(apiKey string) (*otlptrace.Exporter, error) {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint("api.honeycomb.io:443"),
		otlptracegrpc.WithHeaders(map[string]string{
			"x-honeycomb-team": apiKey,
		}),
		otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")),
	}

	client := otlptracegrpc.NewClient(opts...)

	return otlptrace.New(context.Background(), client)
}

func initTracing(gcpProjectID string, honeycombAPIKey string, resources *resource.Resource) (func(), error) {
	gcpExporter, err := texporter.New(texporter.WithProjectID(gcpProjectID))

	if err != nil {
		return nil, fmt.Errorf("could not create GCP tracing exporter: %w", err)
	}

	honeycombExporter, err := createHoneycombExporter(honeycombAPIKey)

	if err != nil {
		return nil, fmt.Errorf("could not create Honeycomb tracing exporter: %w", err)
	}

	provider := trace.NewTracerProvider(
		trace.WithBatcher(gcpExporter),
		trace.WithBatcher(honeycombExporter),
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(resources),
	)

	otel.SetTracerProvider(provider)

	w3Propagator := propagation.TraceContext{}
	gcpPropagator := gcppropagator.CloudTraceOneWayPropagator{}

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(w3Propagator, gcpPropagator))

	http.DefaultTransport = otelhttp.NewTransport(
		http.DefaultTransport,
		otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
		otelhttp.WithSpanNameFormatter(tracing.NameHTTPRequestSpan),
	)

	return func() {
		logrus.Info("Flushing remaining traces...")

		if err := provider.Shutdown(context.Background()); err != nil {
			logrus.WithError(err).Warning("Shutting down tracing provider failed with error.")
		}

		logrus.Info("Flushing complete.")
	}, nil
}

type errorHandler struct{}

func (e *errorHandler) Handle(err error) {
	logrus.WithError(err).Warn("OpenTelemetry reported error.")
}
