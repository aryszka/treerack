package treerack

import (
	"io"
	"unicode"
)

type context struct {
	reader     io.RuneReader
	offset     int
	readOffset int
	readErr    error
	eof        bool
	results    *results
	tokens     []rune
	matchLast  bool
	isPending  [][]int
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

func (c *context) fromResults(id int) bool {
	to, m, ok := c.results.longestResult(c.offset, id)
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

// TODO:
// - try to move this to the parsers
// - try to move more
// - if doens't help performance, try move more from there to here

func (c *context) success(to int) {
	c.offset = to
	c.matchLast = true
}

func (c *context) fail(offset int) {
	c.offset = offset
	c.matchLast = false
}

func (c *context) finalizeParse(rootID int) error {
	if !c.matchLast {
		return ErrInvalidInput
	}

	to, match, found := c.results.longestResult(0, rootID)
	if !found || !match || to < c.readOffset {
		return ErrUnexpectedCharacter
	}

	if !c.eof {
		c.read()
		if !c.eof {
			if c.readErr != nil {
				return c.readErr
			}

			return ErrUnexpectedCharacter
		}
	}

	return nil
}
