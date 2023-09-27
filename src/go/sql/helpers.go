package sql

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

func In(query string, args ...any) (string, []any, error) {
	return sqlx.In(query, args...)
}

type NullTime = sql.NullTime

// Repeat the placeholders n times, separated by comas ','.
func Repeat(placeholder string, n int) string {
	out := strings.Repeat(placeholder+", ", n)
	return out[:len(out)-2]
}

type Where []string

func (w Where) String() string {
	if len(w) == 0 {
		return ""
	}
	return `WHERE ` + strings.Join(w, " AND ")
}

// FormatQuery converts the query and arguments into a single string, ready for
// usage in SQL client. Only '?' placeholders are replaced.
func FormatQuery(in string, args ...interface{}) string {
	quote := func(raw []byte) []byte {
		return append([]byte("'"), append(raw, '\'')...)
	}

	var index int
	out := []byte(in)
	for i := 0; i < len(out); i++ {
		c := out[i]
		if c != '?' {
			continue
		}

		if index >= len(args) {
			panic(fmt.Errorf("index [%d] of '?' is greater than query arguments number = %d -- double check usage of variadic to args...", index, len(args)))
		}

		var v interface{} = args[index]
		var raw []byte
		if dv, ok := v.(driver.Valuer); ok {
			var err error
			v, err = callValuerValue(dv)
			if err != nil {
				panic(fmt.Errorf("driver.Value() error: %s", err))
			}
		}

		switch t := v.(type) {
		case int64:
			raw = []byte(strconv.FormatInt(t, 10))
		case float64:
			raw = []byte(strconv.FormatFloat(t, 'f', -1, 64))
		case bool:
			raw = []byte(strconv.FormatBool(t))
		case []byte:
			raw = quote(t)
		case json.RawMessage:
			raw = quote([]byte(t))
		case string:
			raw = quote([]byte(t))
		case time.Time:
			raw = quote(pq.FormatTimestamp(t))
		case nil:
			raw = []byte("NULL")
		default:
			raw = []byte(fmt.Sprintf("%v", t))
		}
		out = append(out[:i], append(raw, out[i+1:]...)...)

		// move the cursor to the end of the raw
		// decrement i to skip '?'
		i += len(raw)
		i--
		index++
	}

	return strings.Join(strings.Fields(string(out)), " ")
}

var valuerReflectType = reflect.TypeOf((*driver.Valuer)(nil)).Elem()

// callValuerValue returns vr.Value(), with one exception: If vr.Value is an
// auto-generated method on a pointer type and the pointer is nil, it would
// panic at runtime in the panicwrap method. Treat it like nil instead. Issue
// 8415.
//
// This function is copy-pasted from database/sql/driver package.
func callValuerValue(vr driver.Valuer) (v driver.Value, err error) {
	if rv := reflect.ValueOf(vr); rv.Kind() == reflect.Ptr &&
		rv.IsNil() &&
		rv.Type().Elem().Implements(valuerReflectType) {
		return nil, nil
	}
	return vr.Value()
}

func ScanJSON(src, dst any) error {
	switch src := src.(type) {
	case []byte:
		return json.Unmarshal(src, dst)
	case string:
		return json.Unmarshal([]byte(src), dst)
	case nil:
		return nil
	default:
		return fmt.Errorf(`invalid scan pair %T => %T`, src, dst)
	}
}

func ValueJSON(src any) (driver.Value, error) {
	return json.Marshal(src)
}

func ScanText(src any, dst encoding.TextUnmarshaler) error {
	switch src := src.(type) {
	case string:
		return dst.UnmarshalText([]byte(src))
	case []byte:
		return dst.UnmarshalText(src)
	default:
		return fmt.Errorf(`invalid scan pair %T => %T`, src, dst)
	}
}

func ValueText(src encoding.TextMarshaler) (driver.Value, error) {
	return src.MarshalText()
}
