package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sdd.dev/sdd/internal/cli"
)

// legacyFixture is one frozen TypeScript CLI observation. Only version and
// status behaviors that actually exist in the legacy CLI may appear here;
// doctor and start --dry-run are new-contract fixtures elsewhere, never
// fabricated legacy parity.
type legacyFixture struct {
	Name                   string         `json:"name"`
	Provenance             map[string]any `json:"provenance"`
	Command                []string       `json:"command"`
	Stdin                  string         `json:"stdin"`
	Stdout                 string         `json:"stdout"`
	Stderr                 string         `json:"stderr"`
	ExitCode               int            `json:"exitCode"`
	SchemaVersion          string         `json:"schemaVersion"`
	IntentionalDifferences []string       `json:"intentionalDifferences"`
}

func loadLegacyFixtures(t *testing.T) map[string]legacyFixture {
	t.Helper()
	entries, err := os.ReadDir("testdata/legacy-typescript")
	if err != nil {
		t.Fatal(err)
	}
	fixtures := map[string]legacyFixture{}
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".json") {
			t.Fatalf("non-fixture file in legacy corpus: %s", name)
		}
		if !strings.HasPrefix(name, "version") && !strings.HasPrefix(name, "status") {
			t.Fatalf("fixture %s claims parity beyond version/status", name)
		}
		content, err := os.ReadFile(filepath.Join("testdata/legacy-typescript", name))
		if err != nil {
			t.Fatal(err)
		}
		var raw map[string]any
		if err := json.Unmarshal(content, &raw); err != nil {
			t.Fatal(err)
		}
		// Presence checks run on the raw document: typed zero values
		// cannot distinguish a missing field from an empty one.
		for _, field := range []string{"name", "provenance", "command", "stdin", "stdout", "stderr", "exitCode", "schemaVersion", "intentionalDifferences"} {
			if _, ok := raw[field]; !ok {
				t.Fatalf("fixture %s lacks %q", name, field)
			}
		}
		var fixture legacyFixture
		if err := json.Unmarshal(content, &fixture); err != nil {
			t.Fatal(err)
		}
		if len(fixture.Command) == 0 || fixture.SchemaVersion != "legacy-typescript/1" ||
			len(fixture.IntentionalDifferences) == 0 || len(fixture.Provenance) == 0 {
			t.Fatalf("fixture %s lacks command/stdin/stdout/stderr/exit/schema metadata", name)
		}
		assertInventoriedCommand(t, name, fixture.Command)
		fixtures[fixture.Name] = fixture
	}
	return fixtures
}

// assertInventoriedCommand restricts parity to commands the legacy CLI
// actually implements. A fixture file for doctor, start, or any other
// new-contract command fails here, not in review.
func assertInventoriedCommand(t *testing.T, name string, command []string) {
	t.Helper()
	if len(command) == 1 && command[0] == "--version" {
		return
	}
	if len(command) > 0 && command[0] == "status" {
		return
	}
	t.Fatalf("fixture %s claims parity for uninventoried command %q", name, command)
}

func runNew(t *testing.T, argv []string, workdir string) (stdout string, exit int) {
	t.Helper()
	var out, errOut bytes.Buffer
	exit = cli.Execute(context.Background(), strings.NewReader(""), &out, &errOut, argv, testEnv(workdir))
	return out.String(), exit
}

func TestLegacyFixturesAreComplete(t *testing.T) {
	t.Parallel()

	fixtures := loadLegacyFixtures(t)
	for _, name := range []string{"version", "status-empty", "status-empty-json", "status-one", "status-one-json", "status-missing-json"} {
		if _, ok := fixtures[name]; !ok {
			t.Fatalf("missing frozen fixture %q", name)
		}
	}
}

func TestVersionParity(t *testing.T) {
	t.Parallel()

	fixtures := loadLegacyFixtures(t)
	legacy := fixtures["version"]
	if legacy.Stdout != "0.1.0\n" || legacy.ExitCode != 0 {
		t.Fatalf("frozen version fixture changed: %#v", legacy)
	}
	stdout, exit := runNew(t, []string{"version"}, t.TempDir())
	if exit != legacy.ExitCode {
		t.Fatalf("version exit parity broken: legacy=%d new=%d", legacy.ExitCode, exit)
	}
	if !strings.HasPrefix(stdout, "sdd ") || stdout == legacy.Stdout {
		t.Fatalf("version format difference undocumented: %q", stdout)
	}
}

