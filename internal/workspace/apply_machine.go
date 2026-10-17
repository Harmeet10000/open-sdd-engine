package workspace

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"sdd.dev/sdd/internal/protocol/canonicaljson"
)

var (
	ErrApplyInvalid       = errors.New("workspace apply request invalid")
	ErrApplyActionInvalid = errors.New("workspace apply action invalid")
)

// ContentBlob supplies exact bytes by digest. Bytes are copied into the
// candidate and never identified by a caller-controlled staging path.
type ContentBlob struct {
	Digest Digest
	Bytes  []byte
}

// IdempotencyInput keeps the opaque caller key distinct from the canonical
// request fingerprint. Durable storage hashes the key before using a path.
type IdempotencyInput struct {
	OperationKind string
	PrincipalID   string
	Key           string
}

// ApplyCandidate is an immutable, structurally complete request ready for the
// pure apply/recovery machine. Authorization freshness is intentionally
// rechecked only at the transition to PREPARED.
type ApplyCandidate struct {
	transactionID      string
	plan               Plan
	content            []ContentBlob
	authorization      AuthorizationFact
	idempotency        IdempotencyInput
	requestFingerprint Digest
}

// NewApplyCandidate binds exact content and idempotency scope to a plan.
func NewApplyCandidate(transactionID string, plan Plan, content []ContentBlob, authorization AuthorizationFact, idempotency IdempotencyInput) (ApplyCandidate, error) {
	if !validPlanIdentity(transactionID) || !validPlanIdentity(idempotency.OperationKind) ||
		idempotency.PrincipalID != plan.AuthorizationScope().PrincipalID ||
		!validPlanIdentity(idempotency.Key) {
		return ApplyCandidate{}, fmt.Errorf("%w: transaction or idempotency coordinates are invalid", ErrApplyInvalid)
	}
	if authorization.PlanDigest != plan.Digest() || !validSHA256Digest(authorization.FactDigest) {
		return ApplyCandidate{}, fmt.Errorf("%w: authorization fact is not bound to the plan", ErrApplyInvalid)
	}
	cloned, err := validateAndCloneContent(plan, content)
	if err != nil {
		return ApplyCandidate{}, err
	}
	fingerprint, err := computeRequestFingerprint(plan, cloned, authorization, idempotency)
	if err != nil {
		return ApplyCandidate{}, err
	}
	return ApplyCandidate{
		transactionID: transactionID, plan: plan, content: cloned,
		authorization: authorization, idempotency: idempotency,
		requestFingerprint: fingerprint,
	}, nil
}

func validateAndCloneContent(plan Plan, content []ContentBlob) ([]ContentBlob, error) {
	required := make(map[Digest]ContentRef)
	for _, effect := range plan.Effects() {
		if effect.Content != nil {
			required[effect.Content.BlobDigest] = *effect.Content
		}
	}
	if len(content) != len(required) {
		return nil, fmt.Errorf("%w: content set is incomplete or contains extras", ErrApplyInvalid)
	}
	result := make([]ContentBlob, 0, len(content))
	seen := make(map[Digest]struct{}, len(content))
	for _, blob := range content {
		ref, ok := required[blob.Digest]
		if !ok {
			return nil, fmt.Errorf("%w: unreferenced content %q", ErrApplyInvalid, blob.Digest)
		}
		if _, duplicate := seen[blob.Digest]; duplicate {
			return nil, fmt.Errorf("%w: duplicate content %q", ErrApplyInvalid, blob.Digest)
		}
		if uint64(len(blob.Bytes)) != ref.Size || RawBlobDigest(blob.Bytes) != blob.Digest {
			return nil, fmt.Errorf("%w: content %q bytes do not match the plan", ErrApplyInvalid, blob.Digest)
		}
		seen[blob.Digest] = struct{}{}
		result = append(result, ContentBlob{Digest: blob.Digest, Bytes: slices.Clone(blob.Bytes)})
	}
	slices.SortFunc(result, func(left, right ContentBlob) int { return strings.Compare(string(left.Digest), string(right.Digest)) })
	return result, nil
}

func computeRequestFingerprint(plan Plan, content []ContentBlob, authorization AuthorizationFact, idempotency IdempotencyInput) (Digest, error) {
	digests := make([]Digest, 0, len(content))
	for _, blob := range content {
		digests = append(digests, blob.Digest)
	}
	canonical, err := canonicaljson.Marshal(struct {
		Schema              string   `json:"schema"`
		WorkspaceID         string   `json:"workspaceId"`
		OperationKind       string   `json:"operationKind"`
		PrincipalID         string   `json:"principalId"`
		PlanDigest          Digest   `json:"planDigest"`
		ContentDigests      []Digest `json:"contentDigests"`
		AuthorizationDigest Digest   `json:"authorizationDigest"`
	}{
		Schema: "sdd.workspace-request/1", WorkspaceID: plan.WorkspaceID(),
		OperationKind: idempotency.OperationKind, PrincipalID: idempotency.PrincipalID,
		PlanDigest: plan.Digest(), ContentDigests: digests, AuthorizationDigest: authorization.FactDigest,
	})
	if err != nil {
		return "", fmt.Errorf("%w: fingerprint: %w", ErrApplyInvalid, err)
	}
	return digestBytes("sdd.workspace-request/1", canonical), nil
}

