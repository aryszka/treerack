package treerack

const (
	charClassEscape = '\\'
	charClassBanned = "\\[]^-\b\f\n\r\t\v"
)

type charParser struct {
	name            string
	id              int
	not             bool
	chars           []rune
	ranges          [][]rune
	generalizations []int
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
func (p *charParser) builder() builder            { return p }

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

func matchChars(chars []rune, ranges [][]rune, not bool, char rune) bool {
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
	return matchChars(p.chars, p.ranges, p.not, t)
}

func (p *charParser) parse(c *context) {
	if tok, ok := c.token(); !ok || !p.match(tok) {
		if c.offset > c.failOffset {
			c.failOffset = c.offset
			// println("clearing failing parser")
			c.failingParser = nil
		}

		c.fail(c.offset)
		return
	}

	c.success(c.offset + 1)
}

func (p *charParser) build(c *context) ([]*Node, bool) {
	return nil, false
}
