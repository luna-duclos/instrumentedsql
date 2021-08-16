package xray

import (
	"context"
	"database/sql/driver"

	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/luna-duclos/instrumentedsql"
)

type tracer struct{}

type span struct {
	ctx     context.Context
	segment *xray.Segment
}

// NewTracer returns a tracer that will fetch spans using opentracing's SpanFromContext function
func NewTracer() instrumentedsql.Tracer {
	return tracer{}
}

// GetSpan returns a span
func (tracer) GetSpan(ctx context.Context) instrumentedsql.Span {
	if ctx == nil {
		return span{ctx: nil}
	}

	seg := xray.GetSegment(ctx)
	if seg == nil {
		return span{ctx: nil}
	}

	return span{ctx: ctx}
}

// NewChild comply with instrumentedsql.Span
func (s span) NewChild(name string) instrumentedsql.Span {
	if s.ctx == nil {
		return s
	}

	_, seg := xray.BeginSubsegment(s.ctx, name)
	return span{ctx: s.ctx, segment: seg}
}

// SetLabel comply with instrumentedsql.Span
func (s span) SetLabel(k, v string) {
	if s.segment == nil {
		return
	}

	switch k {
	case "args":
		s.segment.GetSQL().SanitizedQuery = v
	}
}

func (s span) SetComponent(v string)        {}
func (s span) SetDbConnectionString(string) {}
func (s span) SetDBName(string)             {}
func (s span) SetDBUser(string)             {}
func (s span) SetDBSystem(string)           {}
func (s span) SetDBStatement(string)        {}
func (s span) SetDBStatementArgs(v string)  { s.SetLabel("args", v) }
func (s span) SetPeerAddress(string)        {}
func (s span) SetPeerHost(string)           {}
func (s span) SetPeerPort(string)           {}

// SetError comply with instrumentedsql.Span
func (s span) SetError(err error) {
	if err == nil || err == driver.ErrSkip {
		return
	}

	if s.segment == nil {
		return
	}

	s.segment.AddError(err)
}

// Finish comply with instrumentedsql.Span
func (s span) Finish() {
	if s.segment == nil {
		return
	}

	s.segment.Close(nil)
}