func (c ApplyCandidate) TransactionID() string            { return c.transactionID }
func (c ApplyCandidate) Plan() Plan                       { return c.plan }
func (c ApplyCandidate) RequestFingerprint() Digest       { return c.requestFingerprint }
func (c ApplyCandidate) Authorization() AuthorizationFact { return c.authorization }
func (c ApplyCandidate) Idempotency() IdempotencyInput    { return c.idempotency }
func (c ApplyCandidate) Content() []ContentBlob {
	result := make([]ContentBlob, len(c.content))
	for index := range c.content {
		result[index] = ContentBlob{Digest: c.content[index].Digest, Bytes: slices.Clone(c.content[index].Bytes)}
	}
	return result
}

// EffectObservation is supplied from the exact path check immediately before
// one prepared effect. "after" is valid only for recovery of this intent.
type EffectObservation string

const (
	EffectMatchesBefore EffectObservation = "matches-before"
	EffectMatchesAfter  EffectObservation = "matches-after"
	EffectConflicts     EffectObservation = "conflicts"
)

// PrepareFacts is the collectively recaptured guard set at the last reversible
// boundary. Equal generation alone is insufficient; the read-set digest binds
// all guarded raw observations and revisions.
type PrepareFacts struct {
	Generation    uint64
	ReadSetDigest Digest
	Now           time.Time
}

// ApplyActionKind is a closed instruction to the future persistence adapter.
type ApplyActionKind string

const (
	ApplyBegin           ApplyActionKind = "begin"
	ApplyBeginStaging    ApplyActionKind = "begin-staging"
	ApplyStageBlob       ApplyActionKind = "stage-blob"
	ApplyPrepare         ApplyActionKind = "prepare"
	ApplyInstallEffect   ApplyActionKind = "install-effect"
	ApplyRecognizeEffect ApplyActionKind = "recognize-effect"
	ApplyCommit          ApplyActionKind = "commit"
	ApplyReplay          ApplyActionKind = "replay"
	ApplyFailPreIntent   ApplyActionKind = "fail-pre-intent"
	ApplyQuarantine      ApplyActionKind = "quarantine"
)

// ApplyAction contains only the next durable primitive and stable references.
type ApplyAction struct {
	Kind        ApplyActionKind
	BlobDigest  Digest
	EffectIndex uint32
	Reason      string
	checked     bool
}

// ApplyRecord is immutable orchestration state. A persistence adapter stores
// its equivalent durable representation after acknowledging each action.
type ApplyRecord struct {
	candidate      ApplyCandidate
	state          TransactionState
	staged         []Digest
	applied        []uint32
	terminalReason string
}

func NewApplyRecord(candidate ApplyCandidate) ApplyRecord {
	return ApplyRecord{candidate: candidate, state: TransactionAbsent}
}

func (r ApplyRecord) State() TransactionState   { return r.state }
func (r ApplyRecord) Candidate() ApplyCandidate { return r.candidate }
func (r ApplyRecord) StagedDigests() []Digest   { return slices.Clone(r.staged) }
func (r ApplyRecord) AppliedEffects() []uint32  { return slices.Clone(r.applied) }
func (r ApplyRecord) TerminalReason() string    { return r.terminalReason }

