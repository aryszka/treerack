package treerack

import (
	"io"
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
		reader:  r,
		results: &results{},
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
		c.fail(p, c.offset)
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

func (c *context) fail(p parser, offset int) {
	c.offset = offset
	c.matchLast = false
	if c.failingParser == nil || c.consumed > c.failOffset {
		// TODO: choice can be retried
		println("setting fail", p.nodeName(), c.failingParser == nil, c.failOffset, c.consumed)
		c.failOffset = c.consumed
		if p.commitType()&userDefined != 0 {
			c.failingParser = p
		}
	}
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

func (c *context) parseError(root parser) error {
	definition := root.nodeName()
	if c.failingParser == nil {
		println("setting fail", c.failOffset, c.consumed)
		c.failOffset = c.consumed
	} else {
		definition = c.failingParser.nodeName()
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
	if !c.matchLast {
		return c.parseError(root)
	}

	to, match, found := c.results.longestResult(0, root.nodeID())

	// TODO: test all three cases
	if !found || !match || to < c.readOffset {
		return c.parseError(root)
	}

	if !c.eof {
		c.read()
		if !c.eof {
			if c.readErr != nil {
				return c.readErr
			}

			return c.parseError(root)
		}
	}

	return nil
}
