package instrumentedsql

// Standard tags for databases according to the opentracing specification
const (
	Component       = "component"
	DBInstance      = "db.instance"
	DBStatement     = "db.statement"
	DBType          = "db.type"
	DBUser          = "db.user"
	SpanKind        = "span.kind"
)

// Internal tags
const (
	DBStatementArgs = "args"
)
