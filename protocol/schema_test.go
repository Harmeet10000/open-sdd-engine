package protocol

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const foundationSchemaID = "https://sdd.dev/schema/foundation-1.0.json"

// schemaFixtures contains the published contract and executable conformance
// examples. Validation never opens the network or depends on a working tree.
//
//go:embed schema/*.json testdata/outcome/*/*.json
var schemaFixtures embed.FS

type offlineLoader struct{}

func (offlineLoader) Load(url string) (any, error) {
	return nil, fmt.Errorf("offline schema registry does not contain %q", url)
}

func TestFoundationSchemaValidFixtures(t *testing.T) {
	schema := compileFoundationSchema(t, "")
	paths, err := fs.Glob(schemaFixtures, "testdata/outcome/valid/*.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no valid protocol fixtures found")
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			document := decodeSchemaJSON(t, path)
			if err := schema.Validate(document); err != nil {
				t.Fatalf("valid fixture rejected: %v", err)
			}
		})
	}
}

func TestFoundationSchemaRejectsInvalidFixtures(t *testing.T) {
	schema := compileFoundationSchema(t, "")
	paths, err := fs.Glob(schemaFixtures, "testdata/outcome/invalid/*.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no invalid protocol fixtures found")
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			document := decodeSchemaJSON(t, path)
			if err := schema.Validate(document); err == nil {
				t.Fatal("invalid fixture accepted")
			}
		})
	}
}

func TestFoundationCommandDiscriminatorIsClosed(t *testing.T) {
	schema := compileFoundationSchema(t, "#/definitions/Command")

	valid := decodeJSONBytes(t, []byte(`{"type":"status","workId":"work-1"}`))
	if err := schema.Validate(valid); err != nil {
		t.Fatalf("valid status command rejected: %v", err)
	}

	unknown := decodeJSONBytes(t, []byte(`{"type":"finish"}`))
	if err := schema.Validate(unknown); err == nil {
		t.Fatal("unpublished command accepted by the closed 1.0 discriminator")
	}
}

func TestStartDryRunFixtureContentDigest(t *testing.T) {
	contents, err := schemaFixtures.ReadFile("testdata/outcome/valid/start-dry-run.json")
	if err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		Result struct {
			Plan struct {
				Operations []struct {
					Content       string `json:"content"`
					ContentDigest string `json:"contentDigest"`
				} `json:"operations"`
			} `json:"plan"`
		} `json:"result"`
	}
	if err := json.Unmarshal(contents, &envelope); err != nil {
		t.Fatal(err)
	}
	if len(envelope.Result.Plan.Operations) != 1 {
		t.Fatalf("expected one fixture operation, got %d", len(envelope.Result.Plan.Operations))
	}
	operation := envelope.Result.Plan.Operations[0]
	want := fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(operation.Content)))
	if operation.ContentDigest != want {
		t.Fatalf("content digest = %q, want %q", operation.ContentDigest, want)
	}
}

func TestOfflineResolverRejectsUnregisteredReference(t *testing.T) {
	compiler := newOfflineCompiler()
	broken := decodeJSONBytes(t, []byte(`{
		"$schema":"http://json-schema.org/draft-07/schema#",
		"$id":"https://sdd.dev/schema/broken.json",
		"$ref":"https://example.invalid/unregistered.json"
	}`))
	if err := compiler.AddResource("https://sdd.dev/schema/broken.json", broken); err != nil {
		t.Fatal(err)
	}
	if _, err := compiler.Compile("https://sdd.dev/schema/broken.json"); err == nil {
		t.Fatal("schema with an unregistered reference compiled")
	}
}

func compileFoundationSchema(t *testing.T, fragment string) *jsonschema.Schema {
	t.Helper()
	compiler := newOfflineCompiler()
	document := decodeSchemaJSON(t, "schema/foundation-1.0.json")
	if err := compiler.AddResource(foundationSchemaID, document); err != nil {
		t.Fatal(err)
	}
	schema, err := compiler.Compile(foundationSchemaID + fragment)
	if err != nil {
		t.Fatalf("compile foundation schema: %v", err)
	}
	return schema
}

func newOfflineCompiler() *jsonschema.Compiler {
	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft7)
	compiler.UseLoader(offlineLoader{})
	return compiler
}

func decodeSchemaJSON(t *testing.T, path string) any {
	t.Helper()
	contents, err := schemaFixtures.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return decodeJSONBytes(t, contents)
}

func decodeJSONBytes(t *testing.T, contents []byte) any {
	t.Helper()
	document, err := jsonschema.UnmarshalJSON(bytes.NewReader(contents))
	if err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
	return document
}
