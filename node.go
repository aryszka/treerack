package parse

import "fmt"

type Node struct {
	Name       string
	Nodes      []*Node
	commitType CommitType
	from, to   int
	tokens     []rune
}

func newNode(name string, ct CommitType, from, to int) *Node {
	return &Node{
		Name:       name,
		commitType: ct,
		from:       from,
		to:         to,
	}
}

func (n *Node) tokenLength() int {
	return n.to - n.from
}

func (n *Node) nodeLength() int {
	return len(n.Nodes)
}

func findNode(in, n *Node) {
	if n == in {
		panic(fmt.Errorf("found self in %s", in.Name))
	}

	for _, ni := range n.Nodes {
		findNode(in, ni)
	}
}

func (n *Node) append(p *Node) {
	findNode(n, p)
	n.Nodes = append(n.Nodes, p)
	// TODO: check rather if n.from <= p.from??? or panic if less? or check rather node length and commit
	// happens in the end anyway?
	if n.from == 0 && n.to == 0 {
		n.from = p.from
	}

	n.to = p.to
}

func (n *Node) clear() {
	n.from = 0
	n.to = 0
	n.Nodes = nil
}

func (n *Node) applyTokens(t []rune) {
	n.tokens = t
	for _, ni := range n.Nodes {
		ni.applyTokens(t)
	}
}

func (n *Node) commit() {
	var nodes []*Node
	for _, ni := range n.Nodes {
		ni.commit()
		if ni.commitType&Alias != 0 {
			nodes = append(nodes, ni.Nodes...)
		} else {
			nodes = append(nodes, ni)
		}
	}

	n.Nodes = nodes
}

func (n *Node) String() string {
	if n.from >= len(n.tokens) || n.to > len(n.tokens) {
		return n.Name + ":incomplete"
	}

	return fmt.Sprintf("%s:%d:%d:%s", n.Name, n.from, n.to, n.Text())
}

func (n *Node) Text() string {
	return string(n.tokens[n.from:n.to])
}
