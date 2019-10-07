package opentracing_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/luna-duclos/instrumentedsql"
	"github.com/luna-duclos/instrumentedsql/opentracing"
	opentracinggo "github.com/opentracing/opentracing-go"
)

// WrapDriverOpentracing demonstrates how to call wrapDriver and register a new driver.
// This example uses MySQL and opentracing to illustrate this
func ExampleWrapDriver_opentracing() {
	sql.Register("instrumented-mysql", instrumentedsql.WrapDriver(mysql.MySQLDriver{}, instrumentedsql.WithTracer(opentracing.NewTracer(false))))
	db, err := sql.Open("instrumented-mysql", "connString")

	// Proceed to handle connection errors and use the database as usual
	_, _ = db, err
}

func TestSpanWithParent(t *testing.T) {
	ctx := opentracinggo.ContextWithSpan(
		context.Background(),
		opentracinggo.GlobalTracer().StartSpan("some_span"),
	)

	tr := opentracing.NewTracer(true)
	span := tr.GetSpan(ctx)
	span.SetLabel("key", "value")

	child := span.NewChild("child")
	child.SetLabel("child_key", "child_value")
	child.SetError(fmt.Errorf("my error"))
	child.Finish()

	span.Finish()
}

func TestSpanWithoutParent(t *testing.T) {
	ctx := context.Background() // Background has no span
	tr := opentracing.NewTracer(true)
	span := tr.GetSpan(ctx)
	span.SetLabel("key", "value")

	child := span.NewChild("child")
	child.SetLabel("child_key", "child_value")
	child.SetError(fmt.Errorf("my error"))
	child.Finish()

	span.Finish()
}
