// Package workspace owns coherent raw workspace observations and, in later
// slices, exact transaction application and recovery.
package workspace

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"path"
	"slices"
	"strings"
)

var (
	ErrSnapshotInvalid  = errors.New("workspace snapshot invalid")
	ErrSnapshotUnstable = errors.New("workspace snapshot unstable")
	ErrRecoveryRequired = errors.New("workspace recovery required")
	ErrLimitExceeded    = errors.New("workspace snapshot limit exceeded")
)

// SnapshotScope selects one closed, implementation-owned read-set recipe.
type SnapshotScope string

const (
	ScopeSelection          SnapshotScope = "selection"
	ScopeWorkCurrent        SnapshotScope = "work-current"
	ScopeArtifactIdentities SnapshotScope = "artifact-identities"
	ScopePlanInputs         SnapshotScope = "plan-inputs"
	ScopeRecoveryStatus     SnapshotScope = "recovery-status"
)

// Presence distinguishes an observed absent path from a zero-valued present
// file. It is part of raw snapshot identity.
type Presence string

const (
	Absent  Presence = "absent"
	Present Presence = "present"
)

// FileKind is deliberately about raw filesystem facts, not artifact meaning.
type FileKind string

const (
	FileKindNone      FileKind = "none"
	FileKindRegular   FileKind = "regular"
	FileKindDirectory FileKind = "directory"
)

// ArtifactIdentityState is a durable workspace observation. Proposed IDs do
// not appear here until an authorized transaction allocates them.
type ArtifactIdentityState string

const (
	ArtifactAllocated     ArtifactIdentityState = "allocated"
	ArtifactTombstoned    ArtifactIdentityState = "tombstoned"
	ArtifactMissingRecord ArtifactIdentityState = "missing-record"
)

// Digest is a domain-separated content identity encoded as algorithm:hex.
type Digest string

// WorkspacePath is normalized POSIX text relative to the workspace root.
type WorkspacePath string

// ParseWorkspacePath rejects ambiguous or escaping representations. Platform
// reserved names, comparison keys, and native length limits belong to the
// native filesystem adapter and are checked again at effect time.
func ParseWorkspacePath(raw string) (WorkspacePath, error) {
	if raw == "" || strings.HasPrefix(raw, "/") || strings.Contains(raw, "\\") {
		return "", fmt.Errorf("%w: path %q is not normalized workspace-relative POSIX text", ErrSnapshotInvalid, raw)
	}
	cleaned := path.Clean(raw)
	if cleaned != raw || cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.Contains(raw, "//") {
		return "", fmt.Errorf("%w: path %q contains an empty, dot, or escaping component", ErrSnapshotInvalid, raw)
	}
	for _, component := range strings.Split(raw, "/") {
		if component == "" || component == "." || component == ".." {
			return "", fmt.Errorf("%w: path %q contains an invalid component", ErrSnapshotInvalid, raw)
		}
	}
	return WorkspacePath(raw), nil
}

type Selection struct {
	ExplicitRoot string
	WorkspaceID  string
	WorkID       string
	WorkSlug     string
}

type Limits struct {
	MaxPaths          int
	MaxBytes          uint64
	StableReadRetries int
}

type SnapshotRequest struct {
	Selection Selection
	Scope     SnapshotScope
	Paths     []WorkspacePath
	Limits    Limits
}

type WorkspaceObservation struct {
	WorkspaceID   string
	CanonicalRoot string
	FormatVersion int
	Revision      uint64
}

type WorkObservation struct {
	WorkID   string
	Slug     string
	Revision uint64
}

type NativeFileIdentity struct {
	Device uint64
	File   uint64
	Opaque string
}

type PathObservation struct {
	Path       WorkspacePath
	Presence   Presence
	Kind       FileKind
	Bytes      []byte
	BlobDigest Digest
	Size       uint64
	Mode       uint32
	NativeID   NativeFileIdentity
	LinkCount  *uint64
	CompareKey string
}

