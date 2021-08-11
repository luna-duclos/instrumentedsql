package instrumentedsql

import (
	"context"
	"database/sql/driver"
	"time"
)

type wrappedConn struct {
	opts
	childSpanFactory
	parent driver.Conn
}

// Compile time validation that our types implement the expected interfaces
var (
	_ driver.Conn               = wrappedConn{}
	_ driver.ConnBeginTx        = wrappedConn{}
	_ driver.ConnPrepareContext = wrappedConn{}
	_ driver.Execer             = wrappedConn{}
	_ driver.ExecerContext      = wrappedConn{}
	_ driver.Pinger             = wrappedConn{}
	_ driver.Queryer            = wrappedConn{}
	_ driver.QueryerContext     = wrappedConn{}
)

func (c wrappedConn) Prepare(query string) (driver.Stmt, error) {
	parent, err := c.parent.Prepare(query)
	if err != nil {
		return nil, err
	}

	return wrappedStmt{opts: c.opts, childSpanFactory: c.childSpanFactory, ctx: context.TODO(), query: query, parent: parent}, nil
}

func (c wrappedConn) Close() error {
	return c.parent.Close()
}

func (c wrappedConn) Begin() (driver.Tx, error) {
	tx, err := c.parent.Begin()
	if err != nil {
		return nil, err
	}

	return wrappedTx{opts: c.opts, childSpanFactory: c.childSpanFactory, ctx: context.TODO(), parent: tx}, nil
}

func (c wrappedConn) BeginTx(ctx context.Context, opts driver.TxOptions) (tx driver.Tx, err error) {
	span := c.NewChildSpan(ctx, OpSQLTxBegin)
	defer func() {
		span.Finish(ctx, err)
	}()

	if connBeginTx, ok := c.parent.(driver.ConnBeginTx); ok {
		tx, err = connBeginTx.BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}

		return wrappedTx{opts: c.opts, childSpanFactory: c.childSpanFactory, ctx: ctx, parent: tx}, nil
	}

	tx, err = c.parent.Begin()
	if err != nil {
		return nil, err
	}

	return wrappedTx{opts: c.opts, childSpanFactory: c.childSpanFactory, ctx: ctx, parent: tx}, nil
}

func (c wrappedConn) PrepareContext(ctx context.Context, query string) (stmt driver.Stmt, err error) {
	span := c.NewChildSpanWithQuery(ctx, OpSQLPrepare, query, nil)
	defer func() {
		span.Finish(ctx, err)
	}()

	if connPrepareCtx, ok := c.parent.(driver.ConnPrepareContext); ok {
		stmt, err := connPrepareCtx.PrepareContext(ctx, query)
		if err != nil {
			return nil, err
		}

		return wrappedStmt{opts: c.opts, childSpanFactory: c.childSpanFactory, ctx: ctx, query: query, parent: stmt}, nil
	}

	return c.Prepare(query)
}

func (c wrappedConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	if execer, ok := c.parent.(driver.Execer); ok {
		res, err := execer.Exec(query, args)
		if err != nil {
			return nil, err
		}

		return wrappedResult{opts: c.opts, childSpanFactory: c.childSpanFactory, ctx: context.TODO(), parent: res}, nil
	}

	return nil, driver.ErrSkip
}

func (c wrappedConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (r driver.Result, err error) {
	span := c.NewChildSpanWithQuery(ctx, OpSQLConnExec, query, args)
	defer func() {
		span.Finish(ctx, err)
	}()

	if execContext, ok := c.parent.(driver.ExecerContext); ok {
		res, err := execContext.ExecContext(ctx, query, args)
		if err != nil {
			return nil, err
		}

		return wrappedResult{opts: c.opts, childSpanFactory: c.childSpanFactory, ctx: ctx, parent: res}, nil
	}

	// Fallback implementation
	dargs, err := namedValueToValue(args)
	if err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return c.Exec(query, dargs)
	}
}

func (c wrappedConn) Ping(ctx context.Context) (err error) {
	if pinger, ok := c.parent.(driver.Pinger); ok {
		span := c.NewChildSpan(ctx, OpSQLPing)
		defer func() {
			span.Finish(ctx, err)
		}()

		return pinger.Ping(ctx)
	}

	c.Log(ctx, OpSQLDummyPing, "duration", time.Duration(0))

	return nil
}

func (c wrappedConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	if queryer, ok := c.parent.(driver.Queryer); ok {
		rows, err := queryer.Query(query, args)
		if err != nil {
			return nil, err
		}

		return wrappedRows{opts: c.opts, childSpanFactory: c.childSpanFactory, parent: rows}, nil
	}

	return nil, driver.ErrSkip
}

func (c wrappedConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (rows driver.Rows, err error) {
	// Quick skip path: If the wrapped connection implements neither QueryerContext nor Queryer, we have absolutely nothing to do
	_, hasQueryerContext := c.parent.(driver.QueryerContext)
	_, hasQueryer := c.parent.(driver.Queryer)
	if !hasQueryerContext && !hasQueryer {
		return nil, driver.ErrSkip
	}

	span := c.NewChildSpanWithQuery(ctx, OpSQLConnQuery, query, args)
	defer func() {
		span.Finish(ctx, err)
	}()

	if queryerContext, ok := c.parent.(driver.QueryerContext); ok {
		rows, err := queryerContext.QueryContext(ctx, query, args)
		if err != nil {
			return nil, err
		}

		return wrappedRows{opts: c.opts, childSpanFactory: c.childSpanFactory, ctx: ctx, parent: rows}, nil
	}

	dargs, err := namedValueToValue(args)
	if err != nil {
		return nil, err
	}

	select {
	default:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return c.Query(query, dargs)
}
