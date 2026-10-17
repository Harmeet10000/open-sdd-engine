package workspace

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// GenerationRecordPath is coordination metadata, not a domain revision.
const GenerationRecordPath WorkspacePath = ".sdd/.runtime/transactions/generation.json"

const generationSchema = "sdd.workspace-generation/1"

var ErrGenerationInvalid = errors.New("workspace generation record invalid")

// GenerationRecord is the checksummed coordination fact observed before and
// after a raw capture. Its zero value is invalid.
type GenerationRecord struct {
	stableGeneration  uint64
	activeTransaction string
}

// NewGenerationRecord constructs a generation fact. An empty transaction ID
// means the generation is stable for readers.
func NewGenerationRecord(stableGeneration uint64, activeTransaction string) GenerationRecord {
	return GenerationRecord{stableGeneration: stableGeneration, activeTransaction: activeTransaction}
}

// StableGeneration reports the last fully committed workspace generation.
func (r GenerationRecord) StableGeneration() uint64 { return r.stableGeneration }

// ActiveTransactionID reports the prepared/committing transaction, if any.
func (r GenerationRecord) ActiveTransactionID() string { return r.activeTransaction }

// MarshalJSON writes the internal record in one deterministic field order.
func (r GenerationRecord) MarshalJSON() ([]byte, error) {
	if r.activeTransaction != "" {
		if err := validateOpaqueID("active transaction", r.activeTransaction); err != nil {
			return nil, err
		}
	}
	return json.Marshal(generationWire{
		Schema:              generationSchema,
		StableGeneration:    r.stableGeneration,
		ActiveTransactionID: r.activeTransaction,
		Checksum:            generationChecksum(r.stableGeneration, r.activeTransaction),
	})
}

// ParseGenerationRecord decodes exactly one strict checksummed record.
func ParseGenerationRecord(data []byte) (GenerationRecord, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var wire generationWire
	if err := decoder.Decode(&wire); err != nil {
		return GenerationRecord{}, fmt.Errorf("%w: decode: %w", ErrGenerationInvalid, err)
	}
	if err := requireJSONEOF(decoder); err != nil {
		return GenerationRecord{}, err
	}
	if wire.Schema != generationSchema {
		return GenerationRecord{}, fmt.Errorf("%w: unsupported schema %q", ErrGenerationInvalid, wire.Schema)
	}
	if wire.ActiveTransactionID != "" {
		if err := validateOpaqueID("active transaction", wire.ActiveTransactionID); err != nil {
			return GenerationRecord{}, err
		}
	}
	want := generationChecksum(wire.StableGeneration, wire.ActiveTransactionID)
	if wire.Checksum != want {
		return GenerationRecord{}, fmt.Errorf("%w: checksum mismatch", ErrGenerationInvalid)
	}
	return NewGenerationRecord(wire.StableGeneration, wire.ActiveTransactionID), nil
}

type generationWire struct {
	Schema              string `json:"schema"`
	StableGeneration    uint64 `json:"stableGeneration"`
	ActiveTransactionID string `json:"activeTransactionId,omitempty"`
	Checksum            Digest `json:"checksum"`
}

func generationChecksum(stable uint64, active string) Digest {
	return digestBytes("sdd.workspace-generation/1", generationChecksumInput(stable, active))
}

func generationChecksumInput(stable uint64, active string) []byte {
	var data bytes.Buffer
	writeUint64(&data, stable)
	writeFramed(&data, active)
	return data.Bytes()
}

func validateOpaqueID(name, value string) error {
	if value == "" || value != string(bytes.TrimSpace([]byte(value))) {
		return fmt.Errorf("%w: %s ID is empty or padded", ErrGenerationInvalid, name)
	}
	for _, character := range value {
		if character < 0x20 || character == 0x7f || character == '/' || character == '\\' {
			return fmt.Errorf("%w: %s ID contains an unsafe character", ErrGenerationInvalid, name)
		}
	}
	return nil
}

func requireJSONEOF(decoder *json.Decoder) error {
	var extra any
	err := decoder.Decode(&extra)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("%w: trailing JSON: %w", ErrGenerationInvalid, err)
	}
	return fmt.Errorf("%w: multiple JSON values", ErrGenerationInvalid)
}
