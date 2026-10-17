// Package filesystem adapts host filesystem primitives to workspace-owned
// ports. It contains no artifact, workflow, or completion policy.
package filesystem

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"sdd.dev/sdd/internal/workspace"
)

var (
	ErrRootInvalid       = errors.New("filesystem capture root invalid")
	ErrPathUnsafe        = errors.New("filesystem path unsafe")
	ErrIdentityChanged   = errors.New("filesystem identity changed during capture")
	ErrUnsupportedObject = errors.New("filesystem object type unsupported")
)

// CaptureConfig binds one adapter instance to a foundation-selected native
// workspace. The observations are copied and then checked against each request.
type CaptureConfig struct {
	Root      string
	Workspace workspace.WorkspaceObservation
	Work      *workspace.WorkObservation
	NewID     func() (string, error)
}

// Capturer provides collectively coherent, bounded, read-only observations.
// It writes no lock, cache, heartbeat, or repair material.
type Capturer struct {
	root      string
	workspace workspace.WorkspaceObservation
	work      *workspace.WorkObservation
	newID     func() (string, error)
}

var _ workspace.Capturer = (*Capturer)(nil)

// NewCapturer validates and freezes one selected native root.
func NewCapturer(config CaptureConfig) (*Capturer, error) {
	if !filepath.IsAbs(config.Root) {
		return nil, fmt.Errorf("%w: root must be absolute", ErrRootInvalid)
	}
	root := filepath.Clean(config.Root)
	if config.Workspace.WorkspaceID == "" || config.Workspace.FormatVersion == 0 {
		return nil, fmt.Errorf("%w: workspace observation is incomplete", ErrRootInvalid)
	}
	if config.Workspace.CanonicalRoot == "" || filepath.Clean(config.Workspace.CanonicalRoot) != root {
		return nil, fmt.Errorf("%w: configured and observed roots differ", ErrRootInvalid)
	}
	newID := config.NewID
	if newID == nil {
		newID = randomSnapshotID
	}
	return &Capturer{
		root:      root,
		workspace: config.Workspace,
		work:      cloneWork(config.Work),
		newID:     newID,
	}, nil
}

// Capture retries the entire read set when any file or generation observation
// changes. A successful result is from one stable transaction generation.
func (c *Capturer) Capture(ctx context.Context, request workspace.SnapshotRequest) (workspace.Capture, error) {
	if err := ctx.Err(); err != nil {
		return workspace.Capture{}, err
	}
	if request.Selection.WorkspaceID != c.workspace.WorkspaceID {
		return workspace.Capture{}, fmt.Errorf("%w: selected workspace does not match capturer", ErrRootInvalid)
	}
	if request.Selection.ExplicitRoot != "" {
		if !filepath.IsAbs(request.Selection.ExplicitRoot) || filepath.Clean(request.Selection.ExplicitRoot) != c.root {
			return workspace.Capture{}, fmt.Errorf("%w: explicit root does not match capturer", ErrRootInvalid)
		}
	}
	if err := validateCollisionKeys(request.Paths); err != nil {
		return workspace.Capture{}, err
	}

	var lastErr error
	for attempt := 0; attempt <= request.Limits.StableReadRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			return workspace.Capture{}, err
		}
		capture, err := c.captureOnce(ctx, request)
		if err == nil {
			return capture, nil
		}
		if !errors.Is(err, ErrIdentityChanged) && !errors.Is(err, workspace.ErrSnapshotUnstable) {
			return workspace.Capture{}, err
		}
		lastErr = err
	}
	return workspace.Capture{}, fmt.Errorf("%w: stable capture retries exhausted: %w", workspace.ErrSnapshotUnstable, lastErr)
}

