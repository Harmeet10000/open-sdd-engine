// Package testkit provides deterministic fakes for engine dependency ports:
// fixed and fault-injecting clocks, sequenced and fault-injecting identity
// sources, an in-memory filesystem fact reader with fault and blocking
// behavior, and canned policy and workspace snapshot observations.
//
// Fakes model the relevant behavior and faults; they are not mocks of
// concrete types. Every fake is safe for concurrent use so tests can run
// with t.Parallel and the race detector. Fakes honor caller cancellation so
// tests can prove the engine propagates context instead of replacing it.
package testkit

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"sdd.dev/sdd/internal/engine"
	"sdd.dev/sdd/internal/workspace"
)

var (
	_ engine.Clock             = (*FixedClock)(nil)
	_ engine.IDSource          = (*SequenceIDSource)(nil)
	_ engine.FileFacts         = (*MemFileFacts)(nil)
	_ engine.PolicySnapshotter = (*FixedPolicySnapshotter)(nil)
	_ engine.Snapshotter       = (*FixedSnapshotter)(nil)
)

// FixedClock is a manually advanced Clock. The zero value reports the Unix
// epoch in UTC; use NewFixedClock or Set to choose another instant.
type FixedClock struct {
	mutex sync.Mutex
	now   time.Time
	set   bool
}

// NewFixedClock returns a clock fixed at the given instant, normalized to UTC.
func NewFixedClock(at time.Time) *FixedClock {
	clock := &FixedClock{}
	clock.Set(at)
	return clock
}

// Now returns the fixed instant, or the Unix epoch when never set.
func (c *FixedClock) Now() time.Time {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if !c.set {
		return time.Unix(0, 0).UTC()
	}
	return c.now
}

// Set pins the clock to a new instant, normalized to UTC.
func (c *FixedClock) Set(at time.Time) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.now = at.UTC()
	c.set = true
}

// Advance moves the fixed instant forward by the given duration.
func (c *FixedClock) Advance(delta time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.now = c.now.Add(delta)
}

// SequenceIDSource mints predictable prefix-counter identities
// ("work-0", "work-1", ...). An injected fault makes NewID fail without
// consuming a sequence value, so tests can distinguish exhaustion from gaps.
type SequenceIDSource struct {
	mutex  sync.Mutex
	prefix string
	next   uint64
	fault  error
}

// NewSequenceIDSource returns an identity source sequencing from zero.
func NewSequenceIDSource(prefix string) *SequenceIDSource {
	return &SequenceIDSource{prefix: prefix}
}

// NewID returns the next sequenced identity or the injected fault.
func (s *SequenceIDSource) NewID() (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.fault != nil {
		return "", s.fault
	}
	identity := fmt.Sprintf("%s-%d", s.prefix, s.next)
	s.next++
	return identity, nil
}

// InjectFault makes subsequent NewID calls fail until ResetFault.
func (s *SequenceIDSource) InjectFault(err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.fault = err
}

// ResetFault clears the injected fault; the sequence resumes unconsumed.
func (s *SequenceIDSource) ResetFault() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.fault = nil
}

// MemFile is one in-memory file fact.
type MemFile struct {
	Bytes   []byte
	Mode    os.FileMode
	ModTime time.Time
	Kind    engine.FileKind
}

// MemFileFacts is an in-memory FileFacts fake. Missing paths fail with an
// error matching os.ErrNotExist, mirroring the production adapter's ENOENT
// behavior. Per-path faults and blocking reads model adapter failures and
// slow filesystems: a blocked ReadFile waits for the release channel or
// caller cancellation, whichever comes first.
type MemFileFacts struct {
	mutex   sync.Mutex
	files   map[string]MemFile
	faults  map[string]error
	blocked map[string]chan struct{}
}

// NewMemFileFacts returns a fake serving defensive copies of the given files.
// A zero Kind means a regular file. Keys are cleaned with filepath.Clean so
// fixtures written with slash separators resolve on every OS.
func NewMemFileFacts(files map[string]MemFile) *MemFileFacts {
	cloned := make(map[string]MemFile, len(files))
	for name, file := range files {
		kind := file.Kind
		if kind == "" {
			kind = engine.FileKindRegular
		}
		cloned[memKey(name)] = MemFile{
			Bytes:   append([]byte{}, file.Bytes...),
			Mode:    file.Mode,
			ModTime: file.ModTime,
			Kind:    kind,
		}
	}
	return &MemFileFacts{files: cloned, faults: map[string]error{}, blocked: map[string]chan struct{}{}}
}

// memKey normalizes a lookup path the same way constructor keys are
// normalized, so slash-written fixtures resolve under the host separator.
func memKey(abspath string) string {
	return NativePath(abspath)
}

