package treerack

import "github.com/aryszka/treerack/self"

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

func mapSelfNode(n *self.Node) *Node {
	nn := Node{
		Name:   n.Name,
		From:   n.From,
		To:     n.To,
		tokens: n.Tokens(),
	}

	for i := range n.Nodes {
		nn.Nodes = append(nn.Nodes, mapSelfNode(n.Nodes[i]))
	}

	return &nn
}
