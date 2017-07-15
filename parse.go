package treerack

import "fmt"

type definition interface {
	nodeName() string
	nodeID() int
	setID(int)
	parser(*registry, []string) (parser, error)
	commitType() CommitType
}

type parser interface {
	nodeName() string
	nodeID() int
	setIncludedBy(parser, []string)
	storeIncluded(*context, *Node)
	parse(Trace, *context)
}

func parserNotFound(name string) error {
	return fmt.Errorf("parser not found: %s", name)
}

func cannotIncludeParsers(name string) error {
	return fmt.Errorf("parser: %s cannot include other parsers", name)
}

func stringsContainDeprecated(ss []string, s string) bool {
	for _, si := range ss {
		if si == s {
			return true
		}
	}

	return false
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
