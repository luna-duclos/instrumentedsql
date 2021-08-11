package instrumentedsql

import "context"

// Tracer is the interface needed to be implemented by any tracing implementation we use
type Tracer interface {
	GetSpan(ctx context.Context) Span
}

// Span is part of the interface needed to be implemented by any tracing implementation we use
type Span interface {
	NewChild(string) Span
	SetLabel(k, v string)
	SetComponent(v string)
	SetDBName(v string)
	SetDBUser(v string)
	SetDBSystem(v string)
	SetDBStatement(v string)
	SetDBStatementArgs(v string)
	SetError(err error)
	Finish()
}

type spanFinisher interface {
	Finish(ctx context.Context, err error)
}

type childSpanFactory interface {
	NewChildSpan(ctx context.Context, operation string) spanFinisher
	NewChildSpanWithQuery(ctx context.Context, operation string, query string, args interface{}) spanFinisher
}