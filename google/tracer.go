package google

import (
	"context"
	"database/sql/driver"

	"cloud.google.com/go/trace"

	"github.com/luna-duclos/instrumentedsql"
)

type tracer struct {
	traceOrphans bool
}

type span struct {
	tracer
	parent *trace.Span
}

// NewTracer returns a tracer that will fetch spans using google tracing's SpanContext function
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

	return span{parent: s.parent.NewChild(name), tracer: s.tracer}
}

func (s span) SetLabel(k, v string) {
	s.parent.SetLabel(k, v)
}

func (s span) SetError(err error) {
	if err == nil || err == driver.ErrSkip {
		return
	}

	s.parent.SetLabel("err", err.Error())
}

func (s span) Finish() {
	s.parent.Finish()
}
