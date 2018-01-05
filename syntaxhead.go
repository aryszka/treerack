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

var ErrInvalidUnicodeCharacter = errors.New("invalid unicode character")

func (pe *ParseError) Error() string {
	return fmt.Sprintf(
		"%s:%d:%d:parse failed, parsing: %s",
		pe.Input,
		pe.Line+1,
		pe.Column+1,
		pe.Definition,
	)
}

func parse(r io.Reader, p parser, b builder) (*Node, error) {
	c := newContext(bufio.NewReader(r))
	p.parse(c)
	if c.readErr != nil {
		return nil, c.readErr
	}

	if err := c.finalizeParse(p); err != nil {
		if perr, ok := err.(*ParseError); ok {
			perr.Input = "<input>"
		}

		return nil, err
	}

	c.offset = 0
	c.results.resetPending()

	n, _ := b.build(c)
	return n[0], nil
}
