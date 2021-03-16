module github.com/batect/service-observability

go 1.16

require (
	cloud.google.com/go v0.79.0
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric v0.18.0
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace v0.18.0
	github.com/charleskorn/logrus-stackdriver-formatter v0.3.1
	github.com/google/uuid v1.2.0
	github.com/onsi/ginkgo v1.15.2
	github.com/onsi/gomega v1.11.0
	github.com/sirupsen/logrus v1.8.1
	go.opentelemetry.io/contrib/instrumentation/host v0.18.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.18.0
	go.opentelemetry.io/contrib/instrumentation/runtime v0.18.0
	go.opentelemetry.io/otel v0.18.0
	go.opentelemetry.io/otel/oteltest v0.18.0
	go.opentelemetry.io/otel/sdk v0.18.0
	go.opentelemetry.io/otel/sdk/metric v0.18.0
	go.opentelemetry.io/otel/trace v0.18.0
)
