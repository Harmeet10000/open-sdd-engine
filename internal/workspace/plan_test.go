package workspace

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestPlanDigestBindsOrderAndExcludesFinalGrant(t *testing.T) {
	t.Parallel()

	input := validPlanInput(t)
	plan, err := NewPlan(input)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Digest() == "" {
		t.Fatal("plan digest is empty")
	}

	reordered := input
	reordered.Effects = cloneEffects(input.Effects)
	reordered.Effects[0], reordered.Effects[1] = reordered.Effects[1], reordered.Effects[0]
	reordered.Effects[0].Index = 0
	reordered.Effects[1].Index = 1
	other, err := NewPlan(reordered)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Digest() == other.Digest() {
		t.Fatal("effect order did not change plan identity")
	}

	now := input.Descriptor.CreatedAt.Add(time.Minute)
	fact := validAuthorization(plan, now)
	fact.GrantID = "another-grant"
	if err := VerifyAuthorization(plan, fact, now); err != nil {
		t.Fatalf("final grant ID unexpectedly changed plan validity: %v", err)
	}
}

func TestPlanOwnsInputSlicesAndContent(t *testing.T) {
	t.Parallel()

	input := validPlanInput(t)
	plan, err := NewPlan(input)
	if err != nil {
		t.Fatal(err)
	}
	want := plan.Digest()
	input.Effects[0].Path = mustPath(t, "changed")
	input.Effects[0].Content.Size++
	input.RevisionIntents[0].Proposed++
	if plan.Digest() != want || plan.Effects()[0].Path == "changed" {
		t.Fatal("plan aliases constructor input")
	}
	returned := plan.Effects()
	returned[0].Content.Size++
	if plan.Effects()[0].Content.Size == returned[0].Content.Size {
		t.Fatal("plan aliases accessor output")
	}
}