type ProvenanceRef struct {
	Kind   string
	ID     string
	Digest Digest
}

type ArtifactIdentityObservation struct {
	WorkspaceID             string
	WorkID                  string
	Namespace               string
	ArtifactID              string
	State                   ArtifactIdentityState
	AllocationRevision      *uint64
	AllocationProvenance    *ProvenanceRef
	CurrentArtifactRevision *uint64
	CurrentPath             *WorkspacePath
	RemovalProvenance       *ProvenanceRef
}

// Capture is the platform adapter's bounded coherent observation. The module
// accepts it only after checking request coverage and generation coherence.
type Capture struct {
	SnapshotID          string
	Workspace           WorkspaceObservation
	Work                *WorkObservation
	GenerationBefore    uint64
	GenerationAfter     uint64
	ActiveTransactionID string
	Paths               []PathObservation
	ArtifactIdentities  []ArtifactIdentityObservation
}

// Snapshot is the accepted RawSnapshot bridge consumed by later pure modules.
// Its slices and byte buffers never alias adapter-owned buffers.
type Snapshot struct {
	SnapshotID         string
	Workspace          WorkspaceObservation
	Work               *WorkObservation
	Scope              SnapshotScope
	Generation         uint64
	ReadSetDigest      Digest
	Paths              []PathObservation
	ArtifactIdentities []ArtifactIdentityObservation
	LimitsApplied      Limits
}

// Capturer is the workspace-owned platform seam. A native adapter and a
// deterministic fault adapter are the two intended implementations.
type Capturer interface {
	Capture(context.Context, SnapshotRequest) (Capture, error)
}

// Reader hides capture validation, ordering, ownership, limit, and digest
// rules behind one read-only operation.
type Reader struct {
	capturer Capturer
}

func NewReader(capturer Capturer) (*Reader, error) {
	if capturer == nil {
		return nil, fmt.Errorf("%w: capturer is required", ErrSnapshotInvalid)
	}
	return &Reader{capturer: capturer}, nil
}

func (r *Reader) Snapshot(ctx context.Context, request SnapshotRequest) (Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return Snapshot{}, err
	}
	normalized, err := normalizeRequest(request)
	if err != nil {
		return Snapshot{}, err
	}
	capture, err := r.capturer.Capture(ctx, normalized)
	if err != nil {
		return Snapshot{}, fmt.Errorf("capture raw workspace snapshot: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return Snapshot{}, err
	}
	return acceptCapture(normalized, capture)
}

func normalizeRequest(request SnapshotRequest) (SnapshotRequest, error) {
	if request.Selection.WorkspaceID == "" {
		return SnapshotRequest{}, fmt.Errorf("%w: workspace ID is required", ErrSnapshotInvalid)
	}
	if request.Selection.WorkID != "" && request.Selection.WorkSlug != "" {
		return SnapshotRequest{}, fmt.Errorf("%w: select work by immutable ID or slug, not both", ErrSnapshotInvalid)
	}
	if !validScope(request.Scope) {
		return SnapshotRequest{}, fmt.Errorf("%w: unsupported snapshot scope %q", ErrSnapshotInvalid, request.Scope)
	}
	if request.Limits.MaxPaths <= 0 || request.Limits.MaxBytes == 0 || request.Limits.StableReadRetries < 0 {
		return SnapshotRequest{}, fmt.Errorf("%w: positive path/byte limits and a non-negative retry bound are required", ErrSnapshotInvalid)
	}
	if len(request.Paths) > request.Limits.MaxPaths {
		return SnapshotRequest{}, fmt.Errorf("%w: %d paths exceed limit %d", ErrLimitExceeded, len(request.Paths), request.Limits.MaxPaths)
	}

	normalized := request
	normalized.Paths = slices.Clone(request.Paths)
	for _, item := range normalized.Paths {
		parsed, err := ParseWorkspacePath(string(item))
		if err != nil || parsed != item {
			if err != nil {
				return SnapshotRequest{}, err
			}
			return SnapshotRequest{}, fmt.Errorf("%w: path %q is not canonical", ErrSnapshotInvalid, item)
		}
	}
	slices.Sort(normalized.Paths)
	for i := 1; i < len(normalized.Paths); i++ {
		if normalized.Paths[i] == normalized.Paths[i-1] {
			return SnapshotRequest{}, fmt.Errorf("%w: duplicate path %q", ErrSnapshotInvalid, normalized.Paths[i])
		}
	}
	return normalized, nil
}

