package treerack

import "fmt"

type Node struct {
	Name     string
	Nodes    []*Node
	From, To int
	tokens   []rune
}

func (n *Node) Tokens() []rune {
	defer func() {
		if err := recover(); err != nil {
			println(len(n.tokens), n.From, n.To)
			panic(err)
		}
	}()
	return n.tokens
}

func (n *Node) String() string {
	return fmt.Sprintf("%s:%d:%d:%s", n.Name, n.From, n.To, n.Text())
}

func (n *Node) Text() string {
	return string(n.Tokens()[n.From:n.To])
}
