package workspace

import (
	"errors"
	"testing"
	"time"
)

func TestApplyMachineDrivesExactCommitAndReplay(t *testing.T) {
	t.Parallel()

	planInput := validPlanInput(t)
	planInput.Effects = planInput.Effects[:1]
	plan, err := NewPlan(planInput)
	if err != nil {
		t.Fatal(err)
	}
	now := planInput.Descriptor.CreatedAt.Add(time.Minute)
	fact := validAuthorization(plan, now)
	content := []ContentBlob{{Digest: planInput.Effects[0].Content.BlobDigest, Bytes: []byte("first")}}
	candidate, err := NewApplyCandidate("tx-1", plan, content, fact, IdempotencyInput{OperationKind: "apply", PrincipalID: "agent-1", Key: "request-1"})
	if err != nil {
		t.Fatal(err)
	}
	if candidate.RequestFingerprint() == "" {
		t.Fatal("request fingerprint is empty")
	}
	record := NewApplyRecord(candidate)
	prepare := &PrepareFacts{Generation: plan.SourceGeneration(), ReadSetDigest: plan.ReadSetDigest(), Now: now}

	wantKinds := []ApplyActionKind{ApplyBegin, ApplyBeginStaging, ApplyStageBlob, ApplyPrepare, ApplyInstallEffect, ApplyCommit, ApplyReplay}
	for _, want := range wantKinds {
		observation := EffectObservation("")
		if want == ApplyInstallEffect {
			observation = EffectMatchesBefore
		}
		action, nextErr := NextApplyAction(record, prepare, observation)
		if nextErr != nil {
			t.Fatalf("NextApplyAction() error = %v", nextErr)
		}
		if action.Kind != want {
			t.Fatalf("action = %q, want %q", action.Kind, want)
		}
		record, err = AcknowledgeApplyAction(record, action)
		if err != nil {
			t.Fatalf("AcknowledgeApplyAction(%q) error = %v", action.Kind, err)
		}
	}
	if record.State() != TransactionCommitted {
		t.Fatalf("state = %q", record.State())
	}
}

func TestApplyMachineRecognizesInstalledEffectAfterLostProgress(t *testing.T) {
	t.Parallel()

	record, prepare := preparedApplyRecord(t)
	action, err := NextApplyAction(record, prepare, EffectMatchesAfter)
	if err != nil {
		t.Fatal(err)
	}
	if action.Kind != ApplyRecognizeEffect || action.EffectIndex != 0 {
		t.Fatalf("recovery action = %#v", action)
	}
	record, err = AcknowledgeApplyAction(record, action)
	if err != nil {
		t.Fatal(err)
	}
	action, err = NextApplyAction(record, prepare, "")
	if err != nil || action.Kind != ApplyCommit {
		t.Fatalf("post-recognition action = %#v, %v", action, err)
	}
}

func TestApplyMachineQuarantinesThirdStateAfterPrepared(t *testing.T) {
	t.Parallel()

	record, prepare := preparedApplyRecord(t)
	action, err := NextApplyAction(record, prepare, EffectConflicts)
	if err != nil {
		t.Fatal(err)
	}
	if action.Kind != ApplyQuarantine {
		t.Fatalf("action = %#v", action)
	}
	record, err = AcknowledgeApplyAction(record, action)
	if err != nil {
		t.Fatal(err)
	}
	if record.State() != TransactionQuarantined {
		t.Fatalf("state = %q", record.State())
	}
	if record.TerminalReason() != "prepared-precondition-changed" {
		t.Fatalf("terminal reason = %q", record.TerminalReason())
	}
}

func TestApplyMachineFailsBeforeIntentOnChangedSnapshotOrGrant(t *testing.T) {
	t.Parallel()

	record, prepare := stagingApplyRecord(t)
	prepare.Generation++
	action, err := NextApplyAction(record, prepare, "")
	if err != nil || action.Kind != ApplyFailPreIntent || action.Reason != "precondition-changed" {
		t.Fatalf("changed snapshot action = %#v, %v", action, err)
	}

	record, prepare = stagingApplyRecord(t)
	prepare.Now = record.candidate.authorization.ExpiresAt
	action, err = NextApplyAction(record, prepare, "")
	if err != nil || action.Kind != ApplyFailPreIntent || action.Reason != "authorization-invalid" {
		t.Fatalf("expired grant action = %#v, %v", action, err)
	}
}

