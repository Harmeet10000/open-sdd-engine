package engine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	// ErrWorkspaceNotFound reports that no usable workspace marker exists:
	// no .sdd directory was found walking upward, or the layout is missing
	// its manifest. Maps to DIAG_WORKSPACE_NOT_FOUND.
	ErrWorkspaceNotFound = errors.New("workspace not found")
	// ErrInvalidRoot reports that an explicitly selected root is unusable.
	// There is no fallback to a valid parent. Maps to
	// DIAG_WORKSPACE_NOT_FOUND.
	ErrInvalidRoot = errors.New("invalid explicit workspace root")
	// ErrWorkspaceIncompatible reports a version or structural contract the
	// foundation does not implement: wrong format/schema versions, duplicate
	// work IDs, malformed records, or unknown fields. Maps to
	// DIAG_WORKSPACE_VERSION_INCOMPATIBLE.
	ErrWorkspaceIncompatible = errors.New("workspace layout incompatible")
	// ErrSelectionAmbiguous reports multiple active work items with no
	// explicit selection. Maps to DIAG_WORKSPACE_SELECTION_AMBIGUOUS.
	ErrSelectionAmbiguous = errors.New("work selection ambiguous")
	// ErrWorkNotFound reports an explicit work ID with no matching record.
	// Maps to DIAG_WORK_ITEM_NOT_FOUND.
	ErrWorkNotFound = errors.New("work item not found")
)

// LayoutDiagnosticCode translates a discovery or selection failure into its
// stable catalog code. Unknown errors map to DIAG_USAGE_ERROR; callers must
// not invent codes.
func LayoutDiagnosticCode(err error) string {
	switch {
	case errors.Is(err, ErrSelectionAmbiguous):
		return "DIAG_WORKSPACE_SELECTION_AMBIGUOUS"
	case errors.Is(err, ErrWorkNotFound):
		return "DIAG_WORK_ITEM_NOT_FOUND"
	case errors.Is(err, ErrWorkspaceIncompatible):
		return "DIAG_WORKSPACE_VERSION_INCOMPATIBLE"
	case errors.Is(err, ErrWorkspaceNotFound), errors.Is(err, ErrInvalidRoot):
		return "DIAG_WORKSPACE_NOT_FOUND"
	default:
		return "DIAG_USAGE_ERROR"
	}
}

// AmbiguousSelectionError carries the candidate work IDs that forced an
// ambiguous-selection failure so diagnostics can name them.
type AmbiguousSelectionError struct {
	Candidates []string
}

func (e *AmbiguousSelectionError) Error() string {
	return fmt.Sprintf("%v: %d candidate work items require explicit selection", ErrSelectionAmbiguous, len(e.Candidates))
}

// Unwrap exposes ErrSelectionAmbiguous for errors.Is/As.
func (e *AmbiguousSelectionError) Unwrap() error { return ErrSelectionAmbiguous }

// AmbiguousCandidates returns the candidate IDs when err is an ambiguous
// selection, or nil otherwise.
func AmbiguousCandidates(err error) []string {
	var ambiguous *AmbiguousSelectionError
	if errors.As(err, &ambiguous) {
		return append([]string{}, ambiguous.Candidates...)
	}
	return nil
}

// WorkRecord is one immutable canonical work entry decoded from
// .sdd/work/<slug>/work.yaml. The slug is display data; identity is the ID.
type WorkRecord struct {
	id          string
	slug        string
	status      string
	composition string
	revision    uint64
}

// ID reports the immutable work identity.
func (r WorkRecord) ID() string { return r.id }

// Slug reports the mutable display slug the record was found under.
func (r WorkRecord) Slug() string { return r.slug }

// Status reports the lifecycle status.
func (r WorkRecord) Status() string { return r.status }

// Composition reports the journey/composition reference.
func (r WorkRecord) Composition() string { return r.composition }

// Revision reports the monotonic work revision.
func (r WorkRecord) Revision() uint64 { return r.revision }

// Layout is the immutable foundation view of a discovered workspace:
// canonical root, workspace identity, and work records ordered by work ID.
// It carries no policy, workflow, or artifact meaning; durable raw
// observation stays in the workspace package. Values are plain data, so a
// copied Layout shares nothing mutable.
type Layout struct {
	root          string
	workspaceID   string
	formatVersion uint64
	revision      uint64
	works         []WorkRecord
}

// Root reports the absolute canonical workspace root.
func (l Layout) Root() string { return l.root }

// WorkspaceID reports the immutable workspace identity.
func (l Layout) WorkspaceID() string { return l.workspaceID }

