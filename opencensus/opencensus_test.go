package opencensus_test

import (
	"database/sql"

	"github.com/go-sql-driver/mysql"
	"github.com/luna-duclos/instrumentedsql"
	"github.com/luna-duclos/instrumentedsql/opencensus"
)

// WrapDriverOpencensus demonstrates how to call wrapDriver and register a new driver.
// This example uses MySQL and opencensus to illustrate this
func ExampleWrapDriver_opencensus() {
	sql.Register("instrumented-mysql", instrumentedsql.WrapDriver(mysql.MySQLDriver{}, instrumentedsql.WithTracer(opencensus.NewTracer(false))))
	db, err := sql.Open("instrumented-mysql", "connString")

	// Proceed to handle connection errors and use the database as usual
	_, _ = db, err
}