// NativePath maps a slash-rooted virtual test path to an absolute path on the
// host OS. Unix keeps /repo as-is; Windows binds it to the temporary
// directory's volume (for example D:\repo). Relative inputs remain relative.
func NativePath(slashPath string) string {
	if filepath.Separator != '\\' || !strings.HasPrefix(slashPath, "/") {
		return filepath.Clean(filepath.FromSlash(slashPath))
	}
	volume := filepath.VolumeName(os.TempDir())
	if volume == "" {
		volume = "C:"
	}
	root := volume + string(filepath.Separator)
	remainder := strings.TrimPrefix(slashPath, "/")
	if remainder == "" {
		return root
	}
	return filepath.Join(root, filepath.FromSlash(remainder))
}

func childPrefix(abspath string) string {
	if filepath.Dir(abspath) == abspath {
		return abspath
	}
	return abspath + string(filepath.Separator)
}

// InjectFault makes Stat and ReadFile for path fail with the given error.
func (m *MemFileFacts) InjectFault(path string, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.faults[memKey(path)] = err
}

// BlockRead parks ReadFile for path until ReleaseRead closes the channel or
// the caller's context ends. Stat is unaffected.
func (m *MemFileFacts) BlockRead(path string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	path = memKey(path)
	if _, ok := m.blocked[path]; !ok {
		m.blocked[path] = make(chan struct{})
	}
}

// ReleaseRead unparks a blocked ReadFile for path.
func (m *MemFileFacts) ReleaseRead(path string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if channel, ok := m.blocked[memKey(path)]; ok {
		close(channel)
		delete(m.blocked, path)
	}
}

// Stat returns the recorded fact for path.
func (m *MemFileFacts) Stat(ctx context.Context, abspath string) (engine.FileFact, error) {
	if abspath == "" {
		return engine.FileFact{}, fmt.Errorf("%w: filesystem path is required", engine.ErrInvalidProtocolValue)
	}
	if err := ctx.Err(); err != nil {
		return engine.FileFact{}, err
	}
	abspath = memKey(abspath)
	m.mutex.Lock()
	fault, hasFault := m.faults[abspath]
	file, ok := m.files[abspath]
	hasChildren := false
	if !ok && !hasFault {
		prefix := childPrefix(abspath)
		for recorded := range m.files {
			if recorded != abspath && len(recorded) > len(prefix) && recorded[:len(prefix)] == prefix {
				hasChildren = true
				break
			}
		}
	}
	m.mutex.Unlock()
	if hasFault {
		return engine.FileFact{}, fault
	}
	if !ok {
		if hasChildren {
			return engine.NewFileFact(engine.FileFactInput{Kind: engine.FileKindDirectory})
		}
		return engine.FileFact{}, fmt.Errorf("%w: stat file facts: %w", engine.ErrDependencyFault, fs.ErrNotExist)
	}
	return engine.NewFileFact(engine.FileFactInput{
		Kind:    file.Kind,
		Size:    uint64(len(file.Bytes)),
		Mode:    file.Mode,
		ModTime: file.ModTime,
	})
}

