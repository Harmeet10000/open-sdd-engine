package canonicaljson

import (
	"errors"
	"testing"
)

func TestMarshalTypedRecordIsDeterministic(t *testing.T) {
	t.Parallel()

	type child struct {
		Name string `json:"name"`
	}
	type record struct {
		Version uint64  `json:"version"`
		Items   []child `json:"items"`
	}
	value := record{Version: 2, Items: []child{{Name: "alpha"}, {Name: "beta"}}}
	first, err := Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) || string(first) != `{"version":2,"items":[{"name":"alpha"},{"name":"beta"}]}` {
		t.Fatalf("canonical bytes = %s and %s", first, second)
	}
}

func TestMarshalRejectsUnversionedMapAndFloatRules(t *testing.T) {
	t.Parallel()

	for _, value := range []any{map[string]string{"a": "b"}, 1.5} {
		if _, err := Marshal(value); !errors.Is(err, ErrUnsupportedValue) {
			t.Fatalf("Marshal(%T) error = %v", value, err)
		}
	}
}
