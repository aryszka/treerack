package treerack

import (
	"io"
	"strings"
	"unicode"
)

type context struct {
	reader        io.RuneReader
	keywords      []parser
	offset        int
	readOffset    int
	consumed      int
	offsetLimit   int
	failOffset    int
	failingParser parser
	readErr       error
	eof           bool
	results       *results
	tokens        []rune
	matchLast     bool
}

func newContext(r io.RuneReader, keywords []parser) *context {
	return &context{
		reader:      r,
		keywords:    keywords,
		results:     &results{},
		offsetLimit: -1,
		failOffset:  -1,
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
	if c.offset == c.offsetLimit {
		return 0, false
	}

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

func (c *context) isKeyword(from, to int) bool {
	ol := c.offsetLimit
	c.offsetLimit = to
	defer func() { c.offsetLimit = ol }()
	for _, kw := range c.keywords {
		c.offset = from
		kw.parse(c)
		if c.matchLast && c.offset == to {
			return true
		}
	}

	return false
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
	fp := c.failingParser
	if fp == nil {
		fp = root
	}

	to, match, found := c.results.longestResult(0, root.nodeID())
	if !found || !match || found && match && to < c.readOffset {
		return c.parseError(fp)
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
