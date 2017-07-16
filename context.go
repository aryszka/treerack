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
	isExcluded []*idSet
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
	if len(c.isExcluded) <= offset || c.isExcluded[offset] == nil {
		return false
	}

	return c.isExcluded[offset].has(id)
}

func (c *context) exclude(offset int, id int) {
	if c.excluded(offset, id) {
		return
	}

	if len(c.isExcluded) <= offset {
		c.isExcluded = append(c.isExcluded, nil)
		if cap(c.isExcluded) > offset {
			c.isExcluded = c.isExcluded[:offset+1]
		} else {
			c.isExcluded = append(
				c.isExcluded[:cap(c.isExcluded)],
				make([]*idSet, offset+1-cap(c.isExcluded))...,
			)
		}
	}

	if c.isExcluded[offset] == nil {
		c.isExcluded[offset] = &idSet{}
	}

	c.isExcluded[offset].set(id)
}

func (c *context) include(offset int, id int) {
	if len(c.isExcluded) <= offset || c.isExcluded[offset] == nil {
		return
	}

	c.isExcluded[offset].unset(id)
}

func (c *context) fromStore(id int) (bool, bool) {
	n, m, ok := c.store.get(c.offset, id)
	if !ok {
		return false, false
	}

	if m {
		c.success(n)
	} else {
		c.fail(c.offset)
	}

	return m, true
}

func (c *context) success(n *Node) {
	c.node = n
	c.offset = n.To
	c.match = true
}

func (c *context) successChar() {
	c.node = nil
	c.offset++
	c.match = true
}

func (c *context) fail(offset int) {
	c.offset = offset
	c.match = false
}

func (c *context) finalize() error {
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
