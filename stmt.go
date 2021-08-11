package instrumentedsql

import (
	"context"
	"database/sql/driver"
)

type wrappedStmt struct {
	opts
	childSpanFactory
	ctx    context.Context
	query  string
	parent driver.Stmt
}

// Compile time validation that our types implement the expected interfaces
var (
	_ driver.Stmt             = wrappedStmt{}
	_ driver.StmtExecContext  = wrappedStmt{}
	_ driver.StmtQueryContext = wrappedStmt{}
)

func (s wrappedStmt) Close() (err error) {
	span := s.NewChildSpan(s.ctx, OpSQLStmtClose)
	defer func() {
		span.Finish(s.ctx, err)
	}()

	return s.parent.Close()
}

func (s wrappedStmt) NumInput() int {
	return s.parent.NumInput()
}

func (s wrappedStmt) Exec(args []driver.Value) (res driver.Result, err error) {
	span := s.NewChildSpanWithQuery(s.ctx, OpSQLStmtExec, s.query, args)
	defer func() {
		span.Finish(s.ctx, err)
	}()

	res, err = s.parent.Exec(args)
	if err != nil {
		return nil, err
	}

	return wrappedResult{opts: s.opts, childSpanFactory: s.childSpanFactory, ctx: s.ctx, parent: res}, nil
}

func (s wrappedStmt) Query(args []driver.Value) (rows driver.Rows, err error) {
	span := s.NewChildSpanWithQuery(s.ctx, OpSQLStmtQuery, s.query, args)
	defer func() {
		span.Finish(s.ctx, err)
	}()

	rows, err = s.parent.Query(args)
	if err != nil {
		return nil, err
	}

	return wrappedRows{opts: s.opts, childSpanFactory: s.childSpanFactory, ctx: s.ctx, parent: rows}, nil
}

func (s wrappedStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (res driver.Result, err error) {
	span := s.NewChildSpanWithQuery(ctx, OpSQLStmtExec, s.query, args)
	defer func() {
		span.Finish(ctx, err)
	}()

	if stmtExecContext, ok := s.parent.(driver.StmtExecContext); ok {
		res, err := stmtExecContext.ExecContext(ctx, args)
		if err != nil {
			return nil, err
		}

		return wrappedResult{opts: s.opts, childSpanFactory: s.childSpanFactory, ctx: ctx, parent: res}, nil
	}

	// Fallback implementation
	dargs, err := namedValueToValue(args)
	if err != nil {
		return nil, err
	}

	select {
	default:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	res, err = s.parent.Exec(dargs)
	if err != nil {
		return nil, err
	}

	return wrappedResult{opts: s.opts, childSpanFactory: s.childSpanFactory, ctx: ctx, parent: res}, nil
}

func (s wrappedStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (rows driver.Rows, err error) {
	span := s.NewChildSpanWithQuery(ctx, OpSQLStmtQuery, s.query, args)
	defer func() {
		span.Finish(ctx, err)
	}()

	if stmtQueryContext, ok := s.parent.(driver.StmtQueryContext); ok {
		rows, err := stmtQueryContext.QueryContext(ctx, args)
		if err != nil {
			return nil, err
		}

		return wrappedRows{opts: s.opts, childSpanFactory: s.childSpanFactory, ctx: ctx, parent: rows}, nil
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

	rows, err = s.parent.Query(dargs)
	if err != nil {
		return nil, err
	}

	return wrappedRows{opts: s.opts, childSpanFactory: s.childSpanFactory, ctx: ctx, parent: rows}, nil
}
