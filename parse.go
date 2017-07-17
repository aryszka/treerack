package treerack

import "fmt"

type definition interface {
	nodeName() string
	nodeID() int
	setID(int)
	parser(*registry, *idSet) (parser, error)
	commitType() CommitType
	// builder() builder
}

type parser interface {
	nodeName() string
	nodeID() int
	setIncludedBy(parser, *idSet)
	storeIncluded(*context, int, int) // can be just an id set, taking what's excluded from the context
	parse(Trace, *context)
}

type builder interface {
	nodeName() string
	nodeID() int
	build(*context) *Node
}

func parserNotFound(name string) error {
	return fmt.Errorf("parser not found: %s", name)
}

func cannotIncludeParsers(name string) error {
	return fmt.Errorf("parser: %s cannot include other parsers", name)
}

func parse(t Trace, p parser, c *context) (*Node, error) {
	p.parse(t, c)
	if c.readErr != nil {
		return nil, c.readErr
	}

	if !c.match {
		return nil, ErrInvalidInput
	}

	if err := c.finalize(); err != nil {
		return nil, err
	}

	return c.node, nil
}
