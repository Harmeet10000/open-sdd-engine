package engine

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

// This file implements the foundation-owned digest set (FND-006). Scope is
// deliberately narrow per AM-04: digests over foundation records (workspace
// manifest bytes, work record meaning, mutation-plan identity, dry-run
// inputs) and the versioned framing they require. Generic Markdown/YAML
// ASTs, profiles, and semantic/derivation canonicalization of artifact
// content belong to artifact-validation, which will define its own digest
// laws over these same framing rules.
//
// Three digest kinds, never interchangeable:
//   - blob: exact observed bytes, for audit and conflict detection;
//   - semantic: typed meaning after safe canonicalization, insensitive to
//     formatting that carries no meaning;
//   - derivation: a semantic digest plus the digests of the upstream
//     definitions and inputs it was derived from.
//
// Every digest renders as lowercase "sha256:<hex>" on the wire.
func formatDigest(sum []byte) string {
	return "sha256:" + hex.EncodeToString(sum)
}

// BlobDigest binds exact bytes under a caller-supplied domain. Two different
// byte sequences never share a blob digest, even when they mean the same
// thing; use a semantic digest when formatting must not matter. The domain
// must name the representation type and schema version.
func BlobDigest(domain string, content []byte) string {
	h := sha256.New()
	writeDigestFrame(h, domain)
	writeDigestBytes(h, content)
	return formatDigest(h.Sum(nil))
}

// SemanticDigestWorkRecord binds the typed meaning of a work record: ID,
// slug, revision, status, and composition. YAML key order, quoting,
// comments, and blank lines do not affect it; any meaning change does.
func SemanticDigestWorkRecord(record WorkRecord) string {
	h := sha256.New()
	writeDigestFrame(h, "sdd.work-record-semantic/1")
	writeDigestFrame(h, record.id)
	writeDigestFrame(h, record.slug)
	writeDigestUint64(h, record.revision)
	writeDigestFrame(h, record.status)
	writeDigestFrame(h, record.composition)
	return formatDigest(h.Sum(nil))
}

// SemanticDigestWorkspace binds the typed meaning of a workspace manifest:
// workspace ID, format version, and revision.
func SemanticDigestWorkspace(workspaceID string, formatVersion, revision uint64) string {
	h := sha256.New()
	writeDigestFrame(h, "sdd.workspace-semantic/1")
	writeDigestFrame(h, workspaceID)
	writeDigestUint64(h, formatVersion)
	writeDigestUint64(h, revision)
	return formatDigest(h.Sum(nil))
}

// DerivationDigest binds a semantic digest to the exact upstream definition
// and input digests it was derived from, in order. Changing any ingredient
// — policy bytes, workflow definition, validator definition, or inputs —
// changes the derivation even when the derived meaning is textually equal.
func DerivationDigest(domain, semantic string, upstream ...string) (string, error) {
	if !validDigest(semantic) {
		return "", fmt.Errorf("%w: malformed semantic digest", ErrInvalidProtocolValue)
	}
	for _, digest := range upstream {
		if !validDigest(digest) {
			return "", fmt.Errorf("%w: malformed upstream digest", ErrInvalidProtocolValue)
		}
	}
	h := sha256.New()
	writeDigestFrame(h, domain)
	writeDigestFrame(h, semantic)
	writeDigestUint64(h, uint64(len(upstream)))
	for _, digest := range upstream {
		writeDigestFrame(h, digest)
	}
	return formatDigest(h.Sum(nil)), nil
}

// PlanDigestInput carries every decision-bearing plan field except the plan
// digest itself. The digest binds identity, revision scope, all four
// definition/input digests, ordered operations and preconditions, and the
// idempotency key: any change requires a new authorization.
type PlanDigestInput struct {
	PlanID                    string
	WorkspaceID               string
	WorkID                    string
	RevisionScope             RevisionScope
	ExpectedRevision          uint64
	InputDigest               string
	PolicyDigest              string
	WorkflowDefinitionDigest  string
	ValidatorDefinitionDigest string
	Operations                []PlanOperation
	Preconditions             []PlanPrecondition
	IdempotencyKey            string
}

