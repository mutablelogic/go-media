package manager_test

import (
	"testing"

	// Packages
	manager "github.com/mutablelogic/go-media/task/manager"
	test "github.com/mutablelogic/go-media/task/test"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	tracetest "go.opentelemetry.io/otel/sdk/trace/tracetest"
	trace "go.opentelemetry.io/otel/trace"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

// traceExporter and tracer back the shared test manager's tracing for the
// whole package's test run (see tracing_test.go). Since they're shared
// across every test, a test that inspects traceExporter must filter by an
// attribute unique to its own task (e.g. "uuid") rather than assuming it's
// the only spans present.
var (
	traceExporter = tracetest.NewInMemoryExporter()
	tracer        trace.Tracer
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func TestMain(m *testing.M) {
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(traceExporter))
	tracer = provider.Tracer("test")
	test.Main(m, nil, manager.WithTracer(tracer))
}
