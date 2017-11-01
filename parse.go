package treerack

import "fmt"

type definition interface {
	nodeName() string
	setNodeName(string)
	nodeID() int
	commitType() CommitType
	setCommitType(CommitType)
	setID(int)
	validate(*registry) error
	init(*registry)
	setIncludedBy(*registry, int)
	parser(*registry) parser
	builder() builder
}

type parser interface {
	nodeName() string
	nodeID() int
	parse(*context)
}

type builder interface {
	nodeName() string
	nodeID() int
	build(*context) ([]*Node, bool)
}

func parserNotFound(name string) error {
	return fmt.Errorf("parser not found: %s", name)
}

func cannotIncludeParsers(name string) error {
	return fmt.Errorf("parser: %s cannot include other parsers", name)
}

func intsContain(is []int, i int) bool {
	for _, ii := range is {
		if ii == i {
			return true
		}
	}

	return false
}

func parse(p parser, c *context) error {
	p.parse(c)
	if c.readErr != nil {
		return c.readErr
	}

	if !c.match {
		return ErrInvalidInput
	}

	if err := c.finalize(p); err != nil {
		return err
	}

	return nil
}

func build(b builder, c *context) *Node {
	c.offset = 0
	n, ok := b.build(c)
	if !ok || len(n) != 1 {
		panic("damaged parse result")
	}

	return n[0]
}
