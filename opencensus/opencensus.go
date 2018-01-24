package opencensus

import (
	"context"

	"go.opencensus.io/trace"

	"github.com/luna-duclos/instrumentedsql"
)

type tracer struct {
	traceOrphans bool
}

type span struct {
	tracer
	parent *trace.Span
}

// NewTracer returns a tracer that will fetch spans using opencensus's FromContext function
// if traceOrphans is set to true, then spans with no parent will be traced anyway, if false, they will not be.
func NewTracer(traceOrphans bool) instrumentedsql.Tracer { return tracer{traceOrphans: traceOrphans} }

// GetSpan fetches a span from the context and wraps it
func (t tracer) GetSpan(ctx context.Context) instrumentedsql.Span {
	if ctx == nil {
		return span{parent: nil, tracer: t}
	}

	return span{parent: trace.FromContext(ctx), tracer: t}
}

func (s span) NewChild(name string) instrumentedsql.Span {
	if s.parent == nil && !s.traceOrphans {
		return s
	}

	return span{parent: s.parent.StartSpan(name), tracer: s.tracer}
}

func (s span) SetLabel(k, v string) {
	s.parent.SetAttributes(trace.StringAttribute{Key: k, Value: v})
}

func (s span) SetError(k string, err error) {
	s.parent.SetAttributes(trace.StringAttribute{Key: k, Value: err.Error()})
}

func (s span) Finish() {
	s.parent.End()
}
