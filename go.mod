module github.com/batect/service-observability

go 1.16

require (
	cloud.google.com/go v0.93.3
	cloud.google.com/go/monitoring v0.1.0 // indirect
	cloud.google.com/go/profiler v0.93.3
	cloud.google.com/go/trace v0.1.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric v0.22.0
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace v1.0.0-RC2
	github.com/charleskorn/logrus-stackdriver-formatter v0.3.1
	github.com/google/uuid v1.3.0
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.15.0
	github.com/sirupsen/logrus v1.8.1
	go.opentelemetry.io/contrib/instrumentation/host v0.21.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.21.0
	go.opentelemetry.io/contrib/instrumentation/runtime v0.21.0
	go.opentelemetry.io/otel v1.0.0-RC2
	go.opentelemetry.io/otel/oteltest v1.0.0-RC2
	go.opentelemetry.io/otel/sdk v1.0.0-RC2
	go.opentelemetry.io/otel/sdk/metric v0.22.0
	go.opentelemetry.io/otel/trace v1.0.0-RC2
)
