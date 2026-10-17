package workspace

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"sdd.dev/sdd/internal/protocol/canonicaljson"
)

var (
	ErrPlanInvalid          = errors.New("workspace mutation plan invalid")
	ErrAuthorizationInvalid = errors.New("workspace authorization invalid")
	ErrAuthorizationRevoked = errors.New("workspace authorization revoked")
)

const (
	planSchema         = "sdd.workspace-plan-core/1"
	effectVocabularyV1 = "sdd.foundation-effects/1"
)

// EffectKind is the closed effect set supported by the first transaction
// walking slice. Extension effects are allocated but are not executable here.
type EffectKind string

const (
	EffectCreateFile EffectKind = "create-file"
	EffectUpdateFile EffectKind = "update-file"
	EffectDeleteFile EffectKind = "delete-file"
)

// RevisionScope names one independently monotonic authority coordinate.
type RevisionScope string

const (
	RevisionWorkspace RevisionScope = "workspace"
	RevisionWork      RevisionScope = "work"
	RevisionArtifact  RevisionScope = "artifact"
)

// ExpectedPath binds an effect to the exact raw observation it may replace.
type ExpectedPath struct {
	Presence   Presence           `json:"presence"`
	Kind       FileKind           `json:"kind"`
	BlobDigest Digest             `json:"blobDigest,omitempty"`
	Size       uint64             `json:"size"`
	Mode       uint32             `json:"mode"`
	NativeID   NativeFileIdentity `json:"nativeId"`
}

// ContentRef identifies bytes supplied separately to Apply. Plan identity
// never depends on a temporary staging filename.
type ContentRef struct {
	BlobDigest Digest `json:"blobDigest"`
	Size       uint64 `json:"size"`
	Mode       uint32 `json:"mode"`
}

// Effect is one exact ordered create/update/delete operation.
type Effect struct {
	Index   uint32        `json:"index"`
	Kind    EffectKind    `json:"kind"`
	Path    WorkspacePath `json:"path"`
	Before  ExpectedPath  `json:"before"`
	Content *ContentRef   `json:"content,omitempty"`
}

// RevisionIntent freezes both sides of a revision allocation before PREPARED.
type RevisionIntent struct {
	Scope    RevisionScope `json:"scope"`
	OwnerID  string        `json:"ownerId"`
	Expected uint64        `json:"expected"`
	Proposed uint64        `json:"proposed"`
}

// PlanDescriptor versions every rule that participates in plan identity.
type PlanDescriptor struct {
	SchemaVersion           string    `json:"schemaVersion"`
	EffectVocabularyVersion string    `json:"effectVocabularyVersion"`
	CanonicalizationVersion string    `json:"canonicalizationVersion"`
	CreatedAt               time.Time `json:"createdAt"`
	FreshUntil              time.Time `json:"freshUntil"`
}

// NewPlanDescriptor returns the only descriptor implemented by this package.
// Public protocol publication is a separate foundation-owned gate.
func NewPlanDescriptor(createdAt, freshUntil time.Time) PlanDescriptor {
	return PlanDescriptor{
		SchemaVersion:           planSchema,
		EffectVocabularyVersion: effectVocabularyV1,
		CanonicalizationVersion: "foundation-canonical-json/1",
		CreatedAt:               createdAt.UTC(),
		FreshUntil:              freshUntil.UTC(),
	}
}

// AuthorizationScope is the non-circular policy binding hashed into the plan.
// The provider's final grant ID/proof is deliberately absent.
type AuthorizationScope struct {
	ProviderID    string `json:"providerId"`
	PrincipalID   string `json:"principalId"`
	Scope         string `json:"scope"`
	PolicyDigest  Digest `json:"policyDigest"`
	PolicyVersion string `json:"policyVersion"`
	GrantKind     string `json:"grantKind"`
}

// PlanLimits bounds the mutation before any durable transaction begins.
type PlanLimits struct {
	MaxEffects     uint32 `json:"maxEffects"`
	MaxStagedBytes uint64 `json:"maxStagedBytes"`
}

// PlanInput contains all decision-bearing data except the computed digest and
// final authorization fact. Slices are copied by NewPlan.
type PlanInput struct {
	Descriptor       PlanDescriptor
	PlanID           string
	WorkspaceID      string
	WorkID           string
	SourceSnapshotID string
	ReadSetDigest    Digest
	SourceGeneration uint64
	Authorization    AuthorizationScope
	Effects          []Effect
	RevisionIntents  []RevisionIntent
	Limits           PlanLimits
}

