package workflow

import (
	"context"
	"encoding/json"
	"testing"
)

func FuzzCompileDefinition(f *testing.F) {
	valid, err := json.Marshal(validDefinition())
	if err != nil {
		f.Fatal(err)
	}
	f.Add(valid)
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"schemaVersion":"1","command":"echo"}`))
	f.Add([]byte{0xff, 0x00, '{', '}'})

	f.Fuzz(func(t *testing.T, contents []byte) {
		if len(contents) > 2<<20 {
			t.Skip()
		}
		source := DefinitionSource{
			MediaType:  DefinitionMediaType,
			SchemaID:   DefinitionSchemaID,
			SourceID:   "fuzz-source",
			Version:    "1",
			Coordinate: SourceCoordinate{Path: "fuzz.json", StartLine: 1, StartColumn: 1, Unit: "utf8-byte"},
			Bytes:      append([]byte(nil), contents...),
		}
		graph, diagnostics := Compile(context.Background(), source, CompileLimits{})
		if graph.valid() == (len(diagnostics) != 0) {
			t.Fatalf("runnable graph/diagnostic invariant violated: valid=%t diagnostics=%d", graph.valid(), len(diagnostics))
		}
	})
}
