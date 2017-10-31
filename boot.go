package treerack

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

var errInvalidDefinition = errors.New("invalid syntax definition")

func stringToCommitType(s string) CommitType {
	switch s {
	case "alias":
		return Alias
	case "ws":
		return Whitespace
	case "nows":
		return NoWhitespace
	case "doc":
		return Documentation
	case "root":
		return Root
	default:
		return None
	}
}

func checkBootDefinitionLength(d []string) error {
	if len(d) < 3 {
		return errInvalidDefinition
	}

	switch d[0] {
	case "chars", "class", "sequence", "choice":
		if len(d) < 4 {
			return errInvalidDefinition
		}
	}

	return nil
}

func parseClass(class []rune) (not bool, chars []rune, ranges [][]rune, err error) {
	if class[0] == '^' {
		not = true
		class = class[1:]
	}

	for {
		if len(class) == 0 {
			return
		}

		var c0 rune
		c0, class = class[0], class[1:]
		switch c0 {
		case '[', ']', '^', '-':
			err = errInvalidDefinition
			return
		}

		if c0 == '\\' {
			if len(class) == 0 {
				err = errInvalidDefinition
				return
			}

			c0, class = unescapeChar(class[0]), class[1:]
		}

		if len(class) < 2 || class[0] != '-' {
			chars = append(chars, c0)
			continue
		}

		var c1 rune
		c1, class = class[1], class[2:]
		switch c1 {
		case '[', ']', '^', '-':
			err = errInvalidDefinition
			return
		}

		if c1 == '\\' {
			if len(class) == 0 {
				err = errInvalidDefinition
				return
			}

			c1, class = unescapeChar(class[0]), class[1:]
		}

		ranges = append(ranges, []rune{c0, c1})
	}
}

func defineBootAnything(s *Syntax, d []string) error {
	ct := stringToCommitType(d[2])
	return s.anyChar(d[1], ct)
}

func defineBootClass(s *Syntax, d []string) error {
	name := d[1]
	ct := stringToCommitType(d[2])

	not, chars, ranges, err := parseClass([]rune(d[3]))
	if err != nil {
		return err
	}

	return s.class(name, ct, not, chars, ranges)
}

func defineBootCharSequence(s *Syntax, d []string) error {
	name := d[1]
	ct := stringToCommitType(d[2])

	chars, err := unescapeCharSequence(d[3])
	if err != nil {
		return err
	}

	return s.charSequence(name, ct, chars)
}

func splitQuantifiedSymbol(s string) (string, int, int) {
	ssplit := strings.Split(s, ":")
	if len(ssplit) != 3 {
		return s, 0, 0
	}

	name := ssplit[0]

	min, err := strconv.Atoi(ssplit[1])
	if err != nil {
		panic(err)
	}

	max, err := strconv.Atoi(ssplit[2])
	if err != nil {
		panic(err)
	}

	return name, min, max
}

func namesToSequenceItemsQuantify(n []string) []SequenceItem {
	si := make([]SequenceItem, len(n))
	for i, ni := range n {
		name, min, max := splitQuantifiedSymbol(ni)
		si[i] = SequenceItem{Name: name, Min: min, Max: max}
	}

	return si
}

func defineBootSequence(s *Syntax, defs []string) error {
	name := defs[1]
	ct := stringToCommitType(defs[2])
	items := namesToSequenceItemsQuantify(defs[3:])
	return s.sequence(name, ct, items...)
}

func defineBootChoice(s *Syntax, defs []string) error {
	name := defs[1]
	ct := stringToCommitType(defs[2])
	items := defs[3:]
	return s.choice(name, ct, items...)
}

func defineBoot(s *Syntax, defs []string) error {
	switch defs[0] {
	case "anything":
		return defineBootAnything(s, defs)
	case "class":
		return defineBootClass(s, defs)
	case "chars":
		return defineBootCharSequence(s, defs)
	case "sequence":
		return defineBootSequence(s, defs)
	case "choice":
		return defineBootChoice(s, defs)
	default:
		return errInvalidDefinition
	}
}

func defineAllBoot(s *Syntax, defs [][]string) error {
	for _, d := range defs {
		if err := defineBoot(s, d); err != nil {
			return err
		}
	}

	return nil
}

func createBoot() (*Syntax, error) {
	s := &Syntax{}
	if err := defineAllBoot(s, bootSyntaxDefs); err != nil {
		return nil, err
	}

	return s, s.Init()
}

func bootSyntax() (*Syntax, error) {
	b, err := createBoot()
	if err != nil {
		return nil, err
	}

	f, err := os.Open("syntax.treerack")
	if err != nil {
		return nil, err
	}

	defer f.Close()

	doc, err := b.Parse(f)
	if err != nil {
		return nil, err
	}

	s := &Syntax{}
	return s, define(s, doc)
}