func acceptCapture(request SnapshotRequest, capture Capture) (Snapshot, error) {
	if capture.ActiveTransactionID != "" {
		return Snapshot{}, fmt.Errorf("%w: active transaction %q", ErrRecoveryRequired, capture.ActiveTransactionID)
	}
	if capture.GenerationBefore != capture.GenerationAfter {
		return Snapshot{}, fmt.Errorf("%w: generation changed from %d to %d", ErrSnapshotUnstable, capture.GenerationBefore, capture.GenerationAfter)
	}
	if capture.SnapshotID == "" || capture.Workspace.WorkspaceID != request.Selection.WorkspaceID {
		return Snapshot{}, fmt.Errorf("%w: capture identity does not match selection", ErrSnapshotInvalid)
	}
	if capture.Workspace.FormatVersion < 1 || capture.Workspace.CanonicalRoot == "" {
		return Snapshot{}, fmt.Errorf("%w: workspace observation is incomplete", ErrSnapshotInvalid)
	}
	if err := validateSelectedWork(request.Selection, capture.Work); err != nil {
		return Snapshot{}, err
	}

	paths := clonePaths(capture.Paths)
	slices.SortFunc(paths, func(a, b PathObservation) int { return strings.Compare(string(a.Path), string(b.Path)) })
	if len(paths) != len(request.Paths) {
		return Snapshot{}, fmt.Errorf("%w: capture returned %d paths for a %d-path read set", ErrSnapshotInvalid, len(paths), len(request.Paths))
	}
	var totalBytes uint64
	for i := range paths {
		if paths[i].Path != request.Paths[i] {
			return Snapshot{}, fmt.Errorf("%w: capture path %q does not match requested path %q", ErrSnapshotInvalid, paths[i].Path, request.Paths[i])
		}
		if err := normalizeObservation(&paths[i]); err != nil {
			return Snapshot{}, err
		}
		if ^uint64(0)-totalBytes < paths[i].Size {
			return Snapshot{}, fmt.Errorf("%w: byte total overflow", ErrLimitExceeded)
		}
		totalBytes += paths[i].Size
		if totalBytes > request.Limits.MaxBytes {
			return Snapshot{}, fmt.Errorf("%w: %d bytes exceed limit %d", ErrLimitExceeded, totalBytes, request.Limits.MaxBytes)
		}
	}

	identities := cloneArtifactIdentities(capture.ArtifactIdentities)
	slices.SortFunc(identities, func(a, b ArtifactIdentityObservation) int {
		if byNamespace := strings.Compare(a.Namespace, b.Namespace); byNamespace != 0 {
			return byNamespace
		}
		return strings.Compare(a.ArtifactID, b.ArtifactID)
	})
	if request.Scope != ScopeArtifactIdentities && len(identities) != 0 {
		return Snapshot{}, fmt.Errorf("%w: artifact identities returned outside artifact-identity scope", ErrSnapshotInvalid)
	}
	for i := range identities {
		if i > 0 && identities[i-1].Namespace == identities[i].Namespace && identities[i-1].ArtifactID == identities[i].ArtifactID {
			return Snapshot{}, fmt.Errorf("%w: duplicate artifact identity %s/%s", ErrSnapshotInvalid, identities[i].Namespace, identities[i].ArtifactID)
		}
		if err := validateArtifactIdentity(capture.Workspace.WorkspaceID, capture.Work, identities[i]); err != nil {
			return Snapshot{}, err
		}
	}

	snapshot := Snapshot{
		SnapshotID:         capture.SnapshotID,
		Workspace:          capture.Workspace,
		Work:               cloneWork(capture.Work),
		Scope:              request.Scope,
		Generation:         capture.GenerationAfter,
		Paths:              paths,
		ArtifactIdentities: identities,
		LimitsApplied:      request.Limits,
	}
	snapshot.ReadSetDigest = digestReadSet(snapshot)
	return snapshot, nil
}

