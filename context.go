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

func (c *context) pending(offset int, id int) bool {
	if len(c.isPending) <= id {
		return false
	}

	for i := range c.isPending[id] {
		if c.isPending[id][i] == offset {
			return true
		}
	}

	return false
}

func (c *context) markPending(offset int, id int) {
	if len(c.isPending) <= id {
		if cap(c.isPending) > id {
			c.isPending = c.isPending[:id+1]
		} else {
			c.isPending = c.isPending[:cap(c.isPending)]
			for i := cap(c.isPending); i <= id; i++ {
				c.isPending = append(c.isPending, nil)
			}
		}
	}

	for i := range c.isPending[id] {
		if c.isPending[id][i] == -1 {
			c.isPending[id][i] = offset
			return
		}
	}

	c.isPending[id] = append(c.isPending[id], offset)
}

func (c *context) unmarkPending(offset int, id int) {
	for i := range c.isPending[id] {
		if c.isPending[id][i] == offset {
			c.isPending[id][i] = -1
			break
		}
	}
}

func (c *context) fromResults(id int) bool {
	to, m, ok := c.results.getMatch(c.offset, id)
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
}

func (c *context) fail(offset int) {
	c.offset = offset
	c.matchLast = false
}

func (c *context) finalizeParse(rootID int) error {
	if !c.matchLast {
		return ErrInvalidInput
	}

	to, match, found := c.results.getMatch(0, rootID)
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
