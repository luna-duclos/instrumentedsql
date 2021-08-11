package instrumentedsql

type opts struct {
	Logger
	Tracer
	opsExcluded   map[string]struct{}
	omitArgs      bool
	componentName string
	dbName   string
	dbSystem string
	dbUser   string
}

func (o *opts) hasOpExcluded(op string) bool {
	_, ok := o.opsExcluded[op]
	return ok
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
		o.opsExcluded = make(map[string]struct{})
		for _, op := range ops {
			o.opsExcluded[op] = struct{}{}
		}
	}
}

// WithTracer sets the tracer of the wrapped driver to the provided tracer
func WithTracer(t Tracer) Opt {
	return func(o *opts) {
		o.Tracer = t
	}
}

// WithIncludeArgs will make it so that query arguments are included in logging and tracing
// Default is not to include the args (for security reasons)
func WithIncludeArgs() Opt {
	return func(o *opts) {
		o.omitArgs = false
	}
}

// WithComponentName allows setting the component name which are included in logging and tracing
// Default is "database/sql"
func WithComponentName(componentName string) Opt {
	return func(o *opts) {
		o.componentName = componentName
	}
}

// WithDBName sets the DB name which is included in logging and tracing
// Default is "unknown"
func WithDBName(dbName string) Opt {
	return func(o *opts) {
		o.dbName = dbName
	}
}

// WithDBUser sets the username used to access the database
// Default is "unknown"
func WithDBUser(userName string) Opt {
	return func(o *opts) {
		o.dbUser = userName
	}
}

// WithDBSystem sets the db system used
// Default is "unknown"
func WithDBSystem(dbSystem string) Opt {
	return func(o *opts) {
		o.dbSystem = dbSystem
	}
}
