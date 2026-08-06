package manager_test

import (
	"testing"

	// Packages
	otel "github.com/mutablelogic/go-client/pkg/otel"
	test "github.com/mutablelogic/go-media/task/test"
	require "github.com/stretchr/testify/require"
	tracetest "go.opentelemetry.io/otel/sdk/trace/tracetest"
	trace "go.opentelemetry.io/otel/trace"
)

// spanWithAttr returns the first span in traceExporter named name whose
// attributes include key=value, or nil if there isn't one. traceExporter
// accumulates spans from every test in the package, so tests must pick out
// their own by an attribute unique to their task (e.g. "uuid") rather than
// assuming they're the only spans recorded.
func spanWithAttr(name, key, value string) *tracetest.SpanStub {
	spans := traceExporter.GetSpans()
	for i, span := range spans {
		if span.Name != name {
			continue
		}
		for _, attr := range span.Attributes {
			if string(attr.Key) == key && attr.Value.AsString() == value {
				return &spans[i]
			}
		}
	}
	return nil
}

// TestManager_StartParentsTaskRunSpan checks that the span wrapping a
// task's execution ("Task.Run") is a genuine child of Start's own span -
// itself a child of whatever span was active on the ctx passed to Start,
// if any - so the whole chain shows up in one trace rather than requiring
// a UI to follow a cross-trace link to find it. Only cancellation is
// decoupled from ctx (via runCtx - see the doc comment on Start), not the
// trace itself.
func TestManager_StartParentsTaskRunSpan(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	task := &fakeTask{done: make(chan struct{})}
	id, err := mgr.Add(ctx, "test", task)
	require.NoError(err)

	// Simulate a caller with its own active span (e.g. an HTTP request) by
	// starting one before calling Start.
	triggerCtx, endTrigger := otel.StartSpan(tracer, ctx, "Trigger")
	triggerSC := trace.SpanContextFromContext(triggerCtx)
	require.NoError(mgr.Start(triggerCtx, id))
	endTrigger(nil)

	close(task.done)
	_, err = mgr.Wait(ctx, id)
	require.NoError(err)

	startSpan := spanWithAttr("Start", "uuid", id.String())
	runSpan := spanWithAttr("Task.Run", "uuid", id.String())
	require.NotNil(startSpan, "expected a Start span")
	require.NotNil(runSpan, "expected a Task.Run span")

	// Start's own span should be a child of Trigger - the caller's span
	// propagates all the way through, not just as far as Start.
	require.Equal(triggerSC.TraceID(), startSpan.SpanContext.TraceID())
	require.Equal(triggerSC.SpanID(), startSpan.Parent.SpanID())

	// Task.Run should be a child of Start specifically - same trace as
	// Trigger, parented directly under Start - not a separately linked span.
	require.Empty(runSpan.Links, "expected no links, now that Task.Run is a real child span")
	require.Equal(triggerSC.TraceID(), runSpan.SpanContext.TraceID())
	require.Equal(startSpan.SpanContext.SpanID(), runSpan.Parent.SpanID())
}
