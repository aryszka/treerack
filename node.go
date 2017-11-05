package treerack

import "fmt"

type Node struct {
	Name     string
	Nodes    []*Node
	From, To int
	tokens   []rune
}

func mapNodes(m func(n *Node) *Node, n []*Node) []*Node {
	var nn []*Node
	for i := range n {
		nn = append(nn, m(n[i]))
	}

	return nn
}

func filterNodes(f func(n *Node) bool, n []*Node) []*Node {
	var nn []*Node
	for i := range n {
		if f(n[i]) {
			nn = append(nn, n[i])
		}
	}

	return nn
}

func (n *Node) String() string {
	return fmt.Sprintf("%s:%d:%d:%s", n.Name, n.From, n.To, n.Text())
}

func (n *Node) Text() string {
	return string(n.tokens[n.From:n.To])
}
