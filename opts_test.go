package instrumentedsql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithLogger(t *testing.T) {
	options := &opts{}
	logger := nullLogger{}
	WithLogger(logger)(options)
	assert.Equal(t, logger, options.Logger)
}

func TestWithOpsExcluded(t *testing.T) {
	cases := [][]string{
		{},
		{OpSQLConnExec},
		{OpSQLConnExec, OpSQLConnQuery},
		{OpSQLConnExec, OpSQLConnQuery, OpSQLRowsNext},
	}
	for _, ops := range cases {
		options := &opts{}
		WithOpsExcluded(ops...)(options)
		assert.Len(t, options.OpsExcluded, len(ops))
		for _, op := range ops {
			assert.True(t, options.hasOpExcluded(op))
		}
	}
}

func TestWithTracer(t *testing.T) {
	options := &opts{}
	tracer := nullTracer{}
	WithTracer(tracer)(options)
	assert.Equal(t, tracer, options.Tracer)
}

func TestWithOmitArgs(t *testing.T) {
	options := &opts{}
	assert.False(t, options.OmitArgs)
	WithOmitArgs()(options)
	assert.True(t, options.OmitArgs)
}

func TestIncludeArgs(t *testing.T) {
	options := &opts{OmitArgs: true}
	assert.True(t, options.OmitArgs)
	WithIncludeArgs()(options)
	assert.False(t, options.OmitArgs)
}
