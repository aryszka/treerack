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
	Whitespace
	NoWhitespace
	FailPass
	Root

	userDefined
)

type formatFlags int

const (
	formatNone   formatFlags = 0
	formatPretty formatFlags = 1 << iota
	formatIncludeComments
)

// if min=0&&max=0, it means min=1,max=1
// else if max<=0, it means no max
// else if min<=0, it means no min
type SequenceItem struct {
	Name     string
	Min, Max int
}

// ParseError is returned when the input text doesn't match
// the used syntax during parsing.
type ParseError struct {

	// Input is the name of the input file or <input> if not
	// available.
	Input string

	// Offset is the index of the right-most failing
	// token in the input text.
	Offset int

	// Line tells the line index of the right-most failing
	// token in the input text.
	//
	// It is zero-based, and for error reporting, it is
	// recommended to increment it by one.
	Line int

	// Column tells the column index of the right-most failing
	// token in the input text.
	Column int

	// Definition tells the right-most unmatched parser definition.
	Definition string

	registry *registry
}

type Syntax struct {
	registry     *registry
	initialized  bool
	initFailed   bool
	explicitRoot bool
	root         definition
	parser       parser
	builder      builder
}

type definition interface {
	nodeName() string
	setName(string)
	nodeID() int
	setID(int)
	commitType() CommitType
	setCommitType(CommitType)
	preinit()
	validate(*registry) error
	init(*registry)
	addGeneralization(int)
	parser() parser
	builder() builder
	format(*registry, formatFlags) string
}

type parser interface {
	nodeName() string
	nodeID() int
	commitType() CommitType
	parse(*context)
}

type builder interface {
	nodeName() string
	nodeID() int
	build(*context) ([]*Node, bool)
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
	ErrRootWhitespace          = errors.New("root node cannot be a whitespace")
	ErrRootFailPass            = errors.New("root node cannot pass failing definition")
	ErrNotImplemented          = errors.New("not implemented")
	ErrMultipleRoots           = errors.New("multiple roots")
	ErrInvalidSymbolName       = errors.New("invalid symbol name")
)

func duplicateDefinition(name string) error {
	return fmt.Errorf("duplicate definition: %s", name)
}

func parserNotFound(name string) error {
	return fmt.Errorf("parser not found: %s", name)
}

const symbolChars = "^\\\\ \\n\\t\\b\\f\\r\\v/.\\[\\]\\\"{}\\^+*?|():=;"

func parseClass(class []rune) (not bool, chars []rune, ranges [][]rune, err error) {
	if class[0] == '^' {
		not = true
		class = class[1:]
	}

	for {
		if len(class) == 0 {
			return
		}

		var c0 rune
		c0, class = class[0], class[1:]

		/*
			this doesn't happen:
			switch c0 {
			case '[', ']', '^', '-':
				err = errInvalidDefinition
				return
			}
		*/

		if c0 == '\\' {
			/*
				this doesn't happen:
				if len(class) == 0 {
					err = errInvalidDefinition
					return
				}
			*/

			c0, class = unescapeChar(class[0]), class[1:]
		}

		if len(class) < 2 || class[0] != '-' {
			chars = append(chars, c0)
			continue
		}

		var c1 rune
		c1, class = class[1], class[2:]

		/*
			this doesn't happen:
			switch c1 {
			case '[', ']', '^', '-':
				err = errInvalidDefinition
				return
			}

			if c1 == '\\' {
				if len(class) == 0 {
					err = errInvalidDefinition
					return
				}

				c1, class = unescapeChar(class[0]), class[1:]
			}
		*/

		ranges = append(ranges, []rune{c0, c1})
	}
}

func parseSymbolChars(c []rune) []rune {
	_, chars, _, _ := parseClass(c)
	return chars
}

var symbolCharRunes = parseSymbolChars([]rune(symbolChars))

func isValidSymbol(n string) bool {
	runes := []rune(n)
	for _, r := range runes {
		if !matchChar(symbolCharRunes, nil, true, r) {
			return false
		}
	}

	return true

}

func intsContain(is []int, i int) bool {
	for _, ii := range is {
		if ii == i {
			return true
		}
	}

	return false
}

func (pe *ParseError) Error() string {
	return fmt.Sprintf(
		"%s:%d:%d:parse failed, parsing: %s",
		pe.Input,
		pe.Line+1,
		pe.Column+1,
		pe.Definition,
	)
}

// func (pe *ParseError) Verbose() string {
// 	return ""
// }

