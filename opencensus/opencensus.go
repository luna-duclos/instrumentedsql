package opencensus

import (
	"context"

	"go.opencensus.io/trace"

	"github.com/luna-duclos/instrumentedsql"
)

type tracer struct{}

type span struct {
	parent *trace.Span
}

// NewTracer returns a tracer that will fetch spans using google tracing's SpanContext function
func NewTracer() instrumentedsql.Tracer { return tracer{} }

// GetSpan fetches a span from the context and wraps it
func (tracer) GetSpan(ctx context.Context) instrumentedsql.Span {
	if ctx == nil {
		return span{parent: nil}
	}

	return span{parent: trace.FromContext(ctx)}
}

func (s span) NewChild(name string) instrumentedsql.Span {
	if s.parent == nil {
		return span{parent: trace.NewSpan(name, trace.StartSpanOptions{})}
	}
	return span{parent: s.parent.StartSpan(name)}
}

func (s span) SetLabel(k, v string) {
	s.parent.SetAttributes(trace.StringAttribute{Key: k, Value: v})
}

func (s span) Finish() {
	s.parent.End()
}
