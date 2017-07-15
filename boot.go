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
	case "chars", "class":
		if len(d) < 4 {
			return errInvalidDefinition
		}

	case "sequence", "choice":
		if len(d) < 4 {
			return errInvalidDefinition
		}
	}

	return nil
}

func parseClass(c []rune) (not bool, chars []rune, ranges [][]rune, err error) {
	if c[0] == '^' {
		not = true
		c = c[1:]
	}

	for {
		if len(c) == 0 {
			return
		}

		var c0 rune
		c0, c = c[0], c[1:]
		switch c0 {
		case '[', ']', '^', '-':
			err = errInvalidDefinition
			return
		}

		if c0 == '\\' {
			if len(c) == 0 {
				err = errInvalidDefinition
				return
			}

			c0, c = unescapeChar(c[0]), c[1:]
		}

		if len(c) < 2 || c[0] != '-' {
			chars = append(chars, c0)
			continue
		}

		var c1 rune
		c1, c = c[1], c[2:]
		if c1 == '\\' {
			if len(c) == 0 {
				err = errInvalidDefinition
				return
			}

			c1, c = unescapeChar(c[0]), c[1:]
		}

		ranges = append(ranges, []rune{c0, c1})
	}
}

func defineBootAnything(s *Syntax, d []string) error {
	ct := stringToCommitType(d[2])
	return s.AnyChar(d[1], ct)
}

func defineBootClass(s *Syntax, d []string) error {
	ct := stringToCommitType(d[2])

	not, chars, ranges, err := parseClass([]rune(d[3]))
	if err != nil {
		return err
	}

	return s.Class(d[1], ct, not, chars, ranges)
}

func defineBootCharSequence(s *Syntax, d []string) error {
	ct := stringToCommitType(d[2])

	chars, err := unescape('\\', []rune{'"', '\\'}, []rune(d[3]))
	if err != nil {
		return err
	}

	return s.CharSequence(d[1], ct, chars)
}

func namesToSequenceItemsQuantify(n []string, quantify bool) []SequenceItem {
	si := make([]SequenceItem, len(n))
	for i, ni := range n {
		var min, max int
		if quantify {
			nis := strings.Split(ni, ":")
			if len(nis) == 3 {
				ni = nis[0]

				var err error

				min, err = strconv.Atoi(nis[1])
				if err != nil {
					panic(err)
				}

				max, err = strconv.Atoi(nis[2])
				if err != nil {
					panic(err)
				}
			}
		}

		si[i] = SequenceItem{Name: ni, Min: min, Max: max}
	}

	return si
}

func namesToSequenceItems(n []string) []SequenceItem {
	return namesToSequenceItemsQuantify(n, false)
}

func defineBootSequence(s *Syntax, d []string) error {
	ct := stringToCommitType(d[2])
	return s.Sequence(d[1], ct, namesToSequenceItemsQuantify(d[3:], true)...)
}

func defineBootChoice(s *Syntax, d []string) error {
	ct := stringToCommitType(d[2])
	return s.Choice(d[1], ct, d[3:]...)
}

func defineBoot(s *Syntax, d []string) error {
	switch d[0] {
	case "anything":
		return defineBootAnything(s, d)
	case "class":
		return defineBootClass(s, d)
	case "chars":
		return defineBootCharSequence(s, d)
	case "sequence":
		return defineBootSequence(s, d)
	case "choice":
		return defineBootChoice(s, d)
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

func initBoot(definitions [][]string) (*Syntax, error) {
	s := NewSyntax()
	if err := defineAllBoot(s, definitions); err != nil {
		return nil, err
	}

	return s, s.Init()
}

func bootSyntax() (*Syntax, error) {
	b, err := initBoot(bootSyntaxDefs)
	if err != nil {
		return nil, err
	}

	f, err := os.Open("syntax.parser")
	if err != nil {
		return nil, err
	}

	defer f.Close()

	doc, err := b.Parse(f)
	if err != nil {
		return nil, err
	}

	s := NewSyntax()
	return s, define(s, doc)
}