func TestApplyMachineRejectsFabricatedGuardedAction(t *testing.T) {
	t.Parallel()

	record, _ := stagingApplyRecord(t)
	if _, err := AcknowledgeApplyAction(record, ApplyAction{Kind: ApplyPrepare}); !errors.Is(err, ErrApplyActionInvalid) {
		t.Fatalf("fabricated prepare error = %v", err)
	}
	record, _ = preparedApplyRecord(t)
	if _, err := AcknowledgeApplyAction(record, ApplyAction{Kind: ApplyInstallEffect, EffectIndex: 0}); !errors.Is(err, ErrApplyActionInvalid) {
		t.Fatalf("fabricated install error = %v", err)
	}
}

func TestApplyCandidateRejectsChangedOrExtraContent(t *testing.T) {
	t.Parallel()

	planInput := validPlanInput(t)
	planInput.Effects = planInput.Effects[:1]
	plan, err := NewPlan(planInput)
	if err != nil {
		t.Fatal(err)
	}
	fact := validAuthorization(plan, planInput.Descriptor.CreatedAt.Add(time.Minute))
	idempotency := IdempotencyInput{OperationKind: "apply", PrincipalID: "agent-1", Key: "request-1"}

	tests := []struct {
		name    string
		content []ContentBlob
	}{
		{name: "changed bytes", content: []ContentBlob{{Digest: planInput.Effects[0].Content.BlobDigest, Bytes: []byte("wrong")}}},
		{name: "extra blob", content: []ContentBlob{{Digest: planInput.Effects[0].Content.BlobDigest, Bytes: []byte("first")}, {Digest: sha256OfBytes([]byte("extra")), Bytes: []byte("extra")}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, candidateErr := NewApplyCandidate("tx-1", plan, test.content, fact, idempotency); !errors.Is(candidateErr, ErrApplyInvalid) {
				t.Fatalf("NewApplyCandidate() error = %v", candidateErr)
			}
		})
	}
}

func preparedApplyRecord(t *testing.T) (ApplyRecord, *PrepareFacts) {
	t.Helper()
	record, prepare := stagingApplyRecord(t)
	action, err := NextApplyAction(record, prepare, "")
	if err != nil || action.Kind != ApplyPrepare {
		t.Fatalf("prepare action = %#v, %v", action, err)
	}
	record, err = AcknowledgeApplyAction(record, action)
	if err != nil {
		t.Fatal(err)
	}
	return record, prepare
}

func stagingApplyRecord(t *testing.T) (ApplyRecord, *PrepareFacts) {
	t.Helper()
	planInput := validPlanInput(t)
	planInput.Effects = planInput.Effects[:1]
	plan, err := NewPlan(planInput)
	if err != nil {
		t.Fatal(err)
	}
	now := planInput.Descriptor.CreatedAt.Add(time.Minute)
	fact := validAuthorization(plan, now)
	candidate, err := NewApplyCandidate("tx-1", plan, []ContentBlob{{Digest: planInput.Effects[0].Content.BlobDigest, Bytes: []byte("first")}}, fact, IdempotencyInput{OperationKind: "apply", PrincipalID: "agent-1", Key: "request-1"})
	if err != nil {
		t.Fatal(err)
	}
	record := NewApplyRecord(candidate)
	for _, observation := range []EffectObservation{"", "", ""} {
		action, nextErr := NextApplyAction(record, nil, observation)
		if nextErr != nil {
			t.Fatal(nextErr)
		}
		record, err = AcknowledgeApplyAction(record, action)
		if err != nil {
			t.Fatal(err)
		}
	}
	return record, &PrepareFacts{Generation: plan.SourceGeneration(), ReadSetDigest: plan.ReadSetDigest(), Now: now}
}
