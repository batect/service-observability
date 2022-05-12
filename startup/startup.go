// Copyright 2019-2022 Charles Korn.
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

package startup

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/profiler"
	mexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric"
	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	gcppropagator "github.com/GoogleCloudPlatform/opentelemetry-operations-go/propagator"
	"github.com/batect/services-common/tracing"
	stackdriver "github.com/charleskorn/logrus-stackdriver-formatter"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/host"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func InitialiseObservability(serviceName string, serviceVersion string, projectID string) (func(), error) {
	initLogging(serviceName, serviceVersion)
	otel.SetErrorHandler(&errorHandler{})

	if err := initProfiling(serviceName, serviceVersion, projectID); err != nil {
		return nil, err
	}

	resources := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(serviceName),
		semconv.ServiceVersionKey.String(serviceVersion),
	)

	flushMetrics, err := initMetrics(projectID, resources)

	if err != nil {
		return nil, err
	}

	flushTraces, err := initTracing(projectID, resources)

	if err != nil {
		return nil, err
	}

	return func() {
		flushMetrics()
		flushTraces()
	}, nil
}

func initLogging(serviceName string, serviceVersion string) {
	logrus.SetFormatter(stackdriver.NewFormatter(
		stackdriver.WithService(serviceName),
		stackdriver.WithVersion(serviceVersion),
	))
}

func initProfiling(serviceName string, serviceVersion string, projectID string) error {
	err := profiler.Start(profiler.Config{
		Service:        serviceName,
		ServiceVersion: serviceVersion,
		ProjectID:      projectID,
		MutexProfiling: true,
	})

	if err != nil {
		return fmt.Errorf("could not create profiler: %w", err)
	}

	return nil
}

func initMetrics(projectID string, resources *resource.Resource) (func(), error) {
	pusher, err := initMetricsPipeline(projectID, resources)

	if err != nil {
		return nil, err
	}

	if err := initEnvironmentMetricsInstrumentation(); err != nil {
		return nil, err
	}

	return func() {
		logrus.Info("Flushing metrics...")

		if err := pusher.Stop(context.Background()); err != nil {
			logrus.WithError(err).Warn("Flushing metrics failed.")
		}

		logrus.Info("Flushing complete.")
	}, nil
}

func initMetricsPipeline(projectID string, resources *resource.Resource) (*controller.Controller, error) {
	opts := []mexporter.Option{
		mexporter.WithProjectID(projectID),
		mexporter.WithOnError(func(err error) {
			logrus.WithError(err).Warn("Metric exporter reported error.")
		}),
	}

	controllerOpts := []controller.Option{
		controller.WithPushTimeout(30 * time.Second),
		controller.WithResource(resources),
	}

	pusher, err := mexporter.InstallNewPipeline(opts, controllerOpts...)

	if err != nil {
		return nil, fmt.Errorf("could not install metrics pipeline: %w", err)
	}

	return pusher, nil
}

func initEnvironmentMetricsInstrumentation() error {
	if err := runtime.Start(); err != nil {
		return fmt.Errorf("could not start collecting runtime metrics: %w", err)
	}

	if err := host.Start(); err != nil {
		return fmt.Errorf("could not start collecting host metrics: %w", err)
	}

	return nil
}

func initTracing(projectID string, resources *resource.Resource) (func(), error) {
	exporter, err := texporter.New(texporter.WithProjectID(projectID))

	provider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(resources),
	)

	otel.SetTracerProvider(provider)

	if err != nil {
		return nil, fmt.Errorf("could not install tracing pipeline: %w", err)
	}

	w3Propagator := propagation.TraceContext{}
	gcpPropagator := gcppropagator.New()

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
