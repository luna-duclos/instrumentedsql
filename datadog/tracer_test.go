package datadog

import (
	"database/sql"

	"github.com/go-sql-driver/mysql"
	"github.com/luna-duclos/instrumentedsql"
)

// WrapDriverDatadog demonstrates how to call wrapDriver and register a new driver.
// This example uses MySQL and datadog to illustrate this
func ExampleWrapDriver_datadog() {
	sql.Register("instrumented-mysql", instrumentedsql.WrapDriver(mysql.MySQLDriver{}, instrumentedsql.WithTracer(NewTracer())))
	db, err := sql.Open("instrumented-mysql", "connString")

	// Proceed to handle connection errors and use the database as usual
	_, _ = db, err
}
