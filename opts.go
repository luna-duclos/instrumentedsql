package instrumentedsql

type opts struct {
	Logger
	Tracer
	OpsExcluded   map[string]struct{}
	OmitArgs      bool
	ComponentName string
	DBInstance    string
	DBUser        string
}

func (o *opts) setDefaults() {
	o.Logger = nullLogger{}
	o.Tracer = nullTracer{}
	o.ComponentName = "database/sql"
	o.DBInstance = "unknown"
	o.DBUser = "unknown"
}

func (o *opts) hasOpExcluded(op string) bool {
	_, ok := o.OpsExcluded[op]
	return ok
}

func (o *opts) setDefaultLabels(span Span) {
	span.SetLabel(SpanKind, "client")
	span.SetLabel(Component, o.ComponentName)
	span.SetLabel(DBType, "sql")
	span.SetLabel(DBInstance, o.DBInstance)
	span.SetLabel(DBUser, o.DBUser)
}

// Opt is a functional option type for the wrapped driver
type Opt func(*opts)

// WithLogger sets the logger of the wrapped driver to the provided logger
func WithLogger(l Logger) Opt {
	return func(o *opts) {
		o.Logger = l
	}
}

// WithOpsExcluded excludes some of OpSQL that are not required
func WithOpsExcluded(ops ...string) Opt {
	return func(o *opts) {
		o.OpsExcluded = make(map[string]struct{})
		for _, op := range ops {
			o.OpsExcluded[op] = struct{}{}
		}
	}
}

// WithTracer sets the tracer of the wrapped driver to the provided tracer
func WithTracer(t Tracer) Opt {
	return func(o *opts) {
		o.Tracer = t
	}
}

// WithOmitArgs will make it so that query arguments are omitted from logging and tracing
func WithOmitArgs() Opt {
	return func(o *opts) {
		o.OmitArgs = true
	}
}

// WithIncludeArgs will make it so that query arguments are included in logging and tracing
// This is the default, but can be used to override WithOmitArgs
func WithIncludeArgs() Opt {
	return func(o *opts) {
		o.OmitArgs = false
	}
}

// WithComponentName allows setting the component name which are included in logging and tracing
// Default is "database/sql"
func WithComponentName(componentName string) Opt {
	return func(o *opts) {
		o.ComponentName = componentName
	}
}

// WithDBInstance sets the DB name which is included in logging and tracing
// Default is "unknown"
func WithDBInstance(dbName string) Opt {
	return func(o *opts) {
		o.DBInstance = dbName
	}
}

// WithDBUser sets the username used to access the database
// Default is "unknown"
func WithDBUser(userName string) Opt {
	return func(o *opts) {
		o.DBUser = userName
	}
}
