package treerack

import "strconv"

func dropComments(n *Node) *Node {
	ncc := *n
	nc := &ncc
	nc.Nodes = filterNodes(func(n *Node) bool { return n.Name != "comment" }, n.Nodes)
	nc.Nodes = mapNodes(dropComments, nc.Nodes)
	return nc
}

func flagsToCommitType(n []*Node) CommitType {
	var ct CommitType
	for _, ni := range n {
		switch ni.Name {
		case "alias":
			ct |= Alias
		case "ws":
			ct |= Whitespace
		case "nows":
			ct |= NoWhitespace
		case "failpass":
			ct |= FailPass
		case "root":
			ct |= Root
		}
	}

	return ct
}

func toRune(c string) rune {
	return []rune(c)[0]
}

func nodeChar(n *Node) rune {
	s := n.Text()
	if s[0] == '\\' {
		return unescapeChar(toRune(s[1:]))
	}

	return toRune(s)
}

func defineMember(s *Syntax, defaultName string, ct CommitType, n *Node) (string, error) {
	switch n.Name {
	case "symbol":
		return n.Text(), nil
	default:
		return defaultName, defineExpression(s, defaultName, ct, n)
	}
}

func defineClass(s *Syntax, name string, ct CommitType, n []*Node) error {
	var (
		not    bool
		chars  []rune
		ranges [][]rune
	)

	if len(n) > 0 && n[0].Name == "class-not" {
		not, n = true, n[1:]
	}

	for _, c := range n {
		switch c.Name {
		case "class-char":
			chars = append(chars, nodeChar(c))
		case "char-range":
			ranges = append(ranges, []rune{nodeChar(c.Nodes[0]), nodeChar(c.Nodes[1])})
		}
	}

	return s.class(name, ct, not, chars, ranges)
}

func defineCharSequence(s *Syntax, name string, ct CommitType, charNodes []*Node) error {
	var chars []rune
	for _, ci := range charNodes {
		chars = append(chars, nodeChar(ci))
	}

	return s.charSequence(name, ct, chars)
}

func getQuantity(n *Node) (min int, max int, err error) {
	switch n.Name {
	case "count-quantifier":
		min, err = strconv.Atoi(n.Nodes[0].Text())
		if err != nil {
			return
		}

		max = min
	case "range-quantifier":
		min = 0
		max = -1
		for _, rq := range n.Nodes {
			switch rq.Name {
			case "range-from":
				min, err = strconv.Atoi(rq.Text())
				if err != nil {
					return
				}
			case "range-to":
				max, err = strconv.Atoi(rq.Text())
				if err != nil {
					return
				}
			}
		}
	case "one-or-more":
		min, max = 1, -1
	case "zero-or-more":
		min, max = 0, -1
	case "zero-or-one":
		min, max = 0, 1
	}

	return
}

func defineSymbol(s *Syntax, name string, ct CommitType, n *Node) error {
	return s.sequence(name, ct, SequenceItem{Name: n.Text()})
}

func defineSequence(s *Syntax, name string, ct CommitType, n ...*Node) error {
	nows := ct & NoWhitespace
	var items []SequenceItem
	for i, ni := range n {
		var (
			item SequenceItem
			err  error
		)

		defaultName := childName(name, i)
		item.Name, err = defineMember(s, defaultName, Alias|nows, ni.Nodes[0])
		if err != nil {
			return err
		}

		if len(ni.Nodes) == 2 {
			item.Min, item.Max, err = getQuantity(ni.Nodes[1])
			if err != nil {
				return err
			}
		}

		items = append(items, item)
	}

	return s.sequence(name, ct, items...)
}

func defineChoice(s *Syntax, name string, ct CommitType, n ...*Node) error {
	var refs []string
	memberCT := ct&NoWhitespace | Alias
	for i, ni := range n {
		nmi := childName(name, i)
		ref, err := defineMember(s, nmi, memberCT, ni)
		if err != nil {
			return err
		}

		refs = append(refs, ref)
	}

	return s.choice(name, ct, refs...)
}

func defineExpression(s *Syntax, name string, ct CommitType, expression *Node) error {
	var err error
	switch expression.Name {
	case "any-char":
		err = s.anyChar(name, ct)
	case "char-class":
		err = defineClass(s, name, ct, expression.Nodes)
	case "char-sequence":
		err = defineCharSequence(s, name, ct, expression.Nodes)
	case "symbol":
		err = defineSymbol(s, name, ct, expression)
	case "sequence":
		err = defineSequence(s, name, ct, expression.Nodes...)
	case "choice":
		err = defineChoice(s, name, ct, expression.Nodes...)
	}

	return err
}

func defineDefinition(s *Syntax, n *Node) error {
	return defineExpression(
		s,
		n.Nodes[0].Text(),
		flagsToCommitType(n.Nodes[1:len(n.Nodes)-1])|userDefined,
		n.Nodes[len(n.Nodes)-1],
	)
}

func define(s *Syntax, n *Node) error {
	n = dropComments(n)

	for _, ni := range n.Nodes {
		if err := defineDefinition(s, ni); err != nil {
			return err
		}
	}

	return nil
}
