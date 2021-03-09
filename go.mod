module github.com/batect/service-observability

go 1.16

require (
	cloud.google.com/go v0.78.0
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric v0.17.1-0.20210308202401-8aa889a7f9a8
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace v0.17.1-0.20210308202401-8aa889a7f9a8
	github.com/charleskorn/logrus-stackdriver-formatter v0.3.1
	github.com/google/uuid v1.2.0
	github.com/onsi/ginkgo v1.15.1
	github.com/onsi/gomega v1.11.0
	github.com/sirupsen/logrus v1.8.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.18.0
	go.opentelemetry.io/otel v0.18.0
	go.opentelemetry.io/otel/oteltest v0.18.0
	go.opentelemetry.io/otel/sdk v0.18.0
	go.opentelemetry.io/otel/sdk/metric v0.18.0
	go.opentelemetry.io/otel/trace v0.18.0
)