// Plan is an immutable validated exact mutation plan.
type Plan struct {
	core   planCore
	digest Digest
}

type planCore struct {
	Descriptor       PlanDescriptor     `json:"descriptor"`
	PlanID           string             `json:"planId"`
	WorkspaceID      string             `json:"workspaceId"`
	WorkID           string             `json:"workId"`
	SourceSnapshotID string             `json:"sourceSnapshotId"`
	ReadSetDigest    Digest             `json:"readSetDigest"`
	SourceGeneration uint64             `json:"sourceGeneration"`
	Authorization    AuthorizationScope `json:"authorization"`
	Effects          []Effect           `json:"effects"`
	RevisionIntents  []RevisionIntent   `json:"revisionIntents"`
	Limits           PlanLimits         `json:"limits"`
}

// NewPlan validates, copies, canonicalizes, and hashes one exact plan.
func NewPlan(input PlanInput) (Plan, error) {
	core := planCore{
		Descriptor:       normalizeDescriptor(input.Descriptor),
		PlanID:           input.PlanID,
		WorkspaceID:      input.WorkspaceID,
		WorkID:           input.WorkID,
		SourceSnapshotID: input.SourceSnapshotID,
		ReadSetDigest:    input.ReadSetDigest,
		SourceGeneration: input.SourceGeneration,
		Authorization:    input.Authorization,
		Effects:          cloneEffects(input.Effects),
		RevisionIntents:  slices.Clone(input.RevisionIntents),
		Limits:           input.Limits,
	}
	if err := validatePlanCore(core); err != nil {
		return Plan{}, err
	}
	canonical, err := canonicaljson.Marshal(core)
	if err != nil {
		return Plan{}, fmt.Errorf("%w: canonicalize: %w", ErrPlanInvalid, err)
	}
	digest := digestCanonicalPlan(canonical)
	return Plan{core: core, digest: digest}, nil
}

func normalizeDescriptor(value PlanDescriptor) PlanDescriptor {
	value.CreatedAt = value.CreatedAt.UTC()
	value.FreshUntil = value.FreshUntil.UTC()
	return value
}

func validatePlanCore(core planCore) error {
	identities := []struct{ name, value string }{
		{"plan", core.PlanID}, {"workspace", core.WorkspaceID}, {"work", core.WorkID},
		{"snapshot", core.SourceSnapshotID}, {"provider", core.Authorization.ProviderID},
		{"principal", core.Authorization.PrincipalID}, {"authorization scope", core.Authorization.Scope},
		{"policy version", core.Authorization.PolicyVersion}, {"grant kind", core.Authorization.GrantKind},
	}
	for _, identity := range identities {
		if !validPlanIdentity(identity.value) {
			return fmt.Errorf("%w: %s identity is empty, padded, or unsafe", ErrPlanInvalid, identity.name)
		}
	}
	if core.Descriptor.SchemaVersion != planSchema || core.Descriptor.EffectVocabularyVersion != effectVocabularyV1 || core.Descriptor.CanonicalizationVersion != "foundation-canonical-json/1" {
		return fmt.Errorf("%w: unsupported plan descriptor", ErrPlanInvalid)
	}
	if core.Descriptor.CreatedAt.IsZero() || !core.Descriptor.FreshUntil.After(core.Descriptor.CreatedAt) {
		return fmt.Errorf("%w: invalid plan freshness interval", ErrPlanInvalid)
	}
	if !validSHA256Digest(core.ReadSetDigest) || !validSHA256Digest(core.Authorization.PolicyDigest) {
		return fmt.Errorf("%w: source or policy digest is malformed", ErrPlanInvalid)
	}
	if core.Limits.MaxEffects == 0 || core.Limits.MaxStagedBytes == 0 {
		return fmt.Errorf("%w: positive effect and staging limits are required", ErrPlanInvalid)
	}
	if len(core.Effects) == 0 || uint64(len(core.Effects)) > uint64(core.Limits.MaxEffects) {
		return fmt.Errorf("%w: effect count is empty or exceeds its limit", ErrPlanInvalid)
	}
	if len(core.RevisionIntents) == 0 {
		return fmt.Errorf("%w: at least one revision intent is required", ErrPlanInvalid)
	}
	if err := validateEffects(core.Effects, core.Limits.MaxStagedBytes); err != nil {
		return err
	}
	return validateRevisionIntents(core.RevisionIntents)
}

