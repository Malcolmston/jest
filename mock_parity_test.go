package jest

import "testing"

func TestParityMockClear(t *testing.T) {
	m := NewMock("f")
	m.Return(42)
	m.Call(1)
	m.Call(2)
	if m.CallCount() != 2 {
		t.Fatalf("want 2 calls, got %d", m.CallCount())
	}
	m.MockClear()
	if m.CallCount() != 0 {
		t.Errorf("MockClear should clear calls, got %d", m.CallCount())
	}
	// Return value is preserved.
	got := m.Call(3)
	if len(got) != 1 || got[0] != 42 {
		t.Errorf("MockClear should preserve return value, got %v", got)
	}
}

func TestParityMockReset(t *testing.T) {
	m := NewMock("f")
	m.Return(42)
	m.Call(1)
	m.MockReset()
	if m.CallCount() != 0 {
		t.Errorf("MockReset should clear calls, got %d", m.CallCount())
	}
	got := m.Call(2)
	if len(got) != 0 {
		t.Errorf("MockReset should discard return value, got %v", got)
	}
}

func TestParityMockName(t *testing.T) {
	m := NewMock("orig")
	if m.MockName("renamed") != m {
		t.Error("MockName should return the mock for chaining")
	}
	if m.Name() != "renamed" {
		t.Errorf("Name=%q, want renamed", m.Name())
	}
}

func TestParityClearAllMocks(t *testing.T) {
	a := NewMock("a")
	b := NewMock("b")
	a.Return(1)
	b.Return(2)
	a.Call()
	b.Call()
	ClearAllMocks()
	if a.CallCount() != 0 || b.CallCount() != 0 {
		t.Errorf("ClearAllMocks should clear all: a=%d b=%d", a.CallCount(), b.CallCount())
	}
	// Return values preserved.
	if got := a.Call(); len(got) != 1 || got[0] != 1 {
		t.Errorf("ClearAllMocks should preserve returns, got %v", got)
	}
}

func TestParityResetAllMocks(t *testing.T) {
	a := NewMock("a")
	a.Return(1)
	a.Call()
	ResetAllMocks()
	if a.CallCount() != 0 {
		t.Errorf("ResetAllMocks should clear calls, got %d", a.CallCount())
	}
	if got := a.Call(); len(got) != 0 {
		t.Errorf("ResetAllMocks should discard returns, got %v", got)
	}
}
