package instrumentedsql

import "context"

// Tracer is the interface needed to be implemented by any tracing implementation we use
type Tracer interface {
	GetSpan(ctx context.Context) Span
}

// Span is part of the interface needed to be implemented by any tracing implementation we use
type Span interface {
	NewChild(string) Span
	SetLabel(string, string)
	SetComponent(string)
	SetDbConnectionString(string)
	SetDBName(string)
	SetDBUser(string)
	SetDBSystem(string)
	SetDBStatement(string)
	SetDBStatementArgs(string)
	SetPeerAddress(string)
	SetPeerHost(string)
	SetPeerPort(string)
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