package instrumentedsql

import (
	"context"
	"database/sql/driver"
)

type wrappedTx struct {
	opts
	childSpanFactory
	ctx    context.Context
	parent driver.Tx
}

// Compile time validation that our types implement the expected interfaces
var (
	_ driver.Tx = wrappedTx{}
)

func (t wrappedTx) Commit() (err error) {
	span := t.NewChildSpan(t.ctx, OpSQLTxCommit)
	defer func() {
		span.Finish(t.ctx, err)
	}()

	return t.parent.Commit()
}

func (t wrappedTx) Rollback() (err error) {
	span := t.NewChildSpan(t.ctx, OpSQLTxRollback)
	defer func() {
		span.Finish(t.ctx, err)
	}()

	return t.parent.Rollback()
}
