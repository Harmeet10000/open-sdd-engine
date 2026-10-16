// Package engine owns the deep application seam and its closed protocol
// values. This file defines the narrow dependency ports the engine consumes:
// clock, identity, immutable workspace observation, policy snapshot, and
// read-only filesystem facts. Interfaces live here because engine is the
// consumer; implementations live with their owners (workspace capture,
// native adapters) or in internal/testkit as deterministic fakes.
//
// Ports propagate the caller's context and never replace it. Signal-derived
// contexts enter at the CLI composition root (a later slice); these ports
// only honor cancellation at and across effect boundaries.
package engine

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"strings"
	"time"

	"sdd.dev/sdd/internal/workspace"
)

// ErrDependencyFault marks a non-context failure inside a dependency port.
// Engine error mapping matches it to decide between caller-correctable and
// operational outcomes. Context cancellation and deadline errors propagate
// unwrapped so errors.Is(err, context.Canceled) keeps working.
var ErrDependencyFault = errors.New("dependency port fault")

// Clock reports the current instant. Production code uses SystemClock;
// tests use a fixed testkit clock. Instants are UTC so persisted and
// machine-rendered timestamps share one zone.
type Clock interface {
	Now() time.Time
}

// SystemClock is the production Clock.
type SystemClock struct{}

// Now returns the current UTC instant.
func (SystemClock) Now() time.Time { return time.Now().UTC() }

// IDSource mints stable identities for work items, plans, and snapshots.
// IDs are random, never hashes of intent text: identical intents submitted
// twice must receive distinct identities so replay and conflict handling can
// tell the submissions apart.
type IDSource interface {
	NewID() (string, error)
}

// RandomIDSource is the production IDSource. It mints 128-bit
// cryptographic-random identities rendered as lowercase hex.
type RandomIDSource struct{}

// NewID returns one fresh random identity.
func (RandomIDSource) NewID() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", errors.Join(ErrDependencyFault, fmt.Errorf("mint identity: %w", err))
	}
	return hex.EncodeToString(raw[:]), nil
}

// Snapshotter is the engine-owned, one-method view over immutable workspace
// observation. *workspace.Reader satisfies it; the engine never touches the
// workspace Capturer seam directly. Read-set recipes, coherence checks, and
// digest rules stay in the workspace package.
type Snapshotter interface {
	Snapshot(ctx context.Context, request workspace.SnapshotRequest) (workspace.Snapshot, error)
}

// PolicySnapshot is a provisional immutable policy observation: source
// identity plus the opaque policy floor bytes the engine plans against. Full
// policy normalization and authority evaluation (IC-10) belong to
// artifact-validation; this value carries facts, not decisions.
type PolicySnapshot struct {
	source string
	digest string
	floor  []byte
}

// PolicySnapshotInput constructs a PolicySnapshot.
type PolicySnapshotInput struct {
	Source string
	Digest string
	Floor  []byte
}

// NewPolicySnapshot validates source identity and defensively copies the floor.
func NewPolicySnapshot(input PolicySnapshotInput) (PolicySnapshot, error) {
	if input.Source == "" || input.Digest == "" {
		return PolicySnapshot{}, fmt.Errorf("%w: policy source and digest are required", ErrInvalidProtocolValue)
	}
	floor := append([]byte{}, input.Floor...)
	return PolicySnapshot{source: input.Source, digest: input.Digest, floor: floor}, nil
}

// Source reports the trusted-source identity the floor was read from.
func (s PolicySnapshot) Source() string { return s.source }

// Digest reports the content digest of the exact policy bytes observed.
func (s PolicySnapshot) Digest() string { return s.digest }

// Floor returns a copy of the observed policy floor bytes.
func (s PolicySnapshot) Floor() []byte { return append([]byte{}, s.floor...) }

// PolicySnapshotter supplies the policy observation the engine plans
// against. Missing, expired, or unverifiable required policy fails closed at
// the engine mapping layer; this port only reports what was observed.
type PolicySnapshotter interface {
	PolicySnapshot(ctx context.Context) (PolicySnapshot, error)
}

// FileKind distinguishes the raw filesystem facts the engine may observe.
// It is deliberately narrower than workspace.FileKind: layout discovery needs
// shape, not durable identity.
type FileKind string

const (
	// FileKindRegular is an ordinary file whose bytes can be read.
	FileKindRegular FileKind = "regular"
	// FileKindDirectory is a directory whose entries can be listed by later slices.
	FileKindDirectory FileKind = "directory"
	// FileKindOther covers symlinks, devices, pipes, and anything else the
	// read-only foundation refuses to interpret as content.
	FileKindOther FileKind = "other"
)

func (k FileKind) valid() bool {
	switch k {
	case FileKindRegular, FileKindDirectory, FileKindOther:
		return true
	default:
		return false
	}
}

// FileFact is one immutable filesystem observation.
type FileFact struct {
	kind    FileKind
	size    uint64
	mode    os.FileMode
	modTime time.Time
}

// FileFactInput constructs a FileFact.
type FileFactInput struct {
	Kind    FileKind
	Size    uint64
	Mode    os.FileMode
	ModTime time.Time
}

// NewFileFact validates the kind and normalizes the timestamp to UTC.
func NewFileFact(input FileFactInput) (FileFact, error) {
	if !input.Kind.valid() {
		return FileFact{}, fmt.Errorf("%w: unsupported file kind %q", ErrInvalidProtocolValue, input.Kind)
	}
	return FileFact{kind: input.Kind, size: input.Size, mode: input.Mode, modTime: input.ModTime.UTC()}, nil
}

