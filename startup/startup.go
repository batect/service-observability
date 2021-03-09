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

package startup

import (
	"context"
	"fmt"
	"net/http"

	"cloud.google.com/go/profiler"
	mexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric"
	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/batect/service-observability/propagators"
	"github.com/batect/service-observability/tracing"
	stackdriver "github.com/charleskorn/logrus-stackdriver-formatter"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	"go.opentelemetry.io/otel/sdk/trace"
)

func InitialiseObservability(serviceName string, serviceVersion string, projectID string) (func(), error) {
	initLogging(serviceName, serviceVersion)

	if err := initProfiling(serviceName, serviceVersion, projectID); err != nil {
		return nil, err
	}

	flushMetrics, err := initMetrics(projectID)

	if err != nil {
		return nil, err
	}

	flushTraces, err := initTracing(projectID)

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

func initMetrics(projectID string) (func(), error) {
	opts := []mexporter.Option{
		mexporter.WithProjectID(projectID),
		mexporter.WithOnError(func(err error) {
			logrus.WithError(err).Warn("Metric exporter reported error.")
		}),
	}

	popts := []controller.Option{}

	pusher, err := mexporter.InstallNewPipeline(opts, popts...)

	if err != nil {
		return nil, fmt.Errorf("could not install metrics pipeline: %w", err)
	}

	return func() {
		logrus.Info("Flushing metrics...")

		if err := pusher.Stop(context.Background()); err != nil {
			logrus.WithError(err).Warn("Flushing metrics failed.")
		}

		logrus.Info("Flushing complete.")
	}, nil
}

func initTracing(projectID string) (func(), error) {
	_, flush, err := texporter.InstallNewPipeline(
		[]texporter.Option{
			texporter.WithProjectID(projectID),
			texporter.WithOnError(func(err error) {
				logrus.WithError(err).Warn("Trace exporter reported error.")
			}),
		},
		trace.WithConfig(trace.Config{DefaultSampler: trace.AlwaysSample()}),
	)

	if err != nil {
		return nil, fmt.Errorf("could not install tracing pipeline: %w", err)
	}

	w3Propagator := propagation.TraceContext{}
	gcpPropagator := propagators.GCPPropagator{}

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(w3Propagator, gcpPropagator))

	http.DefaultTransport = otelhttp.NewTransport(
		http.DefaultTransport,
		otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
		otelhttp.WithSpanNameFormatter(tracing.NameHTTPRequestSpan),
	)

	return func() {
		logrus.Info("Flushing remaining traces...")
		flush()
		logrus.Info("Flushing complete.")
	}, nil
}
