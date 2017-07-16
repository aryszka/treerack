package treerack

import "fmt"

type Node struct {
	Name       string
	id         int
	Nodes      []*Node
	From, To   int
	commitType CommitType
	tokens     []rune
}

func newNode(name string, id int, from, to int, ct CommitType) *Node {
	return &Node{
		Name:       name,
		id:         id,
		From:       from,
		To:         to,
		commitType: ct,
	}
}

func (n *Node) tokenLength() int {
	return n.To - n.From
}

func (n *Node) nodeLength() int {
	return len(n.Nodes)
}

func (n *Node) appendChar(to int) {
	if n.tokenLength() == 0 {
		n.From = to - 1
	}

	n.To = to
}

func (n *Node) append(p *Node) {
	n.Nodes = append(n.Nodes, p)
	if n.tokenLength() == 0 {
		n.From = p.From
	}

	n.To = p.To
}

func (n *Node) commit(t []rune) {
	n.tokens = t

	var nodes []*Node
	for _, ni := range n.Nodes {
		ni.commit(t)
		if ni.commitType&Alias != 0 {
			nodes = append(nodes, ni.Nodes...)
		} else {
			nodes = append(nodes, ni)
		}
	}

	n.Nodes = nodes
}

func (n *Node) String() string {
	if n.From >= len(n.tokens) || n.To > len(n.tokens) {
		return n.Name + ":incomplete"
	}

	return fmt.Sprintf("%s:%d:%d:%s", n.Name, n.From, n.To, n.Text())
}

func (n *Node) Text() string {
	return string(n.tokens[n.From:n.To])
}