func TestPlanRejectsUnsafeOrInexactEffects(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*PlanInput)
		want   error
	}{
		{name: "noncanonical path", mutate: func(input *PlanInput) { input.Effects[0].Path = "../escape" }, want: ErrPlanInvalid},
		{name: "duplicate path", mutate: func(input *PlanInput) { input.Effects[1].Path = input.Effects[0].Path }, want: ErrPlanInvalid},
		{name: "create has old digest", mutate: func(input *PlanInput) { input.Effects[0].Before.BlobDigest = sha256Of("old") }, want: ErrPlanInvalid},
		{name: "update lacks content", mutate: func(input *PlanInput) { input.Effects[1].Content = nil }, want: ErrPlanInvalid},
		{name: "wrong revision jump", mutate: func(input *PlanInput) { input.RevisionIntents[0].Proposed += 2 }, want: ErrPlanInvalid},
		{name: "staging limit", mutate: func(input *PlanInput) { input.Limits.MaxStagedBytes = 1 }, want: ErrLimitExceeded},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			input := validPlanInput(t)
			test.mutate(&input)
			if _, err := NewPlan(input); !errors.Is(err, test.want) {
				t.Fatalf("NewPlan() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestVerifyAuthorizationFailsClosed(t *testing.T) {
	t.Parallel()

	input := validPlanInput(t)
	plan, err := NewPlan(input)
	if err != nil {
		t.Fatal(err)
	}
	now := input.Descriptor.CreatedAt.Add(time.Minute)

	tests := []struct {
		name   string
		mutate func(*AuthorizationFact)
		want   error
	}{
		{name: "wrong plan", mutate: func(fact *AuthorizationFact) { fact.PlanDigest = sha256Of("other") }, want: ErrAuthorizationInvalid},
		{name: "expired", mutate: func(fact *AuthorizationFact) { fact.ExpiresAt = now }, want: ErrAuthorizationInvalid},
		{name: "revoked", mutate: func(fact *AuthorizationFact) { fact.Revoked = true; fact.RevocationID = "rev-1" }, want: ErrAuthorizationRevoked},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			fact := validAuthorization(plan, now)
			test.mutate(&fact)
			if err := VerifyAuthorization(plan, fact, now); !errors.Is(err, test.want) {
				t.Fatalf("VerifyAuthorization() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestPlanJSONDoesNotContainProofOrGrant(t *testing.T) {
	t.Parallel()

	plan, err := NewPlan(validPlanInput(t))
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"proof", "grantId", "transactionId", "staging"} {
		if strings.Contains(string(data), forbidden) {
			t.Fatalf("plan JSON contains non-core field %q: %s", forbidden, data)
		}
	}
}

func TestTransactionStateMachineRejectsRollbackAfterIntent(t *testing.T) {
	t.Parallel()

	state := TransactionAbsent
	for _, event := range []TransactionEvent{TransactionBegin, TransactionBeginStaging, TransactionIntentDurable, TransactionEffectsDurable, TransactionAcknowledgeOrReplay} {
		var err error
		state, err = AdvanceTransaction(state, event)
		if err != nil {
			t.Fatalf("AdvanceTransaction(%q) error = %v", event, err)
		}
	}
	if state != TransactionCommitted {
		t.Fatalf("final state = %q", state)
	}
	if got, err := AdvanceTransaction(TransactionPrepared, TransactionPreIntentFailed); !errors.Is(err, ErrTransactionTransition) || got != TransactionPrepared {
		t.Fatalf("prepared rollback = %q, %v", got, err)
	}
	if got, err := AdvanceTransaction(TransactionPrepared, TransactionPreparedConflict); err != nil || got != TransactionQuarantined {
		t.Fatalf("prepared conflict = %q, %v", got, err)
	}
}

func validPlanInput(t *testing.T) PlanInput {
	t.Helper()
	created := time.Date(2026, 10, 17, 0, 0, 0, 0, time.UTC)
	createContent := []byte("first")
	updateContent := []byte("second")
	return PlanInput{
		Descriptor: NewPlanDescriptor(created, created.Add(10*time.Minute)),
		PlanID:     "plan-1", WorkspaceID: "ws-1", WorkID: "work-1",
		SourceSnapshotID: "snap-1", ReadSetDigest: sha256Of("read-set"), SourceGeneration: 4,
		Authorization: AuthorizationScope{
			ProviderID: "test-authority", PrincipalID: "agent-1", Scope: "workspace.apply",
			PolicyDigest: sha256Of("policy"), PolicyVersion: "policy/1", GrantKind: "exact-plan",
		},
		Effects: []Effect{
			{Index: 0, Kind: EffectCreateFile, Path: mustPath(t, ".sdd/work/item/new.md"), Before: ExpectedPath{Presence: Absent, Kind: FileKindNone}, Content: &ContentRef{BlobDigest: sha256OfBytes(createContent), Size: uint64(len(createContent)), Mode: 0o600}},
			{Index: 1, Kind: EffectUpdateFile, Path: mustPath(t, ".sdd/work/item/work.yaml"), Before: ExpectedPath{Presence: Present, Kind: FileKindRegular, BlobDigest: sha256Of("old"), Size: 3, Mode: 0o600}, Content: &ContentRef{BlobDigest: sha256OfBytes(updateContent), Size: uint64(len(updateContent)), Mode: 0o600}},
		},
		RevisionIntents: []RevisionIntent{{Scope: RevisionWork, OwnerID: "work-1", Expected: 3, Proposed: 4}},
		Limits:          PlanLimits{MaxEffects: 4, MaxStagedBytes: 1024},
	}
}

func validAuthorization(plan Plan, now time.Time) AuthorizationFact {
	scope := plan.AuthorizationScope()
	return AuthorizationFact{
		ProviderID: scope.ProviderID, GrantID: "grant-1", PrincipalID: scope.PrincipalID, Scope: scope.Scope,
		PlanDigest: plan.Digest(), PolicyDigest: scope.PolicyDigest, PolicyVersion: scope.PolicyVersion, GrantKind: scope.GrantKind,
		IssuedAt: now.Add(-time.Minute), NotBefore: now.Add(-time.Minute), ExpiresAt: now.Add(time.Minute),
	}
}

func sha256Of(value string) Digest { return sha256OfBytes([]byte(value)) }

func sha256OfBytes(value []byte) Digest { return digestBytes("test/1", value) }