func TestStatusEmptyParity(t *testing.T) {
	t.Parallel()

	fixtures := loadLegacyFixtures(t)
	legacy := fixtures["status-empty-json"]
	var legacyEnvelope struct {
		Items []any `json:"items"`
	}
	if err := json.Unmarshal([]byte(legacy.Stdout), &legacyEnvelope); err != nil {
		t.Fatal(err)
	}

	// Same native shape on both sides: zero items, exit 0.
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".sdd"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".sdd", "workspace.yaml"),
		[]byte("formatVersion: 1\nworkspaceId: ws-1\nrevision: 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stdout, exit := runNew(t, []string{"status", "--json"}, root)
	if exit != 0 {
		t.Fatalf("empty-workspace status exit = %d", exit)
	}
	var envelope struct {
		Result struct {
			WorkItems []any `json:"workItems"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatal(err)
	}
	if len(envelope.Result.WorkItems) != len(legacyEnvelope.Items) {
		t.Fatal("zero-item semantics diverged")
	}

	// Documented divergence: legacy tolerates a missing workspace with
	// silence; the new CLI fails closed with a structured diagnostic.
	missingOut, missingExit := runNew(t, []string{"status", "--json"}, t.TempDir())
	if missingExit != 3 {
		t.Fatalf("missing-workspace status exit = %d", missingExit)
	}
	var missing struct {
		Diagnostics []struct {
			Code string `json:"code"`
		} `json:"diagnostics"`
	}
	if err := json.Unmarshal([]byte(missingOut), &missing); err != nil {
		t.Fatal(err)
	}
	if len(missing.Diagnostics) != 1 || missing.Diagnostics[0].Code != "DIAG_WORKSPACE_NOT_FOUND" {
		t.Fatalf("missing-workspace diagnostic: %s", missingOut)
	}
}

func TestStatusOneParity(t *testing.T) {
	t.Parallel()

	fixtures := loadLegacyFixtures(t)
	legacy := fixtures["status-one-json"]
	var legacyEnvelope struct {
		Items []any `json:"items"`
	}
	if err := json.Unmarshal([]byte(legacy.Stdout), &legacyEnvelope); err != nil {
		t.Fatal(err)
	}

	root := t.TempDir()
	writeWorkspace(t, root)
	stdout, exit := runNew(t, []string{"status", "--json"}, root)
	if exit != 0 {
		t.Fatalf("one-item status exit = %d", exit)
	}
	var envelope struct {
		Result struct {
			Kind      string `json:"kind"`
			WorkItems []any  `json:"workItems"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Result.Kind != "status" || len(envelope.Result.WorkItems) != len(legacyEnvelope.Items) {
		t.Fatalf("one-item parity broken: %s", stdout)
	}
}

func TestStatusMissingIDDivergence(t *testing.T) {
	t.Parallel()

	fixtures := loadLegacyFixtures(t)
	if fixtures["status-missing-json"].ExitCode != 0 {
		t.Fatal("legacy missing-ID fixture must record the inventing behavior (exit 0)")
	}
	root := t.TempDir()
	writeWorkspace(t, root)
	stdout, exit := runNew(t, []string{"status", "--work", "missing-id", "--json"}, root)
	if exit != 3 {
		t.Fatalf("missing-ID status exit = %d", exit)
	}
	var envelope struct {
		Result      any `json:"result"`
		Diagnostics []struct {
			Code string `json:"code"`
		} `json:"diagnostics"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Result != nil {
		t.Fatalf("missing-ID status must carry a null result: %s", stdout)
	}
	if len(envelope.Diagnostics) != 1 || envelope.Diagnostics[0].Code != "DIAG_WORK_ITEM_NOT_FOUND" {
		t.Fatalf("missing-ID diagnostic: %s", stdout)
	}
}
