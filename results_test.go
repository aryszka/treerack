package treerack

import "testing"

func TestResults(t *testing.T) {
	t.Run("set no match when already has match", func(t *testing.T) {
		r := &results{}
		r.setMatch(0, 0, 1)
		r.setNoMatch(0, 0)
		if !r.hasMatchTo(0, 0, 1) {
			t.Error("set no match for an existing match")
		}
	})

	t.Run("check match for a non-existing offset", func(t *testing.T) {
		r := &results{}
		if r.hasMatchTo(1, 0, 1) {
			t.Error("found a non-existing match")
		}
	})
}

func TestPendingWithinCap(t *testing.T) {
	r := &results{}

	t.Run("parse", func(t *testing.T) {
		for i := 0; i < 16; i++ {
			r.markPending(0, i)
		}

		for i := 0; i < 16; i++ {
			if !r.pending(0, i) {
				t.Error("failed to mark pending")
			}
		}
	})

	r.resetPending()

	t.Run("parse", func(t *testing.T) {
		for i := 0; i < 16; i++ {
			r.markPending(0, i)
		}

		for i := 0; i < 16; i++ {
			if !r.pending(0, i) {
				t.Error("failed to mark build pending")
			}
		}
	})
}
