package treerack

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

type failingReader struct {
	input     []byte
	failIndex int
	index     int
}

func (fr *failingReader) Read(p []byte) (int, error) {
	if fr.index == fr.failIndex {
		return 0, errors.New("test error")
	}

	if len(fr.input) <= fr.index {
		return 0, io.EOF
	}

	available := fr.input[fr.index:]
	copy(p[:1], available)
	fr.index++
	return 1, nil
}

func TestFailingRead(t *testing.T) {
	s := &Syntax{}
	if err := s.AnyChar("A", None); err != nil {
		t.Error(err)
		return
	}

	t.Run("reader error", func(t *testing.T) {
		r := &failingReader{}
		if _, err := s.Parse(r); err == nil {
			t.Error("failed to fail")
		}
	})

	t.Run("invalid unicode", func(t *testing.T) {
		r := bytes.NewBuffer([]byte{255, 255})
		if _, err := s.Parse(r); err == nil {
			t.Error("failed to fail")
		}
	})

	t.Run("fail during finalize", func(t *testing.T) {
		r := &failingReader{
			input:     []byte("aa"),
			failIndex: 1,
		}

		s = &Syntax{}

		if err := s.Class("a", Root, false, []rune("a"), nil); err != nil {
			t.Error(err)
		}

		if _, err := s.Parse(r); err == nil {
			t.Error("failed to fail")
		}
	})
}
