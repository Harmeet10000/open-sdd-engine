package workspace

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
)

func TestGenerationRecordRoundTrip(t *testing.T) {
	t.Parallel()

	want := NewGenerationRecord(17, "tx-01JTEST")
	data, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseGenerationRecord(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.StableGeneration() != want.StableGeneration() || got.ActiveTransactionID() != want.ActiveTransactionID() {
		t.Fatalf("round trip = %#v, want %#v", got, want)
	}
}

func TestGenerationRecordRejectsChangedOrAmbiguousBytes(t *testing.T) {
	t.Parallel()

	valid, err := json.Marshal(NewGenerationRecord(4, ""))
	if err != nil {
		t.Fatal(err)
	}
	changed := bytes.Replace(valid, []byte(`"stableGeneration":4`), []byte(`"stableGeneration":5`), 1)

	tests := []struct {
		name string
		data []byte
	}{
		{name: "checksum mismatch", data: changed},
		{name: "unknown field", data: bytes.Replace(valid, []byte(`"checksum"`), []byte(`"unknown":true,"checksum"`), 1)},
		{name: "two values", data: append(append([]byte{}, valid...), valid...)},
		{name: "truncated", data: valid[:len(valid)-1]},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, parseErr := ParseGenerationRecord(test.data); !errors.Is(parseErr, ErrGenerationInvalid) {
				t.Fatalf("ParseGenerationRecord() error = %v", parseErr)
			}
		})
	}
}

func TestGenerationRecordRejectsUnsafeTransactionID(t *testing.T) {
	t.Parallel()

	if _, err := json.Marshal(NewGenerationRecord(1, "../tx")); !errors.Is(err, ErrGenerationInvalid) {
		t.Fatalf("MarshalJSON() error = %v", err)
	}
}
