// Package canonicaljson owns the deterministic JSON encoding used by
// foundation-compatible typed records. Schemas must use structs, slices, and
// integer/string/bool values; semantic maps and floating-point values require
// a separately versioned canonicalization rule.
package canonicaljson

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"
)

var ErrUnsupportedValue = errors.New("canonical JSON value unsupported")

// Marshal encodes one typed value after rejecting representation classes that
// do not have a deliberately accepted canonical rule. encoding/json preserves
// struct field order, slice order, integer spelling, and sorts string map keys;
// maps are rejected here so schema authors cannot make semantic order implicit.
func Marshal(value any) ([]byte, error) {
	if err := validate(reflect.ValueOf(value)); err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("canonical JSON encode: %w", err)
	}
	return encoded, nil
}

func validate(value reflect.Value) error {
	if !value.IsValid() {
		return nil
	}
	if value.Kind() == reflect.Interface {
		if value.IsNil() {
			return nil
		}
		return validate(value.Elem())
	}
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil
		}
		return validate(value.Elem())
	}
	if value.Type() == reflect.TypeFor[json.Number]() {
		return fmt.Errorf("%w: json.Number", ErrUnsupportedValue)
	}
	if value.Type() == reflect.TypeFor[time.Time]() {
		return nil
	}
	switch value.Kind() {
	case reflect.Bool, reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return nil
	case reflect.Struct:
		for index := 0; index < value.NumField(); index++ {
			field := value.Type().Field(index)
			if !field.IsExported() || field.Tag.Get("json") == "-" {
				continue
			}
			if err := validate(value.Field(index)); err != nil {
				return err
			}
		}
		return nil
	case reflect.Array, reflect.Slice:
		for index := 0; index < value.Len(); index++ {
			if err := validate(value.Index(index)); err != nil {
				return err
			}
		}
		return nil
	case reflect.Map, reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
		reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return fmt.Errorf("%w: %s", ErrUnsupportedValue, value.Kind())
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedValue, value.Kind())
	}
}
