package treerack

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

var errWriteError = errors.New("write failed")

type failingWriter struct {
	failWhenReceived []string
	buffer           *bytes.Buffer
}

func newFailingWriter(failWhenReceived ...string) io.Writer {
	return &failingWriter{
		failWhenReceived: failWhenReceived,
		buffer:           &bytes.Buffer{},
	}
}

func (w *failingWriter) Write(p []byte) (int, error) {
	n, err := w.buffer.Write(p)
	if err != nil {
		panic(err)
	}

	s := w.buffer.String()
	for i := range w.failWhenReceived {
		if !strings.Contains(s, w.failWhenReceived[i]) {
			return n, nil
		}
	}

	return n, errWriteError
}

func TestGenerate(t *testing.T) {
	s, err := openSyntaxFile("syntax.treerack")
	if err != nil {
		t.Fatal(err)
	}

	var b bytes.Buffer
	if err := s.Generate(GeneratorOptions{PackageName: "foo"}, &b); err != nil {
		t.Error(err)
	}
}

func TestGenerateFailingWrite(t *testing.T) {
	s, err := openSyntaxFile("syntax.treerack")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("char parser", func(t *testing.T) {
		if err := s.Generate(GeneratorOptions{PackageName: "foo"}, newFailingWriter("= charParser")); err == nil {
			t.Error("failed to fail")
		}
	})

	t.Run("char builder", func(t *testing.T) {
		if err := s.Generate(GeneratorOptions{PackageName: "foo"}, newFailingWriter("= charBuilder")); err == nil {
			t.Error("failed to fail")
		}
	})

	t.Run("sequence parser", func(t *testing.T) {
		if err := s.Generate(GeneratorOptions{PackageName: "foo"}, newFailingWriter("= sequenceParser")); err == nil {
			t.Error("failed to fail")
		}
	})

	t.Run("sequence builder", func(t *testing.T) {
		if err := s.Generate(GeneratorOptions{PackageName: "foo"}, newFailingWriter("= sequenceBuilder")); err == nil {
			t.Error("failed to fail")
		}
	})

	t.Run("choice parser", func(t *testing.T) {
		if err := s.Generate(GeneratorOptions{PackageName: "foo"}, newFailingWriter("= choiceParser")); err == nil {
			t.Error("failed to fail")
		}
	})

	t.Run("choice builder", func(t *testing.T) {
		if err := s.Generate(GeneratorOptions{PackageName: "foo"}, newFailingWriter("= choiceBuilder")); err == nil {
			t.Error("failed to fail")
		}
	})
}

func TestGenerateFailOnInit(t *testing.T) {
	s := &Syntax{}
	s.Choice("a", None, "b") // undefined b
	if err := s.Generate(GeneratorOptions{PackageName: "foo"}, &bytes.Buffer{}); err == nil {
		t.Error("failed to fail")
	}
}

func TestGenerateDefaulPackageName(t *testing.T) {
	s, err := openSyntaxFile("syntax.treerack")
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := s.Generate(GeneratorOptions{}, &buf); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(buf.String(), "\npackage main\n") {
		t.Error("failed to set default package name")
	}
}
