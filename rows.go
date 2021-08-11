package instrumentedsql

import (
	"context"
	"database/sql/driver"
	"io"
)

// Compile time validation that our types implement the expected interfaces
var (
	_ driver.Rows                           = wrappedRows{}
	_ driver.RowsColumnTypeDatabaseTypeName // TODO
	_ driver.RowsColumnTypeLength           // TODO
	_ driver.RowsColumnTypeNullable         // TODO
	_ driver.RowsColumnTypePrecisionScale   // TODO
	_ driver.RowsColumnTypeScanType         // TODO
	_ driver.RowsNextResultSet              // TODO
)

type wrappedRows struct {
	opts
	childSpanFactory
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
	span := r.NewChildSpan(r.ctx, OpSQLRowsNext)
	defer func() {
		if err != io.EOF {
			span.Finish(r.ctx, err)
		} else {
			span.Finish(r.ctx, nil)
		}
	}()

	return r.parent.Next(dest)
}