func (c *Capturer) captureOnce(ctx context.Context, request workspace.SnapshotRequest) (workspace.Capture, error) {
	root, err := os.OpenRoot(c.root)
	if err != nil {
		return workspace.Capture{}, fmt.Errorf("%w: open selected root: %w", ErrRootInvalid, err)
	}
	defer root.Close()

	before, err := readGeneration(ctx, root, request.Limits.MaxBytes)
	if err != nil {
		return workspace.Capture{}, err
	}
	if before.ActiveTransactionID() != "" {
		return workspace.Capture{
			Workspace:           c.workspace,
			Work:                cloneWork(c.work),
			GenerationBefore:    before.StableGeneration(),
			GenerationAfter:     before.StableGeneration(),
			ActiveTransactionID: before.ActiveTransactionID(),
		}, nil
	}

	paths := make([]workspace.PathObservation, 0, len(request.Paths))
	var total uint64
	for _, name := range request.Paths {
		remaining := request.Limits.MaxBytes - total
		observation, readErr := readStablePath(ctx, root, name, remaining)
		if readErr != nil {
			return workspace.Capture{}, readErr
		}
		if observation.Size > remaining {
			return workspace.Capture{}, fmt.Errorf("%w: read set exceeds %d bytes", workspace.ErrLimitExceeded, request.Limits.MaxBytes)
		}
		total += observation.Size
		paths = append(paths, observation)
	}

	after, err := readGeneration(ctx, root, request.Limits.MaxBytes)
	if err != nil {
		return workspace.Capture{}, err
	}
	if before.StableGeneration() != after.StableGeneration() || before.ActiveTransactionID() != after.ActiveTransactionID() {
		return workspace.Capture{}, fmt.Errorf("%w: generation changed during capture", workspace.ErrSnapshotUnstable)
	}
	if after.ActiveTransactionID() != "" {
		return workspace.Capture{}, fmt.Errorf("%w: transaction became active during capture", workspace.ErrSnapshotUnstable)
	}

	snapshotID, err := c.newID()
	if err != nil {
		return workspace.Capture{}, fmt.Errorf("allocate snapshot ID: %w", err)
	}
	if snapshotID == "" {
		return workspace.Capture{}, fmt.Errorf("allocate snapshot ID: empty result")
	}
	return workspace.Capture{
		SnapshotID:       snapshotID,
		Workspace:        c.workspace,
		Work:             cloneWork(c.work),
		GenerationBefore: before.StableGeneration(),
		GenerationAfter:  after.StableGeneration(),
		Paths:            paths,
	}, nil
}

func readGeneration(ctx context.Context, root *os.Root, maxBytes uint64) (workspace.GenerationRecord, error) {
	if err := ctx.Err(); err != nil {
		return workspace.GenerationRecord{}, err
	}
	data, err := readBoundedFile(root, string(workspace.GenerationRecordPath), maxBytes)
	if err != nil {
		return workspace.GenerationRecord{}, fmt.Errorf("read generation record: %w", err)
	}
	record, err := workspace.ParseGenerationRecord(data)
	if err != nil {
		return workspace.GenerationRecord{}, fmt.Errorf("read generation record: %w", err)
	}
	return record, nil
}

func readStablePath(ctx context.Context, root *os.Root, name workspace.WorkspacePath, maxBytes uint64) (workspace.PathObservation, error) {
	if err := rejectSymlinkComponents(root, string(name)); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return workspace.PathObservation{Path: name, Presence: workspace.Absent, Kind: workspace.FileKindNone, CompareKey: comparisonKey(string(name))}, nil
		}
		return workspace.PathObservation{}, err
	}
	if err := ctx.Err(); err != nil {
		return workspace.PathObservation{}, err
	}

	file, err := root.Open(string(name))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return workspace.PathObservation{Path: name, Presence: workspace.Absent, Kind: workspace.FileKindNone, CompareKey: comparisonKey(string(name))}, nil
		}
		return workspace.PathObservation{}, fmt.Errorf("open %q: %w", name, err)
	}
	defer file.Close()

	before, err := file.Stat()
	if err != nil {
		return workspace.PathObservation{}, fmt.Errorf("stat %q before read: %w", name, err)
	}
	kind, err := observedKind(before.Mode())
	if err != nil {
		return workspace.PathObservation{}, fmt.Errorf("%w: %q: %w", ErrUnsupportedObject, name, err)
	}
	identityBefore, linksBefore := nativeIdentity(before)
	beforeSize, err := fileSize(before)
	if err != nil {
		return workspace.PathObservation{}, fmt.Errorf("stat %q before read: %w", name, err)
	}
	var data []byte
	if kind == workspace.FileKindRegular {
		if beforeSize > maxBytes {
			return workspace.PathObservation{}, fmt.Errorf("%w: %q exceeds remaining byte limit", workspace.ErrLimitExceeded, name)
		}
		data, err = readBounded(file, maxBytes)
		if err != nil {
			return workspace.PathObservation{}, fmt.Errorf("read %q: %w", name, err)
		}
	}
	if contextErr := ctx.Err(); contextErr != nil {
		return workspace.PathObservation{}, contextErr
	}
	after, err := file.Stat()
	if err != nil {
		return workspace.PathObservation{}, fmt.Errorf("stat %q after read: %w", name, err)
	}
	identityAfter, linksAfter := nativeIdentity(after)
	afterSize, err := fileSize(after)
	if err != nil {
		return workspace.PathObservation{}, fmt.Errorf("stat %q after read: %w", name, err)
	}
	if !sameObservation(before, after, identityBefore, identityAfter, linksBefore, linksAfter) {
		return workspace.PathObservation{}, fmt.Errorf("%w: %q", ErrIdentityChanged, name)
	}
	return workspace.PathObservation{
		Path:       name,
		Presence:   workspace.Present,
		Kind:       kind,
		Bytes:      data,
		Size:       afterSize,
		Mode:       uint32(after.Mode().Perm()),
		NativeID:   identityAfter,
		LinkCount:  linksAfter,
		CompareKey: comparisonKey(string(name)),
	}, nil
}

