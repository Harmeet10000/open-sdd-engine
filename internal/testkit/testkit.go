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
	"fmt"
	"io/fs"
	"os"
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
// epoch; use NewFixedClock or Set to choose an instant.
type FixedClock struct {
	mutex sync.Mutex
	now   time.Time
}

// NewFixedClock returns a clock fixed at the given instant, normalized to UTC.
func NewFixedClock(at time.Time) *FixedClock {
	clock := &FixedClock{}
	clock.Set(at)
	return clock
}

// Now returns the fixed instant.
func (c *FixedClock) Now() time.Time {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.now
}

// Set pins the clock to a new instant, normalized to UTC.
func (c *FixedClock) Set(at time.Time) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.now = at.UTC()
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
// A zero Kind means a regular file.
func NewMemFileFacts(files map[string]MemFile) *MemFileFacts {
	cloned := make(map[string]MemFile, len(files))
	for name, file := range files {
		kind := file.Kind
		if kind == "" {
			kind = engine.FileKindRegular
		}
		cloned[name] = MemFile{
			Bytes:   append([]byte{}, file.Bytes...),
			Mode:    file.Mode,
			ModTime: file.ModTime,
			Kind:    kind,
		}
	}
	return &MemFileFacts{files: cloned, faults: map[string]error{}, blocked: map[string]chan struct{}{}}
}

// InjectFault makes Stat and ReadFile for path fail with the given error.
func (m *MemFileFacts) InjectFault(path string, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.faults[path] = err
}

// BlockRead parks ReadFile for path until ReleaseRead closes the channel or
// the caller's context ends. Stat is unaffected.
func (m *MemFileFacts) BlockRead(path string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if _, ok := m.blocked[path]; !ok {
		m.blocked[path] = make(chan struct{})
	}
}

// ReleaseRead unparks a blocked ReadFile for path.
func (m *MemFileFacts) ReleaseRead(path string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if channel, ok := m.blocked[path]; ok {
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
	m.mutex.Lock()
	fault, hasFault := m.faults[abspath]
	file, ok := m.files[abspath]
	m.mutex.Unlock()
	if hasFault {
		return engine.FileFact{}, fault
	}
	if !ok {
		return engine.FileFact{}, fmt.Errorf("stat file facts: %w", fs.ErrNotExist)
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
	m.mutex.Lock()
	fault, hasFault := m.faults[abspath]
	file, ok := m.files[abspath]
	channel, blocked := m.blocked[abspath]
	m.mutex.Unlock()
	if hasFault {
		return nil, fault
	}
	if !ok {
		return nil, fmt.Errorf("read file facts: %w", fs.ErrNotExist)
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
			return nil, fmt.Errorf("read file facts: %w", fs.ErrNotExist)
		}
	}
	return append([]byte{}, file.Bytes...), nil
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
// that need an observation without a real capturer.
type FixedSnapshotter struct {
	mutex    sync.Mutex
	snapshot workspace.Snapshot
	fault    error
}

// NewFixedSnapshotter returns a snapshotter serving the given snapshot.
func NewFixedSnapshotter(snapshot workspace.Snapshot) *FixedSnapshotter {
	return &FixedSnapshotter{snapshot: snapshot}
}

// Snapshot returns the canned snapshot or the injected fault.
func (f *FixedSnapshotter) Snapshot(ctx context.Context, _ workspace.SnapshotRequest) (workspace.Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return workspace.Snapshot{}, err
	}
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if f.fault != nil {
		return workspace.Snapshot{}, f.fault
	}
	return f.snapshot, nil
}

// InjectFault makes subsequent Snapshot calls fail.
func (f *FixedSnapshotter) InjectFault(err error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.fault = err
}
