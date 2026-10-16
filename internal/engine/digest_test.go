package engine_test

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"sdd.dev/sdd/internal/engine"
	"sdd.dev/sdd/internal/testkit"
)

var wireDigestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

func TestBlobDigestBindsExactBytes(t *testing.T) {
	t.Parallel()

	content := []byte("workId: work-1\n")
	digest := engine.BlobDigest("sdd.raw-blob/1", content)
	if !wireDigestPattern.MatchString(digest) {
		t.Fatalf("blob digest is not wire-formed: %q", digest)
	}
	if got := engine.BlobDigest("sdd.raw-blob/1", []byte("workId: work-1\n")); got != digest {
		t.Fatal("blob digest is not deterministic")
	}
	changed := append([]byte{}, content...)
	changed[0] = 'X'
	if got := engine.BlobDigest("sdd.raw-blob/1", changed); got == digest {
		t.Fatal("one changed byte did not change the blob digest")
	}
	if got := engine.BlobDigest("sdd.other-domain/1", content); got == digest {
		t.Fatal("blob digest lacks domain separation")
	}
	// The digest input is the observed bytes, never a Git object: dirty
	// working content not represented by HEAD still binds exactly.
	facts := testkit.NewMemFileFacts(map[string]testkit.MemFile{
		"/repo/dirty.yaml": {Bytes: content},
	})
	observed, err := facts.ReadFile(context.Background(), "/repo/dirty.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if got := engine.BlobDigest("sdd.raw-blob/1", observed); got != digest {
		t.Fatal("blob digest over observed bytes diverged from the byte digest")
	}
}

func TestSemanticDigestIgnoresFormattingOnlyChanges(t *testing.T) {
	t.Parallel()

	compact := "schemaVersion: 1\nworkId: work-1\nslug: one\nrevision: 4\nstatus: active\ncomposition: quick-spec\n"
	formatted := "# human comment\nslug: \"one\"\n\nrevision: 4 # trailing\nstatus: active\ncomposition: quick-spec\nworkId: 'work-1'\nschemaVersion: 1\n"
	facts := testkit.NewMemFileFacts(map[string]testkit.MemFile{
		"/a/.sdd/workspace.yaml":     {Bytes: []byte("formatVersion: 1\nworkspaceId: ws-1\nrevision: 2\n")},
		"/a/.sdd/work/one/work.yaml": {Bytes: []byte(compact)},
		"/b/.sdd/workspace.yaml":     {Bytes: []byte("formatVersion: 1\nworkspaceId: ws-1\nrevision: 2\n")},
		"/b/.sdd/work/one/work.yaml": {Bytes: []byte(formatted)},
	})
	first, err := engine.DiscoverLayout(context.Background(), facts, "/a", "")
	if err != nil {
		t.Fatal(err)
	}
	second, err := engine.DiscoverLayout(context.Background(), facts, "/b", "")
	if err != nil {
		t.Fatal(err)
	}
	firstRecord := first.Works()[0]
	secondRecord := second.Works()[0]
	if engine.SemanticDigestWorkRecord(firstRecord) != engine.SemanticDigestWorkRecord(secondRecord) {
		t.Fatal("formatting-only changes altered the semantic digest")
	}
	firstBytes, _ := facts.ReadFile(context.Background(), "/a/.sdd/work/one/work.yaml")
	secondBytes, _ := facts.ReadFile(context.Background(), "/b/.sdd/work/one/work.yaml")
	if engine.BlobDigest("sdd.raw-blob/1", firstBytes) == engine.BlobDigest("sdd.raw-blob/1", secondBytes) {
		t.Fatal("formatting-only changes did not alter the blob digest")
	}

	changed, err := engine.DiscoverLayout(context.Background(), testkit.NewMemFileFacts(map[string]testkit.MemFile{
		"/c/.sdd/workspace.yaml":     {Bytes: []byte("formatVersion: 1\nworkspaceId: ws-1\nrevision: 2\n")},
		"/c/.sdd/work/one/work.yaml": {Bytes: []byte("schemaVersion: 1\nworkId: work-1\nslug: one\nrevision: 4\nstatus: done\ncomposition: quick-spec\n")},
	}), "/c", "")
	if err != nil {
		t.Fatal(err)
	}
	if engine.SemanticDigestWorkRecord(changed.Works()[0]) == engine.SemanticDigestWorkRecord(firstRecord) {
		t.Fatal("a status change did not alter the semantic digest")
	}
	if got := engine.SemanticDigestWorkspace("ws-1", 1, 2); !wireDigestPattern.MatchString(got) {
		t.Fatalf("workspace semantic digest is not wire-formed: %q", got)
	}
	if engine.SemanticDigestWorkspace("ws-1", 1, 2) == engine.SemanticDigestWorkspace("ws-1", 1, 3) {
		t.Fatal("a revision change did not alter the workspace semantic digest")
	}
}

