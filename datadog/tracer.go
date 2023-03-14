package datadog

import (
	"context"
	"database/sql/driver"

	"github.com/luna-duclos/instrumentedsql"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type trace struct {
	traceOrphans bool
}

type span struct {
	trace
	ctx    context.Context
	parent ddtrace.Span
}

type TraceOption func(t *trace)

// NewTracer returns a tracer which will fetch spans using datadog's SpanFromContext function.
func NewTracer(opts ...TraceOption) instrumentedsql.Tracer {
	t := &trace{
		traceOrphans: false,
	}

	for _, opt := range opts {
		opt(t)
	}
	return t
}

// TraceOrphans will create spans with no parent if true, otherwise only spans with parents will be generated.
// Defaults to false, useful for preventing excessive tracing.
func TraceOrphans(traceOrphans bool) TraceOption {
	return func(t *trace) {
		t.traceOrphans = traceOrphans
	}
}

// GetSpan returns a span.
func (t trace) GetSpan(ctx context.Context) instrumentedsql.Span {
	ddSpan, ok := tracer.SpanFromContext(ctx)
	if !ok {
		return span{parent: nil, ctx: ctx, trace: t}
	}
	return span{parent: ddSpan, ctx: ctx, trace: t}
}

// NewChild starts a child span.
func (s span) NewChild(name string) instrumentedsql.Span {
	if s.parent == nil && !s.traceOrphans {
		return s
	}
	opts := []ddtrace.StartSpanOption{
		tracer.SpanType(ext.SpanTypeSQL),
		tracer.Measured(),
	}
	newSpan, ctx := tracer.StartSpanFromContext(s.ctx, name, opts...)
	return span{parent: newSpan, ctx: ctx, trace: s.trace}
}

// SetLabel sets a tag on the span.
func (s span) SetLabel(k, v string) {
	if s.parent == nil {
		return
	}
	s.parent.SetTag(k, v)
}

// SetError sets a tag with the error.
func (s span) SetError(err error) {
	if err == nil || err == driver.ErrSkip || s.parent == nil {
		return
	}
	s.parent.SetTag("err", err.Error())
}

// Finish finishes the span.
func (s span) Finish() {
	if s.parent == nil {
		return
	}
	s.parent.Finish()
}
