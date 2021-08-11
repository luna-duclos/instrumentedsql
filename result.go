package instrumentedsql

import (
	"context"
	"database/sql/driver"
)

type wrappedResult struct {
	opts
	childSpanFactory
	ctx    context.Context
	parent driver.Result
}

func (r wrappedResult) LastInsertId() (id int64, err error) {
	span := r.NewChildSpan(r.ctx, OpSQLResLastInsertID)
	defer func() {
		span.Finish(r.ctx, err)
	}()

	return r.parent.LastInsertId()
}

func (r wrappedResult) RowsAffected() (num int64, err error) {
	span := r.NewChildSpan(r.ctx, OpSQLResRowsAffected)
	defer func() {
		span.Finish(r.ctx, err)
	}()

	return r.parent.RowsAffected()
}