func validScope(scope SnapshotScope) bool {
	switch scope {
	case ScopeSelection, ScopeWorkCurrent, ScopeArtifactIdentities, ScopePlanInputs, ScopeRecoveryStatus:
		return true
	default:
		return false
	}
}

func validateSelectedWork(selection Selection, work *WorkObservation) error {
	if selection.WorkID == "" && selection.WorkSlug == "" {
		return nil
	}
	if work == nil {
		return fmt.Errorf("%w: selected work is absent", ErrSnapshotInvalid)
	}
	if selection.WorkID != "" && work.WorkID != selection.WorkID {
		return fmt.Errorf("%w: captured work ID %q does not match %q", ErrSnapshotInvalid, work.WorkID, selection.WorkID)
	}
	if selection.WorkSlug != "" && work.Slug != selection.WorkSlug {
		return fmt.Errorf("%w: captured work slug %q does not match %q", ErrSnapshotInvalid, work.Slug, selection.WorkSlug)
	}
	if work.WorkID == "" || work.Slug == "" {
		return fmt.Errorf("%w: selected work observation is incomplete", ErrSnapshotInvalid)
	}
	return nil
}

func normalizeObservation(observation *PathObservation) error {
	parsed, err := ParseWorkspacePath(string(observation.Path))
	if err != nil || parsed != observation.Path {
		if err != nil {
			return err
		}
		return fmt.Errorf("%w: observed path %q is not canonical", ErrSnapshotInvalid, observation.Path)
	}
	switch observation.Presence {
	case Absent:
		if observation.Kind != FileKindNone || len(observation.Bytes) != 0 || observation.Size != 0 || observation.BlobDigest != "" {
			return fmt.Errorf("%w: absent path %q carries present-file facts", ErrSnapshotInvalid, observation.Path)
		}
	case Present:
		if observation.Kind != FileKindRegular && observation.Kind != FileKindDirectory {
			return fmt.Errorf("%w: present path %q has unsupported kind %q", ErrSnapshotInvalid, observation.Path, observation.Kind)
		}
		if observation.Kind == FileKindDirectory && len(observation.Bytes) != 0 {
			return fmt.Errorf("%w: directory %q carries file bytes", ErrSnapshotInvalid, observation.Path)
		}
		if observation.Kind == FileKindRegular && uint64(len(observation.Bytes)) != observation.Size {
			return fmt.Errorf("%w: path %q size does not match bytes", ErrSnapshotInvalid, observation.Path)
		}
		want := digestBytes("sdd.raw-blob/1", observation.Bytes)
		if observation.BlobDigest != "" && observation.BlobDigest != want {
			return fmt.Errorf("%w: path %q blob digest mismatch", ErrSnapshotInvalid, observation.Path)
		}
		observation.BlobDigest = want
	default:
		return fmt.Errorf("%w: path %q has invalid presence %q", ErrSnapshotInvalid, observation.Path, observation.Presence)
	}
	return nil
}

