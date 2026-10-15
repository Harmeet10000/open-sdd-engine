package protocol

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"slices"
	"strings"
	"testing"
)

type inventory struct {
	InventorySchemaVersion string `json:"inventorySchemaVersion"`
	PublishedVersion       string `json:"publishedVersion"`
	LatestAllocatedVersion string `json:"latestAllocatedVersion"`
	Envelope               struct {
		Owner                  string   `json:"owner"`
		VersionField           string   `json:"versionField"`
		ForbiddenVersionFields []string `json:"forbiddenVersionFields"`
		RequiredFields         []string `json:"requiredFields"`
	} `json:"envelope"`
	Versions []struct {
		Version     string   `json:"version"`
		Status      string   `json:"status"`
		Extends     *string  `json:"extends"`
		Commands    []string `json:"commands"`
		ResultKinds []string `json:"resultKinds"`
	} `json:"versions"`
	Allocations []struct {
		ID           string   `json:"id"`
		Owner        string   `json:"owner"`
		IntroducedIn string   `json:"introducedIn"`
		Status       string   `json:"status"`
		Owns         []string `json:"owns"`
		EffectKinds  []string `json:"effectKinds"`
	} `json:"allocations"`
}

type catalog struct {
	CatalogSchemaVersion string `json:"catalogSchemaVersion"`
	Namespaces           []struct {
		Prefix        string   `json:"prefix"`
		RegistryOwner string   `json:"registryOwner"`
		Contributors  []string `json:"contributors"`
	} `json:"namespaces"`
	Codes []struct {
		Code        string   `json:"code"`
		Owner       string   `json:"owner"`
		Severity    string   `json:"severity"`
		ExitCode    int      `json:"exitCode"`
		SourceKinds []string `json:"sourceKinds"`
		Remediation string   `json:"remediation"`
		Retryable   bool     `json:"retryable"`
	} `json:"codes"`
}

func TestExtensionInventoryOwnsOneEnvelope(t *testing.T) {
	var got inventory
	decodeStrict(t, "extension-inventory.json", &got)

	if got.InventorySchemaVersion != "sdd.protocol-extension-inventory/1" {
		t.Fatalf("unexpected inventory schema %q", got.InventorySchemaVersion)
	}
	if got.PublishedVersion != "1.0" || got.LatestAllocatedVersion != "1.1" {
		t.Fatalf("unexpected version allocation: published=%q allocated=%q", got.PublishedVersion, got.LatestAllocatedVersion)
	}
	if got.Envelope.Owner != "sdd-engine-foundation" || got.Envelope.VersionField != "schemaVersion" {
		t.Fatalf("foundation must own schemaVersion envelope: %#v", got.Envelope)
	}
	if !slices.Contains(got.Envelope.ForbiddenVersionFields, "protocolVersion") {
		t.Fatal("protocolVersion must be explicitly forbidden")
	}

	wantFields := []string{"schemaVersion", "command", "ok", "workspace", "work", "result", "diagnostics", "nextActions", "exitCode"}
	if !slices.Equal(got.Envelope.RequiredFields, wantFields) {
		t.Fatalf("envelope changed: got %v want %v", got.Envelope.RequiredFields, wantFields)
	}

	versionStatus := make(map[string]string, len(got.Versions))
	commands := map[string]string{}
	resultKinds := map[string]string{}
	for _, version := range got.Versions {
		if _, exists := versionStatus[version.Version]; exists {
			t.Fatalf("duplicate version %q", version.Version)
		}
		versionStatus[version.Version] = version.Status
		for _, command := range version.Commands {
			if prior, exists := commands[command]; exists {
				t.Fatalf("command %q allocated in %s and %s", command, prior, version.Version)
			}
			commands[command] = version.Version
		}
		for _, kind := range version.ResultKinds {
			if prior, exists := resultKinds[kind]; exists {
				t.Fatalf("result kind %q allocated in %s and %s", kind, prior, version.Version)
			}
			resultKinds[kind] = version.Version
		}
	}
	if versionStatus["1.0"] != "published" || versionStatus["1.1"] != "allocated" {
		t.Fatalf("version statuses do not preserve the publication gate: %v", versionStatus)
	}

	allocationIDs := map[string]struct{}{}
	ownedNames := map[string]string{}
	for _, allocation := range got.Allocations {
		if allocation.ID == "" || allocation.Owner == "" || allocation.IntroducedIn == "" || len(allocation.Owns) == 0 {
			t.Fatalf("incomplete allocation: %#v", allocation)
		}
		if _, exists := versionStatus[allocation.IntroducedIn]; !exists {
			t.Fatalf("allocation %q uses unknown version %q", allocation.ID, allocation.IntroducedIn)
		}
		if _, exists := allocationIDs[allocation.ID]; exists {
			t.Fatalf("duplicate allocation ID %q", allocation.ID)
		}
		allocationIDs[allocation.ID] = struct{}{}
		for _, name := range allocation.Owns {
			if prior, exists := ownedNames[name]; exists {
				t.Fatalf("contract %q has competing owners %q and %q", name, prior, allocation.Owner)
			}
			ownedNames[name] = allocation.Owner
		}
	}
}

