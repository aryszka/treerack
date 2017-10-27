package treerack

import (
	"os"
	"testing"
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

	n0, err := b.Parse(f)
	if err != nil {
		t.Error(err)
		return
	}

	s0 := NewSyntax()
	if err := define(s0, n0); err != nil {
		t.Error(err)
		return
	}
	
	if err := s0.Init(); err != nil {
		t.Error(err)
		return
	}

	if _, err := f.Seek(0, 0); err != nil {
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