// ComputePlanDigest binds a dry-run plan's decision fields and ordered
// effects into one identity digest. Operation and precondition order is
// significant: reordered effects are a different plan.
func ComputePlanDigest(input PlanDigestInput) (string, error) {
	if input.PlanID == "" || input.WorkspaceID == "" || input.WorkID == "" {
		return "", fmt.Errorf("%w: plan, workspace, and work IDs are required", ErrInvalidProtocolValue)
	}
	if input.RevisionScope != RevisionWorkspace && input.RevisionScope != RevisionWork {
		return "", fmt.Errorf("%w: unsupported revision scope %q", ErrInvalidProtocolValue, input.RevisionScope)
	}
	for name, digest := range map[string]string{
		"input":                input.InputDigest,
		"policy":               input.PolicyDigest,
		"workflow definition":  input.WorkflowDefinitionDigest,
		"validator definition": input.ValidatorDefinitionDigest,
	} {
		if !validDigest(digest) {
			return "", fmt.Errorf("%w: %s digest is malformed", ErrInvalidProtocolValue, name)
		}
	}
	if len(input.Operations) == 0 || len(input.Preconditions) == 0 {
		return "", fmt.Errorf("%w: plan operations and preconditions cannot be empty", ErrInvalidProtocolValue)
	}
	h := sha256.New()
	writeDigestFrame(h, "sdd.mutation-plan/1")
	writeDigestFrame(h, input.PlanID)
	writeDigestFrame(h, input.WorkspaceID)
	writeDigestFrame(h, input.WorkID)
	writeDigestFrame(h, string(input.RevisionScope))
	writeDigestUint64(h, input.ExpectedRevision)
	writeDigestFrame(h, input.InputDigest)
	writeDigestFrame(h, input.PolicyDigest)
	writeDigestFrame(h, input.WorkflowDefinitionDigest)
	writeDigestFrame(h, input.ValidatorDefinitionDigest)
	writeDigestUint64(h, uint64(len(input.Operations)))
	for _, operation := range input.Operations {
		writeDigestFrame(h, string(operation.Kind()))
		writeDigestFrame(h, operation.Path())
		if content := operation.Content(); content == nil {
			writeDigestFrame(h, "content-absent")
		} else {
			writeDigestFrame(h, "content-present")
			writeDigestBytes(h, []byte(*content))
		}
		writeDigestFrame(h, operation.ContentDigest())
		writeDigestFrame(h, operation.ExpectedDigest())
		writeDigestUint64(h, operation.ExpectedRevision())
		writeDigestBool(h, operation.Authorizable())
	}
	writeDigestUint64(h, uint64(len(input.Preconditions)))
	for _, precondition := range input.Preconditions {
		writeDigestFrame(h, string(precondition.Kind()))
		writeDigestFrame(h, precondition.Path())
		writeDigestUint64(h, precondition.ExpectedRevision())
		writeDigestFrame(h, precondition.ExpectedDigest())
	}
	writeDigestFrame(h, input.IdempotencyKey)
	return formatDigest(h.Sum(nil)), nil
}

// VerifyPlanDigest recomputes a built plan's identity digest and reports
// whether the recorded digest still binds the plan's fields. A plan whose
// recorded digest differs cannot inherit an earlier authorization.
func VerifyPlanDigest(plan MutationPlan) error {
	computed, err := ComputePlanDigest(PlanDigestInput{
		PlanID:                    plan.PlanID(),
		WorkspaceID:               plan.WorkspaceID(),
		WorkID:                    plan.WorkID(),
		RevisionScope:             plan.RevisionScope(),
		ExpectedRevision:          plan.ExpectedRevision(),
		InputDigest:               plan.InputDigest(),
		PolicyDigest:              plan.PolicyDigest(),
		WorkflowDefinitionDigest:  plan.WorkflowDefinitionDigest(),
		ValidatorDefinitionDigest: plan.ValidatorDefinitionDigest(),
		Operations:                plan.Operations(),
		Preconditions:             plan.Preconditions(),
		IdempotencyKey:            plan.IdempotencyKey(),
	})
	if err != nil {
		return err
	}
	if computed != plan.Digest() {
		return fmt.Errorf("%w: plan digest does not bind the recorded fields", ErrInvalidProtocolValue)
	}
	return nil
}

type digestWriter interface {
	Write([]byte) (int, error)
}

func writeDigestFrame(w digestWriter, value string) {
	writeDigestUint64(w, uint64(len(value)))
	_, _ = w.Write([]byte(value))
}

func writeDigestBytes(w digestWriter, value []byte) {
	writeDigestUint64(w, uint64(len(value)))
	_, _ = w.Write(value)
}

func writeDigestUint64(w digestWriter, value uint64) {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], value)
	_, _ = w.Write(encoded[:])
}

func writeDigestBool(w digestWriter, value bool) {
	if value {
		_, _ = w.Write([]byte{1})
		return
	}
	_, _ = w.Write([]byte{0})
}
