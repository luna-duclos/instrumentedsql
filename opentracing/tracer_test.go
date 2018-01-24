package opentracing

import (
	"context"
	"fmt"
	"testing"

	opentracing "github.com/opentracing/opentracing-go"
)

func TestSpanWithParent(t *testing.T) {
	ctx := opentracing.ContextWithSpan(
		context.Background(),
		opentracing.GlobalTracer().StartSpan("some_span"),
	)

	tr := NewTracer(true)
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
	tr := NewTracer(true)
	span := tr.GetSpan(ctx)
	span.SetLabel("key", "value")

	child := span.NewChild("child")
	child.SetLabel("child_key", "child_value")
	child.SetError(fmt.Errorf("my error"))
	child.Finish()

	span.Finish()
}