func validateEffects(effects []Effect, maxStaged uint64) error {
	paths := make(map[WorkspacePath]struct{}, len(effects))
	var staged uint64
	for index, effect := range effects {
		if uint64(effect.Index) != uint64(index) {
			return fmt.Errorf("%w: effect indexes must be contiguous from zero", ErrPlanInvalid)
		}
		parsed, err := ParseWorkspacePath(string(effect.Path))
		if err != nil || parsed != effect.Path {
			return fmt.Errorf("%w: effect %d path is not canonical", ErrPlanInvalid, effect.Index)
		}
		if _, exists := paths[effect.Path]; exists {
			return fmt.Errorf("%w: path %q is affected more than once", ErrPlanInvalid, effect.Path)
		}
		paths[effect.Path] = struct{}{}
		switch effect.Kind {
		case EffectCreateFile:
			if effect.Before.Presence != Absent || effect.Before.Kind != FileKindNone || effect.Before.BlobDigest != "" || effect.Before.Size != 0 || effect.Before.Mode != 0 || effect.Before.NativeID != (NativeFileIdentity{}) || effect.Content == nil {
				return fmt.Errorf("%w: invalid create effect %d", ErrPlanInvalid, effect.Index)
			}
		case EffectUpdateFile:
			if effect.Before.Presence != Present || effect.Before.Kind != FileKindRegular || !validSHA256Digest(effect.Before.BlobDigest) || effect.Content == nil {
				return fmt.Errorf("%w: invalid update effect %d", ErrPlanInvalid, effect.Index)
			}
		case EffectDeleteFile:
			if effect.Before.Presence != Present || effect.Before.Kind != FileKindRegular || !validSHA256Digest(effect.Before.BlobDigest) || effect.Content != nil {
				return fmt.Errorf("%w: invalid delete effect %d", ErrPlanInvalid, effect.Index)
			}
		default:
			return fmt.Errorf("%w: unsupported effect kind %q", ErrPlanInvalid, effect.Kind)
		}
		if effect.Content != nil {
			if !validSHA256Digest(effect.Content.BlobDigest) {
				return fmt.Errorf("%w: effect %d content digest is malformed", ErrPlanInvalid, effect.Index)
			}
			if math.MaxUint64-staged < effect.Content.Size {
				return fmt.Errorf("%w: staged byte total overflow", ErrPlanInvalid)
			}
			staged += effect.Content.Size
			if staged > maxStaged {
				return fmt.Errorf("%w: staged byte limit exceeded", ErrLimitExceeded)
			}
		}
	}
	return nil
}

func validPlanIdentity(value string) bool {
	if strings.TrimSpace(value) == "" || value != strings.TrimSpace(value) {
		return false
	}
	for _, character := range value {
		if character < 0x20 || character == 0x7f {
			return false
		}
	}
	return true
}

func validateRevisionIntents(intents []RevisionIntent) error {
	seen := make(map[string]struct{}, len(intents))
	for _, intent := range intents {
		if intent.Scope != RevisionWorkspace && intent.Scope != RevisionWork && intent.Scope != RevisionArtifact {
			return fmt.Errorf("%w: unsupported revision scope %q", ErrPlanInvalid, intent.Scope)
		}
		if strings.TrimSpace(intent.OwnerID) == "" || intent.OwnerID != strings.TrimSpace(intent.OwnerID) {
			return fmt.Errorf("%w: revision owner is empty or padded", ErrPlanInvalid)
		}
		if intent.Expected == math.MaxUint64 || intent.Proposed != intent.Expected+1 {
			return fmt.Errorf("%w: revision for %s/%s must advance exactly once", ErrPlanInvalid, intent.Scope, intent.OwnerID)
		}
		key := string(intent.Scope) + "\x00" + intent.OwnerID
		if _, exists := seen[key]; exists {
			return fmt.Errorf("%w: duplicate revision intent for %s/%s", ErrPlanInvalid, intent.Scope, intent.OwnerID)
		}
		seen[key] = struct{}{}
	}
	return nil
}