func (s *Syntax) applyRoot(d definition) error {
	explicitRoot := d.commitType()&Root != 0
	if explicitRoot && s.explicitRoot {
		return ErrMultipleRoots
	}

	if s.root != nil && (explicitRoot || !s.explicitRoot) {
		s.root.setCommitType(s.root.commitType() &^ Root)
	}

	if explicitRoot || !s.explicitRoot {
		s.root = d
		s.root.setCommitType(s.root.commitType() | Root)
	}

	if explicitRoot {
		s.explicitRoot = true
	}

	return nil
}

func (s *Syntax) register(d definition) error {
	if s.initialized {
		return ErrSyntaxInitialized
	}

	if s.registry == nil {
		s.registry = newRegistry()
	}

	if err := s.applyRoot(d); err != nil {
		return err
	}

	return s.registry.setDefinition(d)
}

func (s *Syntax) anyChar(name string, ct CommitType) error {
	return s.class(name, ct, true, nil, nil)
}

func (s *Syntax) AnyChar(name string, ct CommitType) error {
	if !isValidSymbol(name) {
		return ErrInvalidSymbolName
	}

	return s.anyChar(name, ct|userDefined)
}

func childName(name string, childIndex int) string {
	return fmt.Sprintf("%s:%d", name, childIndex)
}

func namesToSequenceItems(n []string) []SequenceItem {
	si := make([]SequenceItem, len(n))
	for i := range n {
		si[i] = SequenceItem{Name: n[i]}
	}

	return si
}

func (s *Syntax) class(name string, ct CommitType, not bool, chars []rune, ranges [][]rune) error {
	cname := childName(name, 0)
	if err := s.register(newChar(cname, not, chars, ranges)); err != nil {
		return err
	}

	return s.sequence(name, ct, SequenceItem{Name: cname})
}

func (s *Syntax) Class(name string, ct CommitType, not bool, chars []rune, ranges [][]rune) error {
	if !isValidSymbol(name) {
		return ErrInvalidSymbolName
	}

	return s.class(name, ct|userDefined, not, chars, ranges)
}

func (s *Syntax) charSequence(name string, ct CommitType, chars []rune) error {
	var refs []string
	for i, ci := range chars {
		ref := childName(name, i)
		refs = append(refs, ref)
		if err := s.register(newChar(ref, false, []rune{ci}, nil)); err != nil {
			return err
		}
	}

	return s.sequence(name, ct|NoWhitespace, namesToSequenceItems(refs)...)
}

func (s *Syntax) CharSequence(name string, ct CommitType, chars []rune) error {
	if !isValidSymbol(name) {
		return ErrInvalidSymbolName
	}

	return s.charSequence(name, ct|userDefined, chars)
}

func (s *Syntax) sequence(name string, ct CommitType, items ...SequenceItem) error {
	return s.register(newSequence(name, ct, items))
}

func (s *Syntax) Sequence(name string, ct CommitType, items ...SequenceItem) error {
	if !isValidSymbol(name) {
		return ErrInvalidSymbolName
	}

	return s.sequence(name, ct|userDefined, items...)
}

func (s *Syntax) choice(name string, ct CommitType, options ...string) error {
	return s.register(newChoice(name, ct, options))
}

func (s *Syntax) Choice(name string, ct CommitType, options ...string) error {
	if !isValidSymbol(name) {
		return ErrInvalidSymbolName
	}

	return s.choice(name, ct|userDefined, options...)
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

	if s.root.commitType()&Whitespace != 0 {
		return ErrRootWhitespace
	}

	if s.root.commitType()&FailPass != 0 {
		return ErrRootFailPass
	}

	defs := s.registry.getDefinitions()
	for i := range defs {
		defs[i].preinit()
	}

	if hasWhitespace(defs) {
		defs, s.root = applyWhitespace(defs)
		s.registry = newRegistry(defs...)
	}

	if err := s.root.validate(s.registry); err != nil {
		s.initFailed = true
		return err
	}

	s.root.init(s.registry)
	s.parser = s.root.parser()
	s.builder = s.root.builder()

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
	s.parser.parse(c)
	if c.readErr != nil {
		return nil, c.readErr
	}

	if err := c.finalizeParse(s.parser); err != nil {
		if perr, ok := err.(*ParseError); ok {
			perr.Input = "<input>"
			perr.registry = s.registry
		}

		return nil, err
	}

	c.offset = 0
	c.results.resetPending()

	n, _ := s.builder.build(c)
	return n[0], nil
}