func TestDerivationDigestBindsUpstreamIngredients(t *testing.T) {
	t.Parallel()

	semantic := engine.BlobDigest("sdd.test-semantic/1", []byte("meaning"))
	policyA := engine.BlobDigest("sdd.test-policy/1", []byte("strict"))
	policyB := engine.BlobDigest("sdd.test-policy/1", []byte("lenient"))
	first, err := engine.DerivationDigest("sdd.dry-run-derivation/1", semantic, policyA)
	if err != nil {
		t.Fatal(err)
	}
	if !wireDigestPattern.MatchString(first) {
		t.Fatalf("derivation digest is not wire-formed: %q", first)
	}
	second, err := engine.DerivationDigest("sdd.dry-run-derivation/1", semantic, policyB)
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("an upstream ingredient change did not alter the derivation digest")
	}
	reordered, err := engine.DerivationDigest("sdd.dry-run-derivation/1", semantic, policyB, policyA)
	if err != nil {
		t.Fatal(err)
	}
	ordered, err := engine.DerivationDigest("sdd.dry-run-derivation/1", semantic, policyA, policyB)
	if err != nil {
		t.Fatal(err)
	}
	if reordered == ordered {
		t.Fatal("upstream order did not alter the derivation digest")
	}

	for _, input := range [][3]string{
		{"sdd.d/1", "bad", policyA},
		{"sdd.d/1", semantic, "bad"},
	} {
		if _, err := engine.DerivationDigest(input[0], input[1], input[2]); !errors.Is(err, engine.ErrInvalidProtocolValue) {
			t.Fatalf("malformed derivation input error = %v", err)
		}
	}
}

func planDigestFixture(t *testing.T) (engine.PlanDigestInput, string) {
	t.Helper()
	content := "workId: work-1\n"
	operation, err := engine.NewPlanOperation(engine.PlanOperationInput{
		Kind:             engine.OperationCreateFile,
		Path:             ".sdd/work/new/work.yaml",
		Content:          &content,
		ContentDigest:    "sha256:133452e2e3d639d26930125c30cbfe15eb23364cb6555c78b63c00842f69f750",
		ExpectedRevision: 4,
	})
	if err != nil {
		t.Fatal(err)
	}
	precondition, err := engine.NewPlanPrecondition(engine.PlanPreconditionInput{
		Kind:             engine.PreconditionAbsent,
		Path:             ".sdd/work/new/work.yaml",
		ExpectedRevision: 4,
	})
	if err != nil {
		t.Fatal(err)
	}
	digests := map[string]string{
		"input":     "sha256:b727d8da7e2cbd385d9e474234dda18a9677913b37f40a8074b410d2efbce58c",
		"policy":    "sha256:6d96b4f675f734210734c22f9d1f78eb21ee03316fe04068cf64f1ebdf97f3a6",
		"workflow":  "sha256:44e103af0242e49bc680a1ee30f96c847814cdf79401b1f6f10b1b7275bc623d",
		"validator": "sha256:e25d7eb5474afb9a12bf09813d869079e847bf64c9d79c12b68939870af944a6",
	}
	input := engine.PlanDigestInput{
		PlanID: "plan-1", WorkspaceID: "ws-1", WorkID: "work-1",
		RevisionScope: engine.RevisionWorkspace, ExpectedRevision: 4,
		InputDigest: digests["input"], PolicyDigest: digests["policy"],
		WorkflowDefinitionDigest: digests["workflow"], ValidatorDefinitionDigest: digests["validator"],
		Operations: []engine.PlanOperation{operation}, Preconditions: []engine.PlanPrecondition{precondition},
		IdempotencyKey: "idem-1",
	}
	digest, err := engine.ComputePlanDigest(input)
	if err != nil {
		t.Fatal(err)
	}
	if !wireDigestPattern.MatchString(digest) {
		t.Fatalf("plan digest is not wire-formed: %q", digest)
	}
	return input, digest
}

