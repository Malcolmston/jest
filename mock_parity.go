package jest

import "sync"

var (
	mockRegistryMu sync.Mutex
	mockRegistry   []*Mock
)

// registerGlobalMock records a mock in the global registry consulted by
// [ClearAllMocks] and [ResetAllMocks]. Every mock created through [NewMock]
// (and therefore every [Fn0]/[Fn1]/[Fn2], [Spy0]/[Spy1]/[Spy2] and [SpyOn]) is
// registered.
func registerGlobalMock(m *Mock) {
	mockRegistryMu.Lock()
	mockRegistry = append(mockRegistry, m)
	mockRegistryMu.Unlock()
}

// MockClear resets the mock's recorded call history (its calls and results)
// without changing any configured implementation or return values, mirroring
// Jest's mockFn.mockClear. It returns the mock to allow chaining.
func (m *Mock) MockClear() *Mock {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = nil
	return m
}

// MockReset restores the mock to its freshly-created state: it clears the call
// history and removes every configured implementation, one-shot behavior and
// return value, mirroring Jest's mockFn.mockReset. It returns the mock to allow
// chaining.
func (m *Mock) MockReset() *Mock {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = nil
	m.fixed = nil
	m.hasFixed = false
	m.seq = nil
	m.seqIndex = 0
	m.impl = nil
	m.once = nil
	return m
}

// MockName sets the mock's descriptive name (used in matcher failure messages)
// and returns the mock to allow chaining, mirroring Jest's mockFn.mockName. The
// current name is read back with [Mock.Name].
func (m *Mock) MockName(name string) *Mock {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.name = name
	return m
}

// ClearAllMocks clears the recorded call history of every mock created through
// [NewMock] and its wrappers, mirroring Jest's jest.clearAllMocks. Configured
// implementations and return values are preserved; use [ResetAllMocks] to also
// discard those.
func ClearAllMocks() {
	mockRegistryMu.Lock()
	mocks := make([]*Mock, len(mockRegistry))
	copy(mocks, mockRegistry)
	mockRegistryMu.Unlock()
	for _, m := range mocks {
		m.MockClear()
	}
}

// ResetAllMocks resets every mock created through [NewMock] and its wrappers to
// its freshly-created state, clearing call history and removing configured
// implementations and return values, mirroring Jest's jest.resetAllMocks. To
// also reinstate the originals replaced by [SpyOn], use [RestoreAllMocks].
func ResetAllMocks() {
	mockRegistryMu.Lock()
	mocks := make([]*Mock, len(mockRegistry))
	copy(mocks, mockRegistry)
	mockRegistryMu.Unlock()
	for _, m := range mocks {
		m.MockReset()
	}
}