func fileSize(info os.FileInfo) (uint64, error) {
	size := info.Size()
	if size < 0 {
		return 0, fmt.Errorf("negative size %d", size)
	}
	return uint64(size), nil
}

func rejectSymlinkComponents(root *os.Root, name string) error {
	components := strings.Split(name, "/")
	for index := range components {
		current := strings.Join(components[:index+1], "/")
		info, err := root.Lstat(current)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) && index == len(components)-1 {
				return fs.ErrNotExist
			}
			return fmt.Errorf("inspect path component %q: %w", current, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("%w: symlink component %q", ErrPathUnsafe, current)
		}
		if index < len(components)-1 && !info.IsDir() {
			return fmt.Errorf("%w: non-directory component %q", ErrPathUnsafe, current)
		}
	}
	return nil
}

func readBoundedFile(root *os.Root, name string, maxBytes uint64) ([]byte, error) {
	file, err := root.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return readBounded(file, maxBytes)
}

func readBounded(reader io.Reader, maxBytes uint64) ([]byte, error) {
	if maxBytes >= math.MaxInt64 {
		return nil, fmt.Errorf("%w: unsupported maximum byte limit", workspace.ErrLimitExceeded)
	}
	limited := io.LimitReader(reader, int64(maxBytes)+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if uint64(len(data)) > maxBytes {
		return nil, workspace.ErrLimitExceeded
	}
	return data, nil
}

func observedKind(mode os.FileMode) (workspace.FileKind, error) {
	switch {
	case mode.IsRegular():
		return workspace.FileKindRegular, nil
	case mode.IsDir():
		return workspace.FileKindDirectory, nil
	default:
		return workspace.FileKindNone, fmt.Errorf("mode %s", mode.Type())
	}
}

func sameObservation(before, after os.FileInfo, idBefore, idAfter workspace.NativeFileIdentity, linksBefore, linksAfter *uint64) bool {
	return before.Size() == after.Size() &&
		before.Mode() == after.Mode() &&
		before.ModTime().Equal(after.ModTime()) &&
		idBefore == idAfter &&
		equalOptionalUint64(linksBefore, linksAfter)
}

func equalOptionalUint64(left, right *uint64) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func validateCollisionKeys(paths []workspace.WorkspacePath) error {
	keys := make([]string, 0, len(paths))
	for _, name := range paths {
		keys = append(keys, comparisonKey(string(name)))
	}
	slices.Sort(keys)
	for index := 1; index < len(keys); index++ {
		if keys[index] == keys[index-1] {
			return fmt.Errorf("%w: requested paths collide under platform comparison", ErrPathUnsafe)
		}
	}
	return nil
}

func comparisonKey(name string) string {
	if runtime.GOOS == "windows" {
		return strings.ToUpper(name)
	}
	return name
}

func randomSnapshotID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return "snap-" + hex.EncodeToString(value[:]), nil
}

func cloneWork(source *workspace.WorkObservation) *workspace.WorkObservation {
	if source == nil {
		return nil
	}
	value := *source
	return &value
}
