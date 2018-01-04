package treerack

import (
	"fmt"
	"io"
)

const (
	charClassEscape = '\\'
	charClassBanned = "\\[]^-\b\f\n\r\t\v"
)

type charParser struct {
	name   string
	id     int
	not    bool
	chars  []rune
	ranges [][]rune
}

type charBuilder struct {
	name string
	id   int
}

func newChar(
	name string,
	not bool,
	chars []rune,
	ranges [][]rune,
) *charParser {
	return &charParser{
		name:   name,
		not:    not,
		chars:  chars,
		ranges: ranges,
	}
}

func (p *charParser) nodeName() string            { return p.name }
func (p *charParser) setName(n string)            { p.name = n }
func (p *charParser) nodeID() int                 { return p.id }
func (p *charParser) setID(id int)                { p.id = id }
func (p *charParser) commitType() CommitType      { return Alias }
func (p *charParser) setCommitType(ct CommitType) {}
func (p *charParser) preinit()                    {}
func (p *charParser) validate(*registry) error    { return nil }
func (p *charParser) init(*registry)              {}
func (p *charParser) addGeneralization(int)       {}
func (p *charParser) parser() parser              { return p }

func (p *charParser) builder() builder {
	return &charBuilder{
		id:   p.id,
		name: p.name,
	}
}

func (p *charParser) isSingleChar() bool {
	return !p.not && len(p.chars) == 1 && len(p.ranges) == 0
}

func (p *charParser) format(_ *registry, f formatFlags) string {
	if p.not && len(p.chars) == 0 && len(p.ranges) == 0 {
		return "."
	}

	esc := func(c ...rune) []rune {
		return escape(charClassEscape, []rune(charClassBanned), c)
	}

	var s []rune
	s = append(s, '[')

	if p.not {
		s = append(s, '^')
	}

	s = append(s, esc(p.chars...)...)

	for i := range p.ranges {
		s = append(s, esc(p.ranges[i][0])...)
		s = append(s, '-')
		s = append(s, esc(p.ranges[i][1])...)
	}

	s = append(s, ']')
	return string(s)
}

func matchChar(chars []rune, ranges [][]rune, not bool, char rune) bool {
	for _, ci := range chars {
		if ci == char {
			return !not
		}
	}

	for _, ri := range ranges {
		if char >= ri[0] && char <= ri[1] {
			return !not
		}
	}

	return not
}

func (p *charParser) match(t rune) bool {
	return matchChar(p.chars, p.ranges, p.not, t)
}

func (p *charParser) parse(c *context) {
	if tok, ok := c.token(); !ok || !p.match(tok) {
		if c.offset > c.failOffset {
			c.failOffset = c.offset
			c.failingParser = nil
		}

		c.fail(c.offset)
		return
	}

	c.success(c.offset + 1)
}

func (p *charParser) generate(w io.Writer, done map[string]bool) error {
	if done[p.name] {
		return nil
	}

	done[p.name] = true

	var err error
	fprintf := func(f string, args ...interface{}) {
		if err != nil {
			return
		}

		_, err = fmt.Fprintf(w, f, args...)
	}

	fprintf("var p%d = charParser{", p.id)
	fprintf("id: %d, not: %t,", p.id, p.not)

	fprintf("chars: []rune{")
	for i := range p.chars {
		fprintf("%d,", p.chars[i])
	}

	fprintf("},")

	fprintf("ranges: [][]rune{")
	for i := range p.ranges {
		fprintf("{%d, %d},", p.ranges[i][0], p.ranges[i][1])
	}

	fprintf("}};")
	return err
}

func (b *charBuilder) nodeName() string { return b.name }
func (b *charBuilder) nodeID() int      { return b.id }

func (b *charBuilder) build(c *context) ([]*Node, bool) {
	return nil, false
}

func (b *charBuilder) generate(w io.Writer, done map[string]bool) error {
	if done[b.name] {
		return nil
	}

	done[b.name] = true
	_, err := fmt.Fprintf(w, "var b%d = charBuilder{};", b.id)
	return err
}