func TestPlanDigestBindsEveryDecisionField(t *testing.T) {
	t.Parallel()

	input, digest := planDigestFixture(t)
	plan, err := engine.NewMutationPlan(engine.MutationPlanInput{
		PlanID: input.PlanID, WorkspaceID: input.WorkspaceID, WorkID: input.WorkID,
		RevisionScope: input.RevisionScope, ExpectedRevision: input.ExpectedRevision,
		InputDigest: input.InputDigest, PolicyDigest: input.PolicyDigest,
		WorkflowDefinitionDigest:  input.WorkflowDefinitionDigest,
		ValidatorDefinitionDigest: input.ValidatorDefinitionDigest,
		Operations:                input.Operations, Preconditions: input.Preconditions,
		Digest: digest, IdempotencyKey: input.IdempotencyKey,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.VerifyPlanDigest(plan); err != nil {
		t.Fatalf("valid plan failed verification: %v", err)
	}

	mutations := map[string]func(*engine.PlanDigestInput){
		"changed content": func(in *engine.PlanDigestInput) {
			content := "workId: work-2\n"
			operation, _ := engine.NewPlanOperation(engine.PlanOperationInput{
				Kind:             engine.OperationCreateFile,
				Path:             ".sdd/work/new/work.yaml",
				Content:          &content,
				ContentDigest:    "sha256:133452e2e3d639d26930125c30cbfe15eb23364cb6555c78b63c00842f69f750",
				ExpectedRevision: 4,
			})
			in.Operations = []engine.PlanOperation{operation}
		},
		"changed revision":      func(in *engine.PlanDigestInput) { in.ExpectedRevision = 5 },
		"changed policy digest": func(in *engine.PlanDigestInput) { in.PolicyDigest = input.InputDigest },
		"changed idempotency":   func(in *engine.PlanDigestInput) { in.IdempotencyKey = "idem-2" },
		"changed work ID":       func(in *engine.PlanDigestInput) { in.WorkID = "work-2" },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			changed := input
			changed.Operations = append([]engine.PlanOperation{}, input.Operations...)
			changed.Preconditions = append([]engine.PlanPrecondition{}, input.Preconditions...)
			mutate(&changed)
			other, err := engine.ComputePlanDigest(changed)
			if err != nil {
				t.Fatal(err)
			}
			if other == digest {
				t.Fatalf("%s did not change the plan digest", name)
			}
			// A recomputed plan carrying the old digest cannot inherit approval.
			stale, err := engine.NewMutationPlan(engine.MutationPlanInput{
				PlanID: changed.PlanID, WorkspaceID: changed.WorkspaceID, WorkID: changed.WorkID,
				RevisionScope: changed.RevisionScope, ExpectedRevision: changed.ExpectedRevision,
				InputDigest: changed.InputDigest, PolicyDigest: changed.PolicyDigest,
				WorkflowDefinitionDigest:  changed.WorkflowDefinitionDigest,
				ValidatorDefinitionDigest: changed.ValidatorDefinitionDigest,
				Operations:                changed.Operations, Preconditions: changed.Preconditions,
				Digest: digest, IdempotencyKey: changed.IdempotencyKey,
			})
			if err != nil {
				t.Fatal(err)
			}
			if err := engine.VerifyPlanDigest(stale); err == nil {
				t.Fatalf("%s: stale digest verified against changed fields", name)
			}
		})
	}
}

func TestPlanDigestRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	input, _ := planDigestFixture(t)
	invalid := []func(*engine.PlanDigestInput){
		func(in *engine.PlanDigestInput) { in.PlanID = "" },
		func(in *engine.PlanDigestInput) { in.RevisionScope = "future" },
		func(in *engine.PlanDigestInput) { in.PolicyDigest = "bad" },
		func(in *engine.PlanDigestInput) { in.Operations = nil },
		func(in *engine.PlanDigestInput) { in.Preconditions = nil },
	}
	for index, mutate := range invalid {
		changed := input
		mutate(&changed)
		if _, err := engine.ComputePlanDigest(changed); !errors.Is(err, engine.ErrInvalidProtocolValue) {
			t.Fatalf("invalid plan digest input %d error = %v", index, err)
		}
	}
}
