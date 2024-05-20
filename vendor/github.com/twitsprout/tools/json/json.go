package json

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/twitsprout/tools/buffer"
)

// Error represents an error due to JSON unmarshalling.
type Error struct {
	msg string
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.msg
}

// Encode writes the raw JSON of the provided value to the io.Writer using the
// given indent.
func Encode(w io.Writer, v interface{}, indent string) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if indent != "" {
		enc.SetIndent("", indent)
	}
	return enc.Encode(v)
}

// Decode read the raw JSON from the provided io.Reader into the value 'v'.
func Decode(r io.Reader, v interface{}) error {
	buf := buffer.Get()
	defer buffer.Put(buf)
	if _, err := buf.ReadFrom(r); err != nil {
		return err
	}
	return Unmarshal(buf.Bytes(), v)
}

// Unmarshal reads the raw JSON from 'b' into the value 'v'.
func Unmarshal(b []byte, v interface{}) error {
	err := json.Unmarshal(b, v)
	if err == nil {
		return nil
	}
	msg := errorMessage(err)
	if !strings.HasPrefix(msg, "json: ") {
		msg = "json: " + msg
	}
	return &Error{msg: fmt.Sprintf("%s: '%s'", msg, b)}
}

func errorMessage(err error) string {
	if ute, ok := err.(*json.UnmarshalTypeError); ok {
		return fromUnmarshalTypeError(ute)
	}
	return err.Error()
}

func fromUnmarshalTypeError(err *json.UnmarshalTypeError) string {
	var buf strings.Builder
	buf.WriteString("json: unexpected value at character ")
	buf.WriteString(strconv.FormatInt(err.Offset, 10))
	if err.Field != "" {
		buf.WriteString(" (field '")
		buf.WriteString(err.Field)
		buf.WriteString("')")
	}
	buf.WriteString(": received '")
	buf.WriteString(err.Value)
	buf.WriteByte('\'')
	exp := jsonTypeFromGo(err.Type, true)
	if exp != "" {
		buf.WriteString(", expecting '")
		buf.WriteString(exp)
		buf.WriteByte('\'')
	}
	return buf.String()
}

func jsonTypeFromGo(t reflect.Type, recursive bool) string {
	//nolint:exhaustive // Use default to catch all other cases.
	switch t.Kind() {
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Array, reflect.Slice:
		if recursive {
			elem := jsonTypeFromGo(t.Elem(), false)
			if elem != "" {
				return "array of " + elem + "s"
			}
		}
		return "array"
	case reflect.Map, reflect.Struct:
		return "object"
	case reflect.String:
		return "string"
	default:
		return t.Name()
	}
}
