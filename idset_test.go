package treerack

import "testing"

func TestIDSet(t *testing.T) {
	s := &idSet{}

	s.set(42)
	if !s.has(42) {
		t.Error("failed to set id")
		return
	}

	if s.has(42 + 64) {
		t.Error("invalid value set")
		return
	}

	s.unset(42 + 64)

	if !s.has(42) {
		t.Error("failed to set id")
		return
	}

	if s.has(42 + 64) {
		t.Error("invalid value set")
		return
	}

	s.unset(42)
	if s.has(42) {
		t.Error("failed to unset id")
		return
	}

	for i := 0; i < 256; i++ {
		s.set(i)
		if !s.has(i) {
			t.Error("failed to set id")
			return
		}
	}
}