// FormatVersion reports the accepted native format version (always 1).
func (l Layout) FormatVersion() uint64 { return l.formatVersion }

// Revision reports the workspace revision.
func (l Layout) Revision() uint64 { return l.revision }

// Works returns the work records in work-ID order.
func (l Layout) Works() []WorkRecord { return append([]WorkRecord{}, l.works...) }

// SelectWork resolves one work record from a layout. An explicit ID matches
// by immutable identity only, never by slug. An empty ID selects the single
// active record when exactly one exists, succeeds vacuously when none exists,
// and fails with an *AmbiguousSelectionError otherwise. Explicit immutable-ID
// selection may address a historical/non-active record for inspection.
func SelectWork(layout Layout, workID string) (WorkRecord, error) {
	if workID != "" {
		for _, record := range layout.works {
			if record.id == workID {
				return record, nil
			}
		}
		return WorkRecord{}, fmt.Errorf("%w: %q", ErrWorkNotFound, workID)
	}
	active := make([]WorkRecord, 0, len(layout.works))
	for _, record := range layout.works {
		if record.status == "active" {
			active = append(active, record)
		}
	}
	switch len(active) {
	case 0:
		return WorkRecord{}, nil
	case 1:
		return active[0], nil
	default:
		candidates := make([]string, 0, len(active))
		for _, record := range active {
			candidates = append(candidates, record.id)
		}
		return WorkRecord{}, &AmbiguousSelectionError{Candidates: candidates}
	}
}

// DiscoverLayout finds and decodes the canonical workspace layout. With an
// explicit root it inspects that root only and never falls back to a parent.
// Without one it walks upward from startDir for the nearest .sdd marker; an
// invalid nearest marker is an error, never permission to keep climbing.
// Every file is read once through the facts port; discovery never mutates.
func DiscoverLayout(ctx context.Context, facts FileFacts, startDir, explicitRoot string) (Layout, error) {
	if facts == nil {
		return Layout{}, fmt.Errorf("%w: file facts port is required", ErrInvalidProtocolValue)
	}
	if err := ctx.Err(); err != nil {
		return Layout{}, err
	}
	if explicitRoot != "" {
		layout, err := readLayout(ctx, facts, explicitRoot)
		if err != nil {
			return Layout{}, fmt.Errorf("%w: %s: %w", ErrInvalidRoot, explicitRoot, err)
		}
		return layout, nil
	}
	if startDir == "" {
		return Layout{}, fmt.Errorf("%w: discovery start directory is required", ErrInvalidProtocolValue)
	}
	if !filepath.IsAbs(startDir) {
		return Layout{}, fmt.Errorf("%w: discovery start directory must be absolute", ErrInvalidProtocolValue)
	}
	current := filepath.Clean(startDir)
	for {
		if err := ctx.Err(); err != nil {
			return Layout{}, err
		}
		marker := filepath.Join(current, ".sdd")
		stat, statErr := facts.Stat(ctx, marker)
		switch {
		case statErr == nil:
			if stat.Kind() != FileKindDirectory {
				return Layout{}, fmt.Errorf("%w: workspace marker %s is not a directory", ErrWorkspaceIncompatible, marker)
			}
			layout, err := readLayout(ctx, facts, current)
			if err != nil {
				return Layout{}, err
			}
			return layout, nil
		case errors.Is(statErr, os.ErrNotExist):
			// No marker here; keep climbing.
		default:
			return Layout{}, statErr
		}
		parent := filepath.Dir(current)
		if parent == current {
			return Layout{}, fmt.Errorf("%w: no .sdd marker above %s", ErrWorkspaceNotFound, startDir)
		}
		current = parent
	}
}

