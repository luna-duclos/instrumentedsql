// +build !go1.9

package instrumentedsql

import "database/sql/driver"

var _ driver.ColumnConverter = wrappedStmt{}

func (s wrappedStmt) ColumnConverter(idx int) driver.ValueConverter {
	if converter, ok := s.parent.(driver.ColumnConverter); ok {
		return converter.ColumnConverter(idx)
	}

	return driver.DefaultParameterConverter
}
