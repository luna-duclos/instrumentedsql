package instrumentedsql_test

import (
	"context"
	"database/sql"
	"log"

	"github.com/mattn/go-sqlite3"

	"github.com/luna-duclos/instrumentedsql"
)

// WrapDriverJustLogging demonstrates how to call wrapDriver and register a new driver.
// This example uses sqlite, and does not trace, but merely logs all calls
func ExampleWrapDriver_justLogging() {
	logger := instrumentedsql.LoggerFunc(func(ctx context.Context, msg string, keyvals ...interface{}) {
		log.Printf("%s %v", msg, keyvals)
	})

	sql.Register("instrumented-sqlite", instrumentedsql.WrapDriver(&sqlite3.SQLiteDriver{}, instrumentedsql.WithLogger(logger)))
	db, err := sql.Open("instrumented-sqlite", "connString")

	// Proceed to handle connection errors and use the database as usual
	_, _ = db, err
}
