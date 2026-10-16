package protocol

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

// TestTransactionSeamContractIsVersionedInput validates the structure of the
// future transaction integration contract without executing any scenario.
// Execution belongs to sdd-workspace-transactions; this test only prevents
// contract rot: the pinned version, the plan-binding field set against the
// published schema, and unique scenario IDs.
func TestTransactionSeamContractIsVersionedInput(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile("testdata/integration/transaction-seam-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var contract struct {
		ContractVersion string `json:"contractVersion"`
		Owner           string `json:"owner"`
		PlanBinding     struct {
			Fields []string `json:"fields"`
			Rules  []string `json:"rules"`
		} `json:"planBinding"`
		ScenarioInventory []struct {
			ID       string         `json:"id"`
			Name     string         `json:"name"`
			Inputs   map[string]any `json:"inputs"`
			Expected string         `json:"expected"`
		} `json:"scenarioInventory"`
	}
	if err := json.Unmarshal(content, &contract); err != nil {
		t.Fatal(err)
	}
	if contract.ContractVersion != "sdd.transaction-seam/1" {
		t.Fatalf("contract version = %q", contract.ContractVersion)
	}
	if contract.Owner != "sdd-workspace-transactions" {
		t.Fatalf("contract owner = %q", contract.Owner)
	}
	if len(contract.PlanBinding.Rules) == 0 {
		t.Fatal("contract records no binding rules")
	}

	schema, ok := decodeSchemaJSON(t, "schema/foundation-1.0.json").(map[string]any)
	if !ok {
		t.Fatal("foundation schema is not an object")
	}
	definitions, ok := schema["definitions"].(map[string]any)
	if !ok {
		t.Fatal("foundation schema lacks definitions")
	}
	plan, ok := definitions["MutationPlan"].(map[string]any)
	if !ok {
		t.Fatal("foundation schema lacks MutationPlan")
	}
	required, ok := plan["required"].([]any)
	if !ok || len(required) == 0 {
		t.Fatal("published schema lacks MutationPlan requirements")
	}
	contractFields := slices.Clone(contract.PlanBinding.Fields)
	slices.Sort(contractFields)
	schemaFields := make([]string, 0, len(required))
	for _, field := range required {
		name, ok := field.(string)
		if !ok {
			t.Fatal("schema requirement is not a string")
		}
		schemaFields = append(schemaFields, name)
	}
	slices.Sort(schemaFields)
	if !slices.Equal(contractFields, schemaFields) {
		t.Fatalf("contract fields %v diverge from schema %v", contractFields, schemaFields)
	}

	seen := map[string]struct{}{}
	for _, scenario := range contract.ScenarioInventory {
		if scenario.ID == "" || scenario.Name == "" || scenario.Expected == "" || len(scenario.Inputs) == 0 {
			t.Fatalf("scenario lacks id/name/inputs/expected: %#v", scenario)
		}
		if _, duplicate := seen[scenario.ID]; duplicate {
			t.Fatalf("duplicate scenario ID %q", scenario.ID)
		}
		seen[scenario.ID] = struct{}{}
	}
	if len(seen) == 0 {
		t.Fatal("contract records no scenarios")
	}
}