func validateArtifactIdentity(workspaceID string, work *WorkObservation, identity ArtifactIdentityObservation) error {
	if identity.WorkspaceID != workspaceID || identity.WorkID == "" || identity.Namespace == "" || identity.ArtifactID == "" {
		return fmt.Errorf("%w: artifact identity coordinates are incomplete", ErrSnapshotInvalid)
	}
	if work != nil && identity.WorkID != work.WorkID {
		return fmt.Errorf("%w: artifact %q belongs to a different work item", ErrSnapshotInvalid, identity.ArtifactID)
	}
	switch identity.State {
	case ArtifactAllocated:
		if identity.AllocationRevision == nil || identity.AllocationProvenance == nil || identity.CurrentArtifactRevision == nil || identity.CurrentPath == nil || identity.RemovalProvenance != nil {
			return fmt.Errorf("%w: allocated artifact %q has incomplete or contradictory facts", ErrSnapshotInvalid, identity.ArtifactID)
		}
		if *identity.AllocationRevision == 0 || *identity.CurrentArtifactRevision == 0 {
			return fmt.Errorf("%w: allocated artifact %q uses a zero persisted revision", ErrSnapshotInvalid, identity.ArtifactID)
		}
		if err := validateProvenance(identity.ArtifactID, identity.AllocationProvenance); err != nil {
			return err
		}
	case ArtifactTombstoned:
		if identity.AllocationRevision == nil || identity.AllocationProvenance == nil || identity.RemovalProvenance == nil || identity.CurrentPath != nil {
			return fmt.Errorf("%w: tombstoned artifact %q has incomplete or contradictory facts", ErrSnapshotInvalid, identity.ArtifactID)
		}
		if *identity.AllocationRevision == 0 || (identity.CurrentArtifactRevision != nil && *identity.CurrentArtifactRevision == 0) {
			return fmt.Errorf("%w: tombstoned artifact %q uses a zero persisted revision", ErrSnapshotInvalid, identity.ArtifactID)
		}
		if err := validateProvenance(identity.ArtifactID, identity.AllocationProvenance); err != nil {
			return err
		}
		if err := validateProvenance(identity.ArtifactID, identity.RemovalProvenance); err != nil {
			return err
		}
	case ArtifactMissingRecord:
		if identity.AllocationRevision != nil || identity.AllocationProvenance != nil || identity.CurrentArtifactRevision != nil || identity.RemovalProvenance != nil {
			return fmt.Errorf("%w: missing-record artifact %q invents durable binding facts", ErrSnapshotInvalid, identity.ArtifactID)
		}
	default:
		return fmt.Errorf("%w: artifact %q has unknown state %q", ErrSnapshotInvalid, identity.ArtifactID, identity.State)
	}
	if identity.CurrentPath != nil {
		parsed, err := ParseWorkspacePath(string(*identity.CurrentPath))
		if err != nil || parsed != *identity.CurrentPath {
			if err != nil {
				return err
			}
			return fmt.Errorf("%w: artifact %q has noncanonical path", ErrSnapshotInvalid, identity.ArtifactID)
		}
	}
	return nil
}

func validateProvenance(artifactID string, provenance *ProvenanceRef) error {
	if provenance == nil || provenance.Kind == "" || provenance.ID == "" || provenance.Digest == "" {
		return fmt.Errorf("%w: artifact %q has incomplete provenance", ErrSnapshotInvalid, artifactID)
	}
	return nil
}

