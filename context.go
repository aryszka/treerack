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
	store      *store
	tokens     []rune
	match      bool
	node       *Node
	isExcluded [][]int
}

func newContext(r io.RuneReader) *context {
	return &context{
		reader: r,
		store:  &store{},
	}
}

func (c *context) read() bool {
	if c.eof || c.readErr != nil {
		return false
	}

	t, n, err := c.reader.ReadRune()
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

	if t == unicode.ReplacementChar {
		c.readErr = ErrInvalidUnicodeCharacter
		return false
	}

	c.tokens = append(c.tokens, t)
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

func (c *context) excluded(offset int, id int) bool {
	if len(c.isExcluded) <= id {
		return false
	}

	for i := range c.isExcluded[id] {
		if c.isExcluded[id][i] == offset {
			return true
		}
	}

	return false
}

func (c *context) exclude(offset int, id int) {
	if len(c.isExcluded) <= id {
		if cap(c.isExcluded) > id {
			c.isExcluded = c.isExcluded[:id+1]
		} else {
			c.isExcluded = c.isExcluded[:cap(c.isExcluded)]
			for i := cap(c.isExcluded); i <= id; i++ {
				c.isExcluded = append(c.isExcluded, nil)
			}
		}
	}

	c.isExcluded[id] = append(c.isExcluded[id], offset)
}

func (c *context) include(offset int, id int) {
	for i := range c.isExcluded[id] {
		if c.isExcluded[id][i] == offset {
			c.isExcluded[id] = append(c.isExcluded[id][:i], c.isExcluded[id][i+1:]...)
			break
		}
	}
}

func (c *context) fromStore(id int) (bool, bool) {
	to, m, ok := c.store.getMatch(c.offset, id)
	if !ok {
		return false, false
	}

	if m {
		c.success(to)
	} else {
		c.fail(c.offset)
	}

	return m, true
}

func (c *context) success(to int) {
	c.offset = to
	c.match = true
}

func (c *context) fail(offset int) {
	c.offset = offset
	c.match = false
}

func (c *context) finalize() error {
	return ErrNotImplemented

	if c.node.To < c.readOffset {
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

	c.node.commit(c.tokens)
	return nil
}
