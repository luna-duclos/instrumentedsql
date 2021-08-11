package opentracing

import (
	"context"
	"database/sql/driver"

	"github.com/luna-duclos/instrumentedsql"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
)

type tracer struct {
	traceOrphans bool
}

type span struct {
	tracer
	parent opentracing.Span
}

// NewTracer returns a tracer that will fetch spans using opentracing's SpanFromContext function
// if traceOrphans is set to true, then spans with no parent will be traced anyway, if false, they will not be.
func NewTracer(traceOrphans bool) instrumentedsql.Tracer { return tracer{traceOrphans: traceOrphans} }

// GetSpan returns a span
func (t tracer) GetSpan(ctx context.Context) instrumentedsql.Span {
	if ctx == nil {
		return span{parent: nil, tracer: t}
	}

	return span{parent: opentracing.SpanFromContext(ctx), tracer: t}
}

func (s span) NewChild(name string) instrumentedsql.Span {
	var child span
	if s.parent == nil {
		if s.traceOrphans {
			child = span{parent: opentracing.StartSpan(name), tracer: s.tracer}
		} else {
			child = s
		}
	} else {
		child = span{parent: s.parent.Tracer().StartSpan(name, opentracing.ChildOf(s.parent.Context())), tracer: s.tracer}
	}

	child.SetLabel("span.kind", "client")
	child.SetLabel("db.type", "sql")

	return child
}

func (s span) SetLabel(k, v string) {
	if s.parent == nil {
		return
	}
	s.parent.SetTag(k, v)
}

func (s span) SetComponent(v string) {
	s.SetLabel("component", v)
}

func (s span) SetDBName(v string) {
	s.SetLabel("db.instance", v)
}

func (s span) SetDBUser(v string) {
	s.SetLabel("db.user", v)
}

func (s span) SetDBSystem(v string) {
	s.SetLabel("db.system", v)
}

func (s span) SetDBStatement(v string) {
	s.SetLabel("db.statement", v)
}

func (s span) SetDBStatementArgs(v string) {
	s.SetLabel("db.statement.args", v)
}

func (s span) SetError(err error) {
	if err == nil || err == driver.ErrSkip {
		return
	}

	if s.parent == nil {
		return
	}

	ext.Error.Set(s.parent, true)
	s.parent.LogFields(
		log.String("event", "error"),
		log.String("message", err.Error()),
	)
}

func (s span) Finish() {
	if s.parent == nil {
		return
	}
	s.parent.Finish()
}