// Kind reports the observed file shape.
func (f FileFact) Kind() FileKind { return f.kind }

// Size reports the observed byte length.
func (f FileFact) Size() uint64 { return f.size }

// Mode reports the observed permission bits.
func (f FileFact) Mode() os.FileMode { return f.mode }

// ModTime reports the observed modification instant in UTC.
func (f FileFact) ModTime() time.Time { return f.modTime }

// FileFacts reads immutable filesystem facts for absolute host paths. It is
// read-only: later slices own mutation behind exact authorized plans. Paths
// are absolute because workspace-relative canonicalization belongs to layout
// discovery (FND-005), which consumes this port. There is intentionally no
// containment, symlink, or hard-link policy here; those decisions belong to
// workspace transactions and bounded execution.
type FileFacts interface {
	Stat(ctx context.Context, abspath string) (FileFact, error)
	ReadFile(ctx context.Context, abspath string) ([]byte, error)
	// ListDir returns the immediate children of abspath in byte-sorted
	// name order so traversal is deterministic on every platform.
	ListDir(ctx context.Context, abspath string) ([]DirEntry, error)
}

// DirEntry is one immediate child observed by ListDir. Name is a single
// path component, never a path with separators.
type DirEntry struct {
	name string
	kind FileKind
}

// DirEntryInput constructs a DirEntry.
type DirEntryInput struct {
	Name string
	Kind FileKind
}

// NewDirEntry rejects empty names, dot components, separators, and unknown kinds.
func NewDirEntry(input DirEntryInput) (DirEntry, error) {
	if input.Name == "" || input.Name == "." || input.Name == ".." || strings.ContainsAny(input.Name, `/\`) {
		return DirEntry{}, fmt.Errorf("%w: invalid directory entry name %q", ErrInvalidProtocolValue, input.Name)
	}
	if !input.Kind.valid() {
		return DirEntry{}, fmt.Errorf("%w: unsupported file kind %q", ErrInvalidProtocolValue, input.Kind)
	}
	return DirEntry{name: input.Name, kind: input.Kind}, nil
}

// Name reports the single-component child name.
func (e DirEntry) Name() string { return e.name }

// Kind reports the observed child shape.
func (e DirEntry) Kind() FileKind { return e.kind }

// OSFileFacts is the production FileFacts adapter over the host filesystem.
type OSFileFacts struct{}

// Stat observes one path without following any policy beyond the OS default.
func (OSFileFacts) Stat(ctx context.Context, abspath string) (FileFact, error) {
	if abspath == "" {
		return FileFact{}, fmt.Errorf("%w: filesystem path is required", ErrInvalidProtocolValue)
	}
	if err := ctx.Err(); err != nil {
		return FileFact{}, err
	}
	info, err := os.Stat(abspath)
	if err != nil {
		return FileFact{}, errors.Join(ErrDependencyFault, fmt.Errorf("stat file facts: %w", err))
	}
	if info.Size() < 0 {
		return FileFact{}, errors.Join(ErrDependencyFault, fmt.Errorf("stat file facts: negative size"))
	}
	return NewFileFact(FileFactInput{
		Kind: classifyFileInfo(info),
		// #nosec G115 -- info.Size() is proven non-negative by the guard above.
		Size:    uint64(info.Size()),
		Mode:    info.Mode(),
		ModTime: info.ModTime(),
	})
}

// ReadFile returns a copy of one file's bytes. Cancellation is honored
// before the read starts; mid-read interruption belongs to bounded execution.
func (OSFileFacts) ReadFile(ctx context.Context, abspath string) ([]byte, error) {
	if abspath == "" {
		return nil, fmt.Errorf("%w: filesystem path is required", ErrInvalidProtocolValue)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// #nosec G304 -- absolute-path reads are the port contract; containment and
	// traversal policy belong to workspace transactions and bounded execution.
	content, err := os.ReadFile(abspath)
	if err != nil {
		return nil, errors.Join(ErrDependencyFault, fmt.Errorf("read file facts: %w", err))
	}
	return content, nil
}

// ListDir returns the immediate children of abspath in sorted name order.
// A missing path fails with an os.ErrNotExist-compatible error, matching
// Stat behavior; a non-directory path fails with a dependency fault.
func (OSFileFacts) ListDir(ctx context.Context, abspath string) ([]DirEntry, error) {
	if abspath == "" {
		return nil, fmt.Errorf("%w: filesystem path is required", ErrInvalidProtocolValue)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// #nosec G304 -- absolute-path reads are the port contract; containment and
	// traversal policy belong to workspace transactions and bounded execution.
	infos, err := os.ReadDir(abspath)
	if err != nil {
		return nil, errors.Join(ErrDependencyFault, fmt.Errorf("list directory facts: %w", err))
	}
	entries := make([]DirEntry, 0, len(infos))
	for _, info := range infos {
		var kind FileKind
		switch {
		case info.IsDir():
			kind = FileKindDirectory
		case info.Type().IsRegular():
			kind = FileKindRegular
		default:
			kind = FileKindOther
		}
		entry, err := NewDirEntry(DirEntryInput{Name: info.Name(), Kind: kind})
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	slices.SortFunc(entries, func(a, b DirEntry) int { return strings.Compare(a.Name(), b.Name()) })
	return entries, nil
}

func classifyFileInfo(info fs.FileInfo) FileKind {
	switch {
	case info.IsDir():
		return FileKindDirectory
	case info.Mode().IsRegular():
		return FileKindRegular
	default:
		return FileKindOther
	}
}
