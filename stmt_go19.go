// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.golang file.

// +build go1.9

package instrumentedsql

import (
	"database/sql/driver"
	"fmt"
	"reflect"
)

var _ driver.NamedValueChecker = wrappedStmt{}

func (s wrappedStmt) CheckNamedValue(v *driver.NamedValue) error {
	if checker, ok := s.parent.(driver.NamedValueChecker); ok {
		err := checker.CheckNamedValue(v)
		if err != driver.ErrSkip {
			return err
		}
	}

	if converter, ok := s.parent.(driver.ColumnConverter); ok {
		cc := ccChecker{
			cci:  converter,
			want: s.NumInput(),
		}
		return cc.CheckNamedValue(v)
	}

	return driver.ErrSkip
}

// ccChecker wraps the driver.ColumnConverter and allows it to be used
// as if it were a NamedValueChecker. If the driver ColumnConverter
// is not present then the NamedValueChecker will return driver.ErrSkip.
type ccChecker struct {
	cci  driver.ColumnConverter
	want int
}

func (c ccChecker) CheckNamedValue(nv *driver.NamedValue) error {
	if c.cci == nil {
		return driver.ErrSkip
	}
	// The column converter shouldn't be called on any index
	// it isn't expecting. The final error will be thrown
	// in the argument converter loop.
	index := nv.Ordinal - 1
	if c.want <= index {
		return nil
	}

	// First, see if the value itself knows how to convert
	// itself to a driver type. For example, a NullString
	// struct changing into a string or nil.
	if vr, ok := nv.Value.(driver.Valuer); ok {
		sv, err := callValuerValue(vr)
		if err != nil {
			return err
		}
		if !driver.IsValue(sv) {
			return fmt.Errorf("non-subset type %T returned from Value", sv)
		}
		nv.Value = sv
	}

	// Second, ask the column to sanity check itself. For
	// example, drivers might use this to make sure that
	// an int64 values being inserted into a 16-bit
	// integer field is in range (before getting
	// truncated), or that a nil can't go into a NOT NULL
	// column before going across the network to get the
	// same error.
	var err error
	arg := nv.Value
	nv.Value, err = c.cci.ColumnConverter(index).ConvertValue(arg)
	if err != nil {
		return err
	}
	if !driver.IsValue(nv.Value) {
		return fmt.Errorf("driver ColumnConverter error converted %T to unsupported type %T", arg, nv.Value)
	}
	return nil
}

var valuerReflectType = reflect.TypeOf((*driver.Valuer)(nil)).Elem()

// callValuerValue returns vr.Value(), with one exception:
// If vr.Value is an auto-generated method on a pointer type and the
// pointer is nil, it would panic at runtime in the panicwrap
// method. Treat it like nil instead.
// Issue 8415.
//
// This is so people can implement driver.Value on value types and
// still use nil pointers to those types to mean nil/NULL, just like
// string/*string.
//
// This function is mirrored in the database/sql/driver package.
func callValuerValue(vr driver.Valuer) (v driver.Value, err error) {
	if rv := reflect.ValueOf(vr); rv.Kind() == reflect.Ptr &&
		rv.IsNil() &&
		rv.Type().Elem().Implements(valuerReflectType) {
		return nil, nil
	}
	return vr.Value()
}
