package treerack

import (
	"fmt"
	"io"
)

const (
	charClassEscape = '\\'
	charClassBanned = "\\[]^-\b\f\n\r\t\v"
)

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

func (p *charParser) setName(n string)            { p.name = n }
func (p *charParser) setID(id int)                { p.id = id }
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
	fprintf("id: %d,", p.id)
	if p.not {
		fprintf("not: true,")
	}

	if len(p.chars) > 0 {
		fprintf("chars: []rune{")
		for i := range p.chars {
			fprintf("%d,", p.chars[i])
		}

		fprintf("},")
	}

	if len(p.ranges) > 0 {
		fprintf("ranges: [][]rune{")
		for i := range p.ranges {
			fprintf("{%d, %d},", p.ranges[i][0], p.ranges[i][1])
		}

		fprintf("},")
	}

	fprintf("};")
	return err
}

func (b *charBuilder) generate(w io.Writer, done map[string]bool) error {
	if done[b.name] {
		return nil
	}

	done[b.name] = true
	_, err := fmt.Fprintf(w, "var b%d = charBuilder{};", b.id)
	return err
}
