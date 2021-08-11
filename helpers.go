package instrumentedsql

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

func formatArgs(args interface{}) string {
	argsVal := reflect.ValueOf(args)
	if argsVal.Kind() != reflect.Slice {
		return "<unknown>"
	}

	strArgs := make([]string, 0, argsVal.Len())
	for i := 0; i < argsVal.Len(); i++ {
		strArgs = append(strArgs, formatArg(argsVal.Index(i).Interface()))
	}

	return fmt.Sprintf("{%s}", strings.Join(strArgs, ", "))
}

func formatArg(arg interface{}) string {
	strArg := ""
	switch arg := arg.(type) {
	case []uint8:
		strArg = fmt.Sprintf("[%T len:%d]", arg, len(arg))
	case string:
		strArg = fmt.Sprintf("[%T %q]", arg, arg)
	case driver.NamedValue:
		if arg.Name != "" {
			strArg = fmt.Sprintf("[%T %s=%v]", arg.Value, arg.Name, formatArg(arg.Value))
		} else {
			strArg = formatArg(arg.Value)
		}
	default:
		strArg = fmt.Sprintf("[%T %v]", arg, arg)
	}

	return strArg
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