func TestDiagnosticCatalogIsCompleteAndDeterministic(t *testing.T) {
	var got catalog
	decodeStrict(t, "diagnostics.catalog.json", &got)
	if got.CatalogSchemaVersion != "sdd.diagnostic-catalog/1" {
		t.Fatalf("unexpected catalog schema %q", got.CatalogSchemaVersion)
	}

	validSeverity := map[string]bool{"info": true, "warning": true, "error": true, "critical": true}
	validExit := map[int]bool{0: true, 1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 130: true}
	namespaces := map[string]struct{}{}
	for _, namespace := range got.Namespaces {
		if namespace.Prefix == "" || namespace.RegistryOwner != "sdd-engine-foundation" || len(namespace.Contributors) == 0 {
			t.Errorf("invalid diagnostic namespace: %#v", namespace)
		}
		if _, exists := namespaces[namespace.Prefix]; exists {
			t.Errorf("duplicate diagnostic namespace %q", namespace.Prefix)
		}
		namespaces[namespace.Prefix] = struct{}{}
	}
	for _, prefix := range []string{"DIAG_", "WORKFLOW_", "VALIDATION_", "COMPLETION_", "BUG_", "QUICK_SPEC_"} {
		if _, exists := namespaces[prefix]; !exists {
			t.Errorf("required diagnostic namespace %q is missing", prefix)
		}
	}

	seen := map[string]struct{}{}
	for _, entry := range got.Codes {
		if !strings.HasPrefix(entry.Code, "DIAG_") {
			t.Errorf("code %q does not use the foundation namespace", entry.Code)
		}
		if _, exists := seen[entry.Code]; exists {
			t.Errorf("duplicate diagnostic code %q", entry.Code)
		}
		seen[entry.Code] = struct{}{}
		if entry.Owner == "" || !validSeverity[entry.Severity] || !validExit[entry.ExitCode] || len(entry.SourceKinds) == 0 || entry.Remediation == "" {
			t.Errorf("incomplete diagnostic entry: %#v", entry)
		}
	}

	for _, code := range []string{
		"DIAG_SUCCESS",
		"DIAG_USAGE_ERROR",
		"DIAG_COMMAND_UNKNOWN",
		"DIAG_WORKSPACE_NOT_FOUND",
		"DIAG_WORKSPACE_VERSION_INCOMPATIBLE",
		"DIAG_WORKSPACE_SELECTION_AMBIGUOUS",
		"DIAG_WORK_ITEM_NOT_FOUND",
		"DIAG_STALE_REVISION",
		"DIAG_PLAN_NOT_AUTHORIZABLE",
		"DIAG_INTERRUPTED",
		"DIAG_SNAPSHOT_UNSTABLE",
		"DIAG_RECOVERY_REQUIRED",
	} {
		if _, exists := seen[code]; !exists {
			t.Errorf("required diagnostic %q is missing", code)
		}
	}
}

func TestVersionFileTracksPublishedContract(t *testing.T) {
	b, err := os.ReadFile("VERSION")
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(b)); got != "1.0" {
		t.Fatalf("VERSION advertises %q; allocated 1.1 is not published yet", got)
	}
}

func decodeStrict(t *testing.T, path string, target any) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(strings.NewReader(string(b)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		t.Fatalf("decode %s: expected one JSON document, got %v", path, err)
	}
}