// ReadFile returns a copy of the recorded bytes for path.
func (m *MemFileFacts) ReadFile(ctx context.Context, abspath string) ([]byte, error) {
	if abspath == "" {
		return nil, fmt.Errorf("%w: filesystem path is required", engine.ErrInvalidProtocolValue)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	abspath = memKey(abspath)
	m.mutex.Lock()
	fault, hasFault := m.faults[abspath]
	file, ok := m.files[abspath]
	channel, blocked := m.blocked[abspath]
	m.mutex.Unlock()
	if hasFault {
		return nil, fault
	}
	if !ok {
		return nil, fmt.Errorf("%w: read file facts: %w", engine.ErrDependencyFault, fs.ErrNotExist)
	}
	if file.Kind == engine.FileKindDirectory {
		return nil, fmt.Errorf("%w: read file facts on a directory", engine.ErrDependencyFault)
	}
	if blocked {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-channel:
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		m.mutex.Lock()
		file, ok = m.files[abspath]
		m.mutex.Unlock()
		if !ok {
			return nil, fmt.Errorf("%w: read file facts: %w", engine.ErrDependencyFault, fs.ErrNotExist)
		}
	}
	return append([]byte{}, file.Bytes...), nil
}

// ListDir returns the immediate children of path in sorted name order. A
// recorded path with further components beneath it is reported as a
// directory; an explicitly recorded directory entry keeps its kind. Unknown
// paths fail with an os.ErrNotExist-compatible error.
func (m *MemFileFacts) ListDir(ctx context.Context, abspath string) ([]engine.DirEntry, error) {
	if abspath == "" {
		return nil, fmt.Errorf("%w: filesystem path is required", engine.ErrInvalidProtocolValue)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	abspath = memKey(abspath)
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if fault, hasFault := m.faults[abspath]; hasFault {
		return nil, fault
	}
	prefix := childPrefix(abspath)
	children := map[string]engine.FileKind{}
	known := false
	for recorded, file := range m.files {
		if recorded == abspath {
			known = true
			if file.Kind == engine.FileKindDirectory {
				continue
			}
			return nil, fmt.Errorf("%w: list directory facts on a non-directory", engine.ErrDependencyFault)
		}
		rest, ok := cutPrefix(recorded, prefix)
		if !ok {
			continue
		}
		known = true
		name, _, hasMore := splitComponent(rest)
		kind := file.Kind
		if hasMore {
			kind = engine.FileKindDirectory
		}
		if current, seen := children[name]; !seen || (current != engine.FileKindDirectory && kind == engine.FileKindDirectory) {
			children[name] = kind
		}
	}
	if !known {
		return nil, fmt.Errorf("%w: list directory facts: %w", engine.ErrDependencyFault, fs.ErrNotExist)
	}
	names := slices.Sorted(maps.Keys(children))
	entries := make([]engine.DirEntry, 0, len(names))
	for _, name := range names {
		entry, err := engine.NewDirEntry(engine.DirEntryInput{Name: name, Kind: children[name]})
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func cutPrefix(recorded, prefix string) (string, bool) {
	if len(recorded) <= len(prefix) || recorded[:len(prefix)] != prefix {
		return "", false
	}
	return recorded[len(prefix):], true
}

func splitComponent(rest string) (name, remainder string, hasMore bool) {
	for i := 0; i < len(rest); i++ {
		if rest[i] == '/' || rest[i] == '\\' {
			return rest[:i], rest[i+1:], true
		}
	}
	return rest, "", false
}

// NewFixedPolicySnapshotter returns a snapshotter serving a copy of the
// given canned policy observation. An injected fault makes PolicySnapshot
// fail; the canned value is never aliased.
func NewFixedPolicySnapshotter(snapshot engine.PolicySnapshot) *FixedPolicySnapshotter {
	return &FixedPolicySnapshotter{snapshot: snapshot}
}

// FixedPolicySnapshotter serves one canned policy observation.
type FixedPolicySnapshotter struct {
	mutex    sync.Mutex
	snapshot engine.PolicySnapshot
	fault    error
}

// PolicySnapshot returns the canned observation or the injected fault.
func (f *FixedPolicySnapshotter) PolicySnapshot(ctx context.Context) (engine.PolicySnapshot, error) {
	if err := ctx.Err(); err != nil {
		return engine.PolicySnapshot{}, err
	}
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if f.fault != nil {
		return engine.PolicySnapshot{}, f.fault
	}
	return f.snapshot, nil
}

// InjectFault makes subsequent PolicySnapshot calls fail.
func (f *FixedPolicySnapshotter) InjectFault(err error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.fault = err
}

// FixedSnapshotter serves one canned workspace snapshot for engine tests
// that need an observation without a real capturer. The canned value is
// deep-copied on construction and on every return, so mutating a returned
// snapshot (or the constructor input) can never affect later observations.
type FixedSnapshotter struct {
	mutex    sync.Mutex
	snapshot workspace.Snapshot
	fault    error
}

// NewFixedSnapshotter returns a snapshotter serving the given snapshot.
func NewFixedSnapshotter(snapshot workspace.Snapshot) *FixedSnapshotter {
	return &FixedSnapshotter{snapshot: copySnapshot(snapshot)}
}

// Snapshot returns an isolated copy of the canned snapshot or the injected fault.
func (f *FixedSnapshotter) Snapshot(ctx context.Context, _ workspace.SnapshotRequest) (workspace.Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return workspace.Snapshot{}, err
	}
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if f.fault != nil {
		return workspace.Snapshot{}, f.fault
	}
	return copySnapshot(f.snapshot), nil
}

func copySnapshot(snapshot workspace.Snapshot) workspace.Snapshot {
	// Snapshot holds only JSON-safe values, so this marshal cannot fail;
	// panicking keeps a broken fake loud instead of serving zero snapshots.
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		panic("testkit: snapshot is not JSON-copyable: " + err.Error())
	}
	var copied workspace.Snapshot
	if err := json.Unmarshal(encoded, &copied); err != nil {
		panic("testkit: snapshot copy failed: " + err.Error())
	}
	return copied
}

// InjectFault makes subsequent Snapshot calls fail.
func (f *FixedSnapshotter) InjectFault(err error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.fault = err
}
