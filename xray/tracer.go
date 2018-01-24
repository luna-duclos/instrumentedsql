package xray

import (
	"context"

	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/luna-duclos/instrumentedsql"
)

const (
	labelError = "err"
	labelQuery = "query"
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
	case labelQuery:
		s.segment.GetSQL().SanitizedQuery = v
	}
}

// SetError comply with instrumentedsql.Span
func (s span) SetError(k string, err error) {
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
