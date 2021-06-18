module github.com/batect/service-observability

go 1.16

require (
	cloud.google.com/go v0.84.0
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric v0.20.1
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace v0.20.1
	github.com/charleskorn/logrus-stackdriver-formatter v0.3.1
	github.com/google/uuid v1.2.0
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/sirupsen/logrus v1.8.1
	go.opentelemetry.io/contrib/instrumentation/host v0.19.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.20.0
	go.opentelemetry.io/contrib/instrumentation/runtime v0.19.0
	go.opentelemetry.io/otel v0.20.0
	go.opentelemetry.io/otel/oteltest v0.20.0
	go.opentelemetry.io/otel/sdk v0.20.0
	go.opentelemetry.io/otel/sdk/metric v0.21.0
	go.opentelemetry.io/otel/trace v0.20.0
)
