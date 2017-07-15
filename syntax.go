package treerack

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

type CommitType int

const (
	None  CommitType = 0
	Alias CommitType = 1 << iota
	Documentation
	Root
)

type SequenceItem struct {
	Name     string
	Min, Max int // 0,0 considered as 1,1, x,0 considered as x,-1
}

type Syntax struct {
	trace        Trace
	registry     *registry
	initialized  bool
	initFailed   bool
	explicitRoot bool
	root         definition
	parser       parser
}

var (
	ErrSyntaxInitialized       = errors.New("syntax initialized")
	ErrInitFailed              = errors.New("init failed")
	ErrNoParsersDefined        = errors.New("no parsers defined")
	ErrInvalidInput            = errors.New("invalid input")
	ErrInvalidUnicodeCharacter = errors.New("invalid unicode character")
	ErrInvalidEscapeCharacter  = errors.New("invalid escape character")
	ErrUnexpectedCharacter     = errors.New("unexpected character")
	ErrInvalidSyntax           = errors.New("invalid syntax")
	ErrRootAlias               = errors.New("root node cannot be an alias")
	ErrNotImplemented          = errors.New("not implemented")
)

func duplicateDefinition(name string) error {
	return fmt.Errorf("duplicate definition: %s", name)
}

func NewSyntax() *Syntax {
	return NewSyntaxTrace(nil)
}

func NewSyntaxTrace(t Trace) *Syntax {
	if t == nil {
		t = NopTrace{}
	}

	return &Syntax{
		trace:    t,
		registry: newRegistry(),
	}
}

func (s *Syntax) register(d definition) error {
	if s.initialized {
		return ErrSyntaxInitialized
	}

	if d.commitType()&Root != 0 {
		s.root = d
		s.explicitRoot = true
	} else if !s.explicitRoot {
		s.root = d
	}

	return s.registry.setDefinition(d)
}

func (s *Syntax) AnyChar(name string, ct CommitType) error {
	return s.Class(name, ct, true, nil, nil)
}

func (s *Syntax) Class(name string, ct CommitType, not bool, chars []rune, ranges [][]rune) error {
	return s.register(newChar(name, ct, not, chars, ranges))
}

func childName(name string, childIndex int) string {
	return fmt.Sprintf("%s:%d", name, childIndex)
}

func (s *Syntax) CharSequence(name string, ct CommitType, chars []rune) error {
	var refs []string
	for i, ci := range chars {
		ref := childName(name, i)
		refs = append(refs, ref)
		if err := s.register(newChar(ref, Alias, false, []rune{ci}, nil)); err != nil {
			return err
		}
	}

	return s.Sequence(name, ct, namesToSequenceItems(refs)...)
}

func (s *Syntax) Sequence(name string, ct CommitType, items ...SequenceItem) error {
	return s.register(newSequence(name, ct, items))
}

func (s *Syntax) Choice(name string, ct CommitType, elements ...string) error {
	return s.register(newChoice(name, ct, elements))
}

func (s *Syntax) Read(r io.Reader) error {
	if s.initialized {
		return ErrSyntaxInitialized
	}

	return ErrNotImplemented
}

func (s *Syntax) Init() error {
	if s.initFailed {
		return ErrInitFailed
	}

	if s.initialized {
		return nil
	}

	if s.root == nil {
		return ErrNoParsersDefined
	}

	if s.root.commitType()&Alias != 0 {
		return ErrRootAlias
	}

	var err error
	s.parser, err = s.root.parser(s.registry, nil)
	if err != nil {
		s.initFailed = true
		return err
	}

	s.initialized = true
	return nil
}

func (s *Syntax) Generate(w io.Writer) error {
	if err := s.Init(); err != nil {
		return err
	}

	return ErrNotImplemented
}

func (s *Syntax) Parse(r io.Reader) (*Node, error) {
	if err := s.Init(); err != nil {
		return nil, err
	}

	c := newContext(bufio.NewReader(r))
	return parse(s.trace, s.parser, c)
}
