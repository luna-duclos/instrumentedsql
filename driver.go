package instrumentedsql

import (
	"context"
	"database/sql/driver"
	"time"
)

// WrappedDriver wraps a driver and adds instrumentation.
// Use WrapDriver to create a new WrappedDriver.
type WrappedDriver struct {
	opts
	parent driver.Driver
}

// Compile time validation that our types implement the expected interfaces
var (
	_ driver.Driver = WrappedDriver{}
)

// WrapDriver will wrap the passed SQL driver and return a new sql driver that uses it and also logs
// and traces calls using the passed logger and tracer. The returned driver will still have to be
// registered with the sql package before it can be used.
//
// Important note: Seeing as the context passed into the various instrumentation calls this package
// calls. Any call without a context passed will not be instrumented. Please be sure to use the
// ___Context() and BeginTx() function calls added in Go 1.8 instead of the older calls which do not
// accept a context.
func WrapDriver(driver driver.Driver, opts ...Opt) WrappedDriver {
	d := WrappedDriver{
		parent: driver,
	}
	d.Logger = nullLogger{}
	d.Tracer = nullTracer{}
	d.omitArgs = true
	d.componentName = "database/sql"

	for _, opt := range opts {
		opt(&d.opts)
	}

	return d
}

// Open implements the database/sql/driver.Driver interface for WrappedDriver.
func (d WrappedDriver) Open(name string) (driver.Conn, error) {
	conn, err := d.parent.Open(name)
	if err != nil {
		return nil, err
	}

	var details dbConnDetails
	if !d.omitDbConnectionTags {
		details = newDBConnDetails(name)
	}

	return wrappedConn{
		Logger: d.opts.Logger,
		childSpanFactory: childSpanFactoryImpl{
			opts:          d.opts,
			dbConnDetails: details,
		},
		parent: conn,
	}, nil
}

type spanFinisherImpl struct {
	Logger
	omitArgs  bool
	operation string
	span      Span
	query     *string
	args      interface{}
	start     time.Time
}

func (f *spanFinisherImpl) Finish(ctx context.Context, err error) {
	f.span.SetError(err)
	f.span.Finish()

	keyvals := []interface{}{
		"err", err,
		"duration", time.Since(f.start),
	}

	if f.query != nil {
		keyvals = append(keyvals, "query", *f.query)
	}

	if !f.omitArgs && f.args != nil {
		keyvals = append(keyvals, "args", formatArgs(f.args))
	}

	f.Log(ctx, f.operation, keyvals...)
}

type childSpanFactoryImpl struct {
	opts
	dbConnDetails
}

func (c childSpanFactoryImpl) NewChildSpan(ctx context.Context, operation string) spanFinisher {
	if !c.hasOpExcluded(operation) {
		span := c.GetSpan(ctx).NewChild(operation)
		span.SetComponent(c.componentName)

		if !c.omitDbConnectionTags {
			if c.address != "" {
				span.SetPeerAddress(c.address)
			}

			if c.host != "" {
				span.SetPeerHost(c.host)
			}

			if c.port != "" {
				span.SetPeerPort(c.port)
			}

			if c.user != "" {
				span.SetDBUser(c.user)
			}

			if c.dbSystem != "" {
				span.SetDBSystem(c.dbSystem)
			}

			if c.dbName != "" {
				span.SetDBName(c.dbName)
			}

			span.SetDbConnectionString(c.rawString)
		}

		return &spanFinisherImpl{Logger: c.Logger, omitArgs: c.omitArgs, operation: operation, span: span, start: time.Now()}
	}
	return nullSpanFinisher{}
}

func (c childSpanFactoryImpl) NewChildSpanWithQuery(ctx context.Context, operation string, query string, args interface{}) spanFinisher {
	if !c.hasOpExcluded(operation) {
		finisher := c.NewChildSpan(ctx, operation)
		f, _ := finisher.(*spanFinisherImpl)
		f.span.SetDBStatement(query)
		if !c.omitArgs {
			f.span.SetDBStatementArgs(formatArgs(args))
		}

		f.query = &query
		f.args = args
		return f
	}
	return nullSpanFinisher{}
}

type nullSpanFinisher struct{}

func (f nullSpanFinisher) Finish(context.Context, error) {}

type nullTracer struct{}

func (nullTracer) GetSpan(context.Context) Span { return nullSpan{} }

type nullSpan struct{}

func (nullSpan) NewChild(string) Span         { return nullSpan{} }
func (nullSpan) SetLabel(string, string)      {}
func (nullSpan) SetComponent(string)          {}
func (nullSpan) SetDbConnectionString(string) {}
func (nullSpan) SetDBName(string)             {}
func (nullSpan) SetDBUser(string)             {}
func (nullSpan) SetDBSystem(string)           {}
func (nullSpan) SetDBStatement(string)        {}
func (nullSpan) SetDBStatementArgs(string)    {}
func (nullSpan) SetPeerAddress(string)        {}
func (nullSpan) SetPeerHost(string)           {}
func (nullSpan) SetPeerPort(string)           {}
func (nullSpan) Finish()                      {}
func (nullSpan) SetError(error)               {}