func digestReadSet(snapshot Snapshot) Digest {
	h := sha256.New()
	writeFramed(h, "sdd.raw-read-set/1")
	writeFramed(h, snapshot.Workspace.WorkspaceID)
	writeUint64(h, uint64(snapshot.Workspace.FormatVersion))
	writeUint64(h, snapshot.Workspace.Revision)
	if snapshot.Work != nil {
		writeFramed(h, snapshot.Work.WorkID)
		writeFramed(h, snapshot.Work.Slug)
		writeUint64(h, snapshot.Work.Revision)
	} else {
		writeFramed(h, "")
	}
	writeFramed(h, string(snapshot.Scope))
	writeUint64(h, snapshot.Generation)
	for _, observation := range snapshot.Paths {
		writeFramed(h, string(observation.Path))
		writeFramed(h, string(observation.Presence))
		writeFramed(h, string(observation.Kind))
		writeFramed(h, string(observation.BlobDigest))
		writeUint64(h, observation.Size)
		writeUint64(h, uint64(observation.Mode))
		writeUint64(h, observation.NativeID.Device)
		writeUint64(h, observation.NativeID.File)
		writeFramed(h, observation.NativeID.Opaque)
		writeOptionalUint64(h, observation.LinkCount)
	}
	for _, identity := range snapshot.ArtifactIdentities {
		writeFramed(h, identity.WorkspaceID)
		writeFramed(h, identity.WorkID)
		writeFramed(h, identity.Namespace)
		writeFramed(h, identity.ArtifactID)
		writeFramed(h, string(identity.State))
		writeOptionalUint64(h, identity.AllocationRevision)
		writeOptionalProvenance(h, identity.AllocationProvenance)
		writeOptionalUint64(h, identity.CurrentArtifactRevision)
		writeOptionalPath(h, identity.CurrentPath)
		writeOptionalProvenance(h, identity.RemovalProvenance)
	}
	return Digest("sha256:" + hex.EncodeToString(h.Sum(nil)))
}

type framedWriter interface {
	Write([]byte) (int, error)
}

func writeFramed(w framedWriter, value string) {
	writeUint64(w, uint64(len(value)))
	_, _ = w.Write([]byte(value))
}

func writeUint64(w framedWriter, value uint64) {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], value)
	_, _ = w.Write(b[:])
}

func writeOptionalUint64(w framedWriter, value *uint64) {
	if value == nil {
		writeFramed(w, "absent")
		return
	}
	writeFramed(w, "present")
	writeUint64(w, *value)
}

func writeOptionalPath(w framedWriter, value *WorkspacePath) {
	if value == nil {
		writeFramed(w, "absent")
		return
	}
	writeFramed(w, "present")
	writeFramed(w, string(*value))
}

func writeOptionalProvenance(w framedWriter, value *ProvenanceRef) {
	if value == nil {
		writeFramed(w, "absent")
		return
	}
	writeFramed(w, "present")
	writeFramed(w, value.Kind)
	writeFramed(w, value.ID)
	writeFramed(w, string(value.Digest))
}

func digestBytes(domain string, value []byte) Digest {
	h := sha256.New()
	writeFramed(h, domain)
	writeUint64(h, uint64(len(value)))
	_, _ = h.Write(value)
	return Digest("sha256:" + hex.EncodeToString(h.Sum(nil)))
}

func clonePaths(source []PathObservation) []PathObservation {
	result := make([]PathObservation, len(source))
	for i := range source {
		result[i] = source[i]
		result[i].Bytes = slices.Clone(source[i].Bytes)
		if source[i].LinkCount != nil {
			value := *source[i].LinkCount
			result[i].LinkCount = &value
		}
	}
	return result
}

func cloneArtifactIdentities(source []ArtifactIdentityObservation) []ArtifactIdentityObservation {
	result := make([]ArtifactIdentityObservation, len(source))
	for i := range source {
		result[i] = source[i]
		result[i].AllocationRevision = cloneUint64(source[i].AllocationRevision)
		result[i].CurrentArtifactRevision = cloneUint64(source[i].CurrentArtifactRevision)
		if source[i].CurrentPath != nil {
			value := *source[i].CurrentPath
			result[i].CurrentPath = &value
		}
		result[i].AllocationProvenance = cloneProvenance(source[i].AllocationProvenance)
		result[i].RemovalProvenance = cloneProvenance(source[i].RemovalProvenance)
	}
	return result
}

func cloneWork(source *WorkObservation) *WorkObservation {
	if source == nil {
		return nil
	}
	value := *source
	return &value
}

func cloneUint64(source *uint64) *uint64 {
	if source == nil {
		return nil
	}
	value := *source
	return &value
}

func cloneProvenance(source *ProvenanceRef) *ProvenanceRef {
	if source == nil {
		return nil
	}
	value := *source
	return &value
}
