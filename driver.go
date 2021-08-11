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
	childSpanFactory
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
	d := WrappedDriver{parent: driver}
	d.setDefaults()

	for _, opt := range opts {
		opt(&d.opts)
	}

	d.childSpanFactory = childSpanFactoryImpl{opts: d.opts}

	return d
}

// Open implements the database/sql/driver.Driver interface for WrappedDriver.
func (d WrappedDriver) Open(name string) (driver.Conn, error) {
	conn, err := d.parent.Open(name)
	if err != nil {
		return nil, err
	}

	return wrappedConn{opts: d.opts, childSpanFactory: d.childSpanFactory, parent: conn}, nil
}

func (d *WrappedDriver) setDefaults() {
	d.Logger = nullLogger{}
	d.Tracer = nullTracer{}
	d.omitArgs = true
	d.componentName = "database/sql"
	d.dbName = "unknown"
	d.dbUser = "unknown"
	d.dbSystem = "unknown"
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
}

func (c childSpanFactoryImpl) NewChildSpan(ctx context.Context, operation string) spanFinisher {
	if !c.hasOpExcluded(operation) {
		span := c.GetSpan(ctx).NewChild(operation)
		span.SetComponent(c.componentName)
		span.SetDBName(c.dbName)
		span.SetDBUser(c.dbUser)
		span.SetDBSystem(c.dbSystem)
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

func (nullSpan) NewChild(string) Span        { return nullSpan{} }
func (nullSpan) SetLabel(k, v string)        {}
func (nullSpan) SetComponent(v string)       {}
func (nullSpan) SetDBName(v string)          {}
func (nullSpan) SetDBUser(v string)          {}
func (nullSpan) SetDBSystem(v string)        {}
func (nullSpan) SetDBStatement(v string)     {}
func (nullSpan) SetDBStatementArgs(v string) {}
func (nullSpan) Finish()                     {}
func (nullSpan) SetError(err error)          {}
