package treerack

import (
	"io"
	"strings"
	"unicode"
)

type context struct {
	reader        io.RuneReader
	offset        int
	readOffset    int
	consumed      int
	failOffset    int
	failingParser parser
	readErr       error
	eof           bool
	results       *results
	tokens        []rune
	matchLast     bool
}

func newContext(r io.RuneReader) *context {
	return &context{
		reader:     r,
		results:    &results{},
		failOffset: -1,
	}
}

func (c *context) read() bool {
	if c.eof || c.readErr != nil {
		return false
	}

	token, n, err := c.reader.ReadRune()
	if err != nil {
		if err == io.EOF {
			if n == 0 {
				c.eof = true
				return false
			}
		} else {
			c.readErr = err
			return false
		}
	}

	c.readOffset++

	if token == unicode.ReplacementChar {
		c.readErr = ErrInvalidUnicodeCharacter
		return false
	}

	c.tokens = append(c.tokens, token)
	return true
}

func (c *context) token() (rune, bool) {
	if c.offset == c.readOffset {
		if !c.read() {
			return 0, false
		}
	}

	return c.tokens[c.offset], true
}

func (c *context) fromResults(p parser) bool {
	to, m, ok := c.results.longestResult(c.offset, p.nodeID())
	if !ok {
		return false
	}

	if m {
		c.success(to)
	} else {
		c.fail(c.offset)
	}

	return true
}

func (c *context) success(to int) {
	c.offset = to
	c.matchLast = true
	if to > c.consumed {
		c.consumed = to
	}
}

func (c *context) fail(offset int) {
	c.offset = offset
	c.matchLast = false
}

func findLine(tokens []rune, offset int) (line, column int) {
	tokens = tokens[:offset]
	for i := range tokens {
		column++
		if tokens[i] == '\n' {
			column = 0
			line++
		}
	}

	return
}

func (c *context) parseError(p parser) error {
	definition := p.nodeName()
	flagIndex := strings.Index(definition, ":")
	if flagIndex > 0 {
		definition = definition[:flagIndex]
	}

	if c.failingParser == nil {
		c.failOffset = c.consumed
	}

	line, col := findLine(c.tokens, c.failOffset)

	return &ParseError{
		Offset:     c.failOffset,
		Line:       line,
		Column:     col,
		Definition: definition,
	}
}

func (c *context) finalizeParse(root parser) error {
	p := c.failingParser
	if p == nil {
		p = root
	}

	to, match, found := c.results.longestResult(0, root.nodeID())
	if !found || !match || found && match && to < c.readOffset {
		return c.parseError(p)
	}

	c.read()
	if c.eof {
		return nil
	}

	if c.readErr != nil {
		return c.readErr
	}

	return c.parseError(root)
}