// NextApplyAction deterministically chooses one next persistence primitive.
// effectObservation is required only while a prepared effect is pending.
func NextApplyAction(record ApplyRecord, prepare *PrepareFacts, effectObservation EffectObservation) (ApplyAction, error) {
	switch record.state {
	case TransactionAbsent:
		return ApplyAction{Kind: ApplyBegin}, nil
	case TransactionPending:
		return ApplyAction{Kind: ApplyBeginStaging}, nil
	case TransactionStaging:
		for _, blob := range record.candidate.content {
			if !containsDigest(record.staged, blob.Digest) {
				return ApplyAction{Kind: ApplyStageBlob, BlobDigest: blob.Digest}, nil
			}
		}
		if prepare == nil {
			return ApplyAction{}, fmt.Errorf("%w: prepare facts are required", ErrApplyInvalid)
		}
		plan := record.candidate.plan
		if prepare.Generation != plan.SourceGeneration() || prepare.ReadSetDigest != plan.ReadSetDigest() {
			return ApplyAction{Kind: ApplyFailPreIntent, Reason: "precondition-changed", checked: true}, nil
		}
		if err := VerifyAuthorization(plan, record.candidate.authorization, prepare.Now); err != nil {
			reason := "authorization-invalid"
			if errors.Is(err, ErrAuthorizationRevoked) {
				reason = "authorization-revoked"
			}
			return ApplyAction{Kind: ApplyFailPreIntent, Reason: reason, checked: true}, nil
		}
		return ApplyAction{Kind: ApplyPrepare, checked: true}, nil
	case TransactionPrepared:
		effects := record.candidate.plan.Effects()
		for _, effect := range effects {
			if containsEffect(record.applied, effect.Index) {
				continue
			}
			switch effectObservation {
			case EffectMatchesBefore:
				return ApplyAction{Kind: ApplyInstallEffect, EffectIndex: effect.Index, checked: true}, nil
			case EffectMatchesAfter:
				return ApplyAction{Kind: ApplyRecognizeEffect, EffectIndex: effect.Index, checked: true}, nil
			case EffectConflicts:
				return ApplyAction{Kind: ApplyQuarantine, EffectIndex: effect.Index, Reason: "prepared-precondition-changed", checked: true}, nil
			default:
				return ApplyAction{}, fmt.Errorf("%w: effect observation is required", ErrApplyInvalid)
			}
		}
		return ApplyAction{Kind: ApplyCommit}, nil
	case TransactionCommitted:
		return ApplyAction{Kind: ApplyReplay}, nil
	case TransactionFailed, TransactionQuarantined:
		return ApplyAction{}, fmt.Errorf("%w: transaction is terminal in %s", ErrApplyInvalid, record.state)
	default:
		return ApplyAction{}, fmt.Errorf("%w: unknown transaction state %q", ErrApplyInvalid, record.state)
	}
}

// AcknowledgeApplyAction returns the state that may be persisted only after the
// requested primitive succeeded durably. It refuses skipped/out-of-order work.
func AcknowledgeApplyAction(record ApplyRecord, action ApplyAction) (ApplyRecord, error) {
	next := cloneApplyRecord(record)
	switch action.Kind {
	case ApplyBegin:
		if record.state != TransactionAbsent {
			return record, invalidApplyAction(record, action)
		}
		next.state = TransactionPending
	case ApplyBeginStaging:
		if record.state != TransactionPending {
			return record, invalidApplyAction(record, action)
		}
		next.state = TransactionStaging
	case ApplyStageBlob:
		if record.state != TransactionStaging || !candidateHasContent(record.candidate, action.BlobDigest) || containsDigest(record.staged, action.BlobDigest) {
			return record, invalidApplyAction(record, action)
		}
		next.staged = append(next.staged, action.BlobDigest)
		slices.Sort(next.staged)
	case ApplyPrepare:
		if !action.checked || record.state != TransactionStaging || len(record.staged) != len(record.candidate.content) {
			return record, invalidApplyAction(record, action)
		}
		next.state = TransactionPrepared
	case ApplyInstallEffect, ApplyRecognizeEffect:
		if !action.checked || record.state != TransactionPrepared || !isFirstPendingEffect(record, action.EffectIndex) {
			return record, invalidApplyAction(record, action)
		}
		next.applied = append(next.applied, action.EffectIndex)
	case ApplyCommit:
		if record.state != TransactionPrepared || len(record.applied) != len(record.candidate.plan.Effects()) {
			return record, invalidApplyAction(record, action)
		}
		next.state = TransactionCommitted
	case ApplyReplay:
		if record.state != TransactionCommitted {
			return record, invalidApplyAction(record, action)
		}
	case ApplyFailPreIntent:
		if !action.checked || (record.state != TransactionPending && record.state != TransactionStaging) {
			return record, invalidApplyAction(record, action)
		}
		next.state = TransactionFailed
		next.terminalReason = action.Reason
	case ApplyQuarantine:
		if !action.checked || record.state != TransactionPrepared {
			return record, invalidApplyAction(record, action)
		}
		next.state = TransactionQuarantined
		next.terminalReason = action.Reason
	default:
		return record, invalidApplyAction(record, action)
	}
	return next, nil
}

func cloneApplyRecord(record ApplyRecord) ApplyRecord {
	record.staged = slices.Clone(record.staged)
	record.applied = slices.Clone(record.applied)
	return record
}

func invalidApplyAction(record ApplyRecord, action ApplyAction) error {
	return fmt.Errorf("%w: %s in %s", ErrApplyActionInvalid, action.Kind, record.state)
}

func candidateHasContent(candidate ApplyCandidate, digest Digest) bool {
	for _, blob := range candidate.content {
		if blob.Digest == digest {
			return true
		}
	}
	return false
}

func containsDigest(values []Digest, wanted Digest) bool { return slices.Contains(values, wanted) }
func containsEffect(values []uint32, wanted uint32) bool { return slices.Contains(values, wanted) }

func isFirstPendingEffect(record ApplyRecord, wanted uint32) bool {
	for _, effect := range record.candidate.plan.Effects() {
		if !containsEffect(record.applied, effect.Index) {
			return effect.Index == wanted
		}
	}
	return false
}
