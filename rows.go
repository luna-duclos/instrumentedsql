package instrumentedsql

import (
	"context"
	"database/sql/driver"
	"errors"
	"io"
	"time"
)

type wrappedRows struct {
	opts
	ctx    context.Context
	parent driver.Rows
}

func (r wrappedRows) Columns() []string {
	return r.parent.Columns()
}

func (r wrappedRows) Close() error {
	return r.parent.Close()
}

func (r wrappedRows) Next(dest []driver.Value) (err error) {
	if !r.hasOpExcluded(OpSQLRowsNext) {
		span := r.GetSpan(r.ctx).NewChild(OpSQLRowsNext)
		span.SetLabel("component", "database/sql")
		defer func() {
			if err != io.EOF {
				span.SetError(err)
			}
			span.Finish()
		}()
	}

	start := time.Now()
	defer func() {
		r.Log(r.ctx, OpSQLRowsNext, "err", err, "duration", time.Since(start))
	}()

	return r.parent.Next(dest)
}


// namedValueToValue is a helper function copied from the database/sql package
func namedValueToValue(named []driver.NamedValue) ([]driver.Value, error) {
	dargs := make([]driver.Value, len(named))
	for n, param := range named {
		if len(param.Name) > 0 {
			return nil, errors.New("sql: driver does not support the use of Named Parameters")
		}
		dargs[n] = param.Value
	}
	return dargs, nil
}
