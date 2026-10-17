package workflow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

type rejectingSchemaLoader struct{}

func (rejectingSchemaLoader) Load(url string) (any, error) {
	return nil, fmt.Errorf("offline workflow schema registry does not contain %q", url)
}

func TestDefinitionSchemaMatchesDecoderContract(t *testing.T) {
	t.Parallel()
	schema := compileDefinitionSchema(t)
	validBytes, err := json.Marshal(validDefinition())
	if err != nil {
		t.Fatal(err)
	}
	valid := decodeSchemaDocument(t, validBytes)
	if validateErr := schema.Validate(valid); validateErr != nil {
		t.Fatalf("valid definition rejected: %v", validateErr)
	}

	var unknown map[string]any
	if unmarshalErr := json.Unmarshal(validBytes, &unknown); unmarshalErr != nil {
		t.Fatal(unmarshalErr)
	}
	unknown["command"] = "rm"
	unknownBytes, err := json.Marshal(unknown)
	if err != nil {
		t.Fatal(err)
	}
	if validateErr := schema.Validate(decodeSchemaDocument(t, unknownBytes)); validateErr == nil {
		t.Fatal("schema accepted unknown executable-looking field")
	}

	missingCardinality := validDefinition()
	missingCardinality.Nodes[1].Inputs[0].Cardinality = ""
	invalidBytes, err := json.Marshal(missingCardinality)
	if err != nil {
		t.Fatal(err)
	}
	if validateErr := schema.Validate(decodeSchemaDocument(t, invalidBytes)); validateErr == nil {
		t.Fatal("schema accepted missing required cardinality")
	}
}

func compileDefinitionSchema(t *testing.T) *jsonschema.Schema {
	t.Helper()
	path := filepath.Join("..", "..", "workflow", "schema", "definition.schema.json")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	document := decodeSchemaDocument(t, contents)
	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft7)
	compiler.UseLoader(rejectingSchemaLoader{})
	if addErr := compiler.AddResource(DefinitionSchemaID, document); addErr != nil {
		t.Fatal(addErr)
	}
	schema, err := compiler.Compile(DefinitionSchemaID)
	if err != nil {
		t.Fatalf("compile workflow definition schema: %v", err)
	}
	return schema
}

func decodeSchemaDocument(t *testing.T, contents []byte) any {
	t.Helper()
	document, err := jsonschema.UnmarshalJSON(bytes.NewReader(contents))
	if err != nil {
		t.Fatalf("decode schema JSON: %v", err)
	}
	return document
}