func readLayout(ctx context.Context, facts FileFacts, root string) (Layout, error) {
	manifest, err := facts.ReadFile(ctx, filepath.Join(root, ".sdd", "workspace.yaml"))
	if err != nil {
		return Layout{}, fmt.Errorf("%w: unreadable workspace manifest: %w", ErrWorkspaceNotFound, err)
	}
	workspaceID, revision, err := decodeWorkspaceManifest(manifest)
	if err != nil {
		return Layout{}, err
	}
	entries, err := facts.ListDir(ctx, filepath.Join(root, ".sdd", "work"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Layout{root: root, workspaceID: workspaceID, formatVersion: 1, revision: revision}, nil
		}
		return Layout{}, fmt.Errorf("%w: unreadable work directory: %w", ErrWorkspaceNotFound, err)
	}
	works := make([]WorkRecord, 0, len(entries))
	seen := map[string]string{}
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return Layout{}, err
		}
		if entry.Kind() != FileKindDirectory {
			continue
		}
		record, err := readWorkRecord(ctx, facts, filepath.Join(root, ".sdd", "work", entry.Name()))
		if err != nil {
			return Layout{}, err
		}
		if previous, duplicate := seen[record.id]; duplicate {
			return Layout{}, fmt.Errorf(
				"%w: duplicate work ID %q under %q and %q",
				ErrWorkspaceIncompatible,
				record.id,
				previous,
				entry.Name(),
			)
		}
		seen[record.id] = entry.Name()
		if record.slug != entry.Name() {
			return Layout{}, fmt.Errorf(
				"%w: work %q slug %q does not match directory %q",
				ErrWorkspaceIncompatible,
				record.id,
				record.slug,
				entry.Name(),
			)
		}
		works = append(works, record)
	}
	slices.SortFunc(works, func(a, b WorkRecord) int { return strings.Compare(a.id, b.id) })
	return Layout{root: root, workspaceID: workspaceID, formatVersion: 1, revision: revision, works: works}, nil
}

type workspaceManifest struct {
	FormatVersion uint64  `yaml:"formatVersion"`
	WorkspaceID   string  `yaml:"workspaceId"`
	Revision      *uint64 `yaml:"revision"`
}

func decodeWorkspaceManifest(content []byte) (workspaceID string, revision uint64, err error) {
	var manifest workspaceManifest
	if err := decodeStrict(content, &manifest); err != nil {
		return "", 0, fmt.Errorf("%w: malformed workspace manifest: %w", ErrWorkspaceIncompatible, err)
	}
	if manifest.FormatVersion != 1 {
		return "", 0, fmt.Errorf("%w: unsupported workspace format version %d", ErrWorkspaceIncompatible, manifest.FormatVersion)
	}
	if manifest.WorkspaceID == "" {
		return "", 0, fmt.Errorf("%w: workspace manifest lacks a workspace ID", ErrWorkspaceIncompatible)
	}
	if manifest.Revision == nil {
		return "", 0, fmt.Errorf("%w: workspace manifest lacks a revision", ErrWorkspaceIncompatible)
	}
	return manifest.WorkspaceID, *manifest.Revision, nil
}

type workManifest struct {
	SchemaVersion uint64  `yaml:"schemaVersion"`
	WorkID        string  `yaml:"workId"`
	Slug          string  `yaml:"slug"`
	Revision      *uint64 `yaml:"revision"`
	Status        string  `yaml:"status"`
	Composition   string  `yaml:"composition"`
}

func readWorkRecord(ctx context.Context, facts FileFacts, dir string) (WorkRecord, error) {
	content, err := facts.ReadFile(ctx, filepath.Join(dir, "work.yaml"))
	if err != nil {
		return WorkRecord{}, fmt.Errorf("%w: unreadable work record: %w", ErrWorkspaceIncompatible, err)
	}
	var manifest workManifest
	if err := decodeStrict(content, &manifest); err != nil {
		return WorkRecord{}, fmt.Errorf("%w: malformed work record: %w", ErrWorkspaceIncompatible, err)
	}
	if manifest.SchemaVersion != 1 {
		return WorkRecord{}, fmt.Errorf("%w: unsupported work schema version %d", ErrWorkspaceIncompatible, manifest.SchemaVersion)
	}
	if manifest.WorkID == "" || manifest.Slug == "" || manifest.Status == "" || manifest.Composition == "" {
		return WorkRecord{}, fmt.Errorf("%w: work record lacks identity, status, or composition", ErrWorkspaceIncompatible)
	}
	if manifest.Revision == nil {
		return WorkRecord{}, fmt.Errorf("%w: work record lacks a revision", ErrWorkspaceIncompatible)
	}
	return WorkRecord{
		id:          manifest.WorkID,
		slug:        manifest.Slug,
		status:      manifest.Status,
		composition: manifest.Composition,
		revision:    *manifest.Revision,
	}, nil
}

func decodeStrict(content []byte, target any) error {
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	decoder.KnownFields(true)
	if err := decoder.Decode(target); err != nil {
		return err
	}
	// A second document of any shape, including an ostensibly empty one,
	// is rejected: trailing records must never slip past the first.
	var extra any
	switch err := decoder.Decode(&extra); {
	case errors.Is(err, io.EOF):
		return nil
	case err != nil:
		return err
	default:
		return fmt.Errorf("unexpected trailing document")
	}
}