func digestCanonicalPlan(canonical []byte) Digest {
	hash := sha256.New()
	writeFramed(hash, planSchema)
	writeUint64(hash, uint64(len(canonical)))
	_, _ = hash.Write(canonical)
	return Digest("sha256:" + hex.EncodeToString(hash.Sum(nil)))
}

func validSHA256Digest(value Digest) bool {
	text := string(value)
	if !strings.HasPrefix(text, "sha256:") || len(text) != len("sha256:")+sha256.Size*2 {
		return false
	}
	decoded, err := hex.DecodeString(strings.TrimPrefix(text, "sha256:"))
	return err == nil && len(decoded) == sha256.Size && strings.ToLower(text) == text
}

func cloneEffects(source []Effect) []Effect {
	result := slices.Clone(source)
	for index := range result {
		if source[index].Content != nil {
			content := *source[index].Content
			result[index].Content = &content
		}
	}
	return result
}

// Digest reports the exact plan identity to which authorization must bind.
func (p Plan) Digest() Digest { return p.digest }

func (p Plan) PlanID() string                         { return p.core.PlanID }
func (p Plan) WorkspaceID() string                    { return p.core.WorkspaceID }
func (p Plan) WorkID() string                         { return p.core.WorkID }
func (p Plan) SourceSnapshotID() string               { return p.core.SourceSnapshotID }
func (p Plan) ReadSetDigest() Digest                  { return p.core.ReadSetDigest }
func (p Plan) SourceGeneration() uint64               { return p.core.SourceGeneration }
func (p Plan) Descriptor() PlanDescriptor             { return p.core.Descriptor }
func (p Plan) AuthorizationScope() AuthorizationScope { return p.core.Authorization }
func (p Plan) Limits() PlanLimits                     { return p.core.Limits }
func (p Plan) Effects() []Effect                      { return cloneEffects(p.core.Effects) }
func (p Plan) RevisionIntents() []RevisionIntent      { return slices.Clone(p.core.RevisionIntents) }

// MarshalJSON exposes the typed core plus digest without creating a second
// protocol envelope. Public publication remains a foundation-owned gate.
func (p Plan) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		planCore
		Digest Digest `json:"digest"`
	}{planCore: p.core, Digest: p.digest})
}

// AuthorizationFact is the verified provider claim compared to one plan.
// Proof decoding/authenticity belongs to the injected provider adapter.
type AuthorizationFact struct {
	ProviderID    string
	GrantID       string
	PrincipalID   string
	Scope         string
	PlanDigest    Digest
	PolicyDigest  Digest
	PolicyVersion string
	GrantKind     string
	IssuedAt      time.Time
	NotBefore     time.Time
	ExpiresAt     time.Time
	Revoked       bool
	RevocationID  string
}

// VerifyAuthorization checks binding and freshness immediately before
// PREPARED. It does not evaluate artifact or workflow policy.
func VerifyAuthorization(plan Plan, fact AuthorizationFact, now time.Time) error {
	scope := plan.AuthorizationScope()
	if fact.ProviderID != scope.ProviderID || fact.PrincipalID != scope.PrincipalID || fact.Scope != scope.Scope ||
		fact.PlanDigest != plan.Digest() || fact.PolicyDigest != scope.PolicyDigest || fact.PolicyVersion != scope.PolicyVersion ||
		fact.GrantKind != scope.GrantKind || !validPlanIdentity(fact.GrantID) {
		return fmt.Errorf("%w: grant does not bind the exact plan and scope", ErrAuthorizationInvalid)
	}
	if fact.Revoked {
		return fmt.Errorf("%w: %s", ErrAuthorizationRevoked, fact.RevocationID)
	}
	now = now.UTC()
	descriptor := plan.Descriptor()
	if fact.IssuedAt.IsZero() || fact.NotBefore.IsZero() || fact.ExpiresAt.IsZero() || !fact.ExpiresAt.After(fact.NotBefore) ||
		now.Before(fact.NotBefore.UTC()) || !now.Before(fact.ExpiresAt.UTC()) || fact.IssuedAt.UTC().After(now) ||
		now.Before(descriptor.CreatedAt) || !now.Before(descriptor.FreshUntil) {
		return fmt.Errorf("%w: grant or plan is outside its freshness interval", ErrAuthorizationInvalid)
	}
	return nil
}
