package treerack

import (
	"os"
	"testing"
	"time"
)

func TestBoot(t *testing.T) {
	b, err := initBoot(bootSyntaxDefs)
	if err != nil {
		t.Error(err)
		return
	}

	f, err := os.Open("syntax.parser")
	if err != nil {
		t.Error(err)
		return
	}

	defer f.Close()

	var d time.Duration
	const n = 120
	var n0 *Node
	for i := 0; i < n; i++ {
		if _, err := f.Seek(0, 0); err != nil {
			t.Error(err)
			return
		}

		start := time.Now()
		n0, err = b.Parse(f)
		d += time.Now().Sub(start)

		if err != nil {
			t.Error(err)
			return
		}
	}

	t.Log("duration:", d/n)

	s0 := NewSyntax()
	if err := define(s0, n0); err != nil {
		t.Error(err)
		return
	}

	_, err = f.Seek(0, 0)
	if err != nil {
		t.Error(err)
		return
	}

	err = s0.Init()
	if err != nil {
		t.Error(err)
		return
	}

	n1, err := s0.Parse(f)
	if err != nil {
		t.Error(err)
		return
	}

	checkNode(t, n1, n0)
	if t.Failed() {
		return
	}

	s1 := NewSyntax()
	if err := define(s1, n1); err != nil {
		t.Error(err)
		return
	}

	_, err = f.Seek(0, 0)
	if err != nil {
		t.Error(err)
		return
	}

	n2, err := s1.Parse(f)
	if err != nil {
		t.Error(err)
		return
	}

	checkNode(t, n2, n1)
}
