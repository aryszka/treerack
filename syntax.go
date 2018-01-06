package treerack

import (
	"errors"
	"fmt"
	"io"

	"github.com/aryszka/treerack/self"
)

// if min=0&&max=0, it means min=1,max=1
// else if max<=0, it means no max
// else if min<=0, it means no min
type SequenceItem struct {
	Name     string
	Min, Max int
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

type GeneratorOptions struct {
	PackageName string
}

// applied in a non-type-checked way
type generator interface {
	generate(io.Writer, map[string]bool) error
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

var (
	ErrSyntaxInitialized      = errors.New("syntax initialized")
	ErrInitFailed             = errors.New("init failed")
	ErrNoParsersDefined       = errors.New("no parsers defined")
	ErrInvalidEscapeCharacter = errors.New("invalid escape character")
	ErrRootAlias              = errors.New("root node cannot be an alias")
	ErrRootWhitespace         = errors.New("root node cannot be a whitespace")
	ErrRootFailPass           = errors.New("root node cannot pass failing definition")
	ErrMultipleRoots          = errors.New("multiple roots")
	ErrInvalidSymbolName      = errors.New("invalid symbol name")
)

func duplicateDefinition(name string) error {
	return fmt.Errorf("duplicate definition: %s", name)
}

func parserNotFound(name string) error {
	return fmt.Errorf("parser not found: %s", name)
}

var symbolChars = []rune("\\ \n\t\b\f\r\v/.[]\"{}^+*?|():=;")

func isValidSymbol(n string) bool {
	runes := []rune(n)
	for _, r := range runes {
		if !matchChar(symbolChars, nil, true, r) {
			return false
		}
	}

	return true

}

// func (pe *ParseError) Verbose() string {
// 	return ""
// }

func intsContain(is []int, i int) bool {
	for _, ii := range is {
		if ii == i {
			return true
		}
	}

	return false
}

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

func (s *Syntax) ReadSyntax(r io.Reader) error {
	if s.initialized {
		return ErrSyntaxInitialized
	}

	sn, err := self.Parse(r)
	if err != nil {
		return err
	}

	n := mapSelfNode(sn)
	return define(s, n)
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

func (s *Syntax) Generate(o GeneratorOptions, w io.Writer) error {
	if err := s.Init(); err != nil {
		return err
	}

	if o.PackageName == "" {
		o.PackageName = "main"
	}

	var err error
	fprintf := func(f string, args ...interface{}) {
		if err != nil {
			return
		}

		_, err = fmt.Fprintf(w, f, args...)
	}

	fprint := func(args ...interface{}) {
		if err != nil {
			return
		}

		_, err = fmt.Fprint(w, args...)
	}

	fprintln := func() {
		fprint("\n")
	}

	fprint(gendoc)
	fprintln()
	fprintln()

	fprintf("package %s", o.PackageName)
	fprintln()
	fprintln()

	// generate headCode with scripts/createhead.go
	fprint(headCode)
	fprintln()
	fprintln()

	fprint(`func Parse(r io.Reader) (*Node, error) {`)
	fprintln()

	done := make(map[string]bool)
	if err := s.parser.(generator).generate(w, done); err != nil {
		return err
	}

	done = make(map[string]bool)
	if err := s.builder.(generator).generate(w, done); err != nil {
		return err
	}

	fprintln()
	fprintln()
	fprintf(`return parse(r, &p%d, &b%d)`, s.parser.nodeID(), s.builder.nodeID())
	fprintln()
	fprint(`}`)
	fprintln()

	return nil
}

func (s *Syntax) Parse(r io.Reader) (*Node, error) {
	if err := s.Init(); err != nil {
		return nil, err
	}

	return parse(r, s.parser, s.builder)
}
