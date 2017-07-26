package treerack

import "fmt"

type definition interface {
	nodeName() string
	nodeID() int
	setID(int)
	init(*registry) error
	setIncludedBy(*registry, int, *idSet) error
	parser(*registry, *idSet) (parser, error)
	commitType() CommitType
	// builder() builder
}

type parser interface {
	nodeName() string
	nodeID() int
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

func intsContain(is []int, i int) bool {
	for _, ii := range is {
		if ii == i {
			return true
		}
	}

	return false
}

func appendIfMissing(is []int, i int) []int {
	if intsContain(is, i) {
		return is
	}

	return append(is, i)
}

func setItemsIncludedBy(r *registry, items []string, includedBy int, parsers *idSet) error {
	for _, item := range items {
		di, ok := r.definition(item)
		if !ok {
			return ErrNoParsersDefined
		}

		di.setIncludedBy(r, includedBy, parsers)
	}

	return nil
}

func sequenceItemNames(items []SequenceItem) []string {
	names := make([]string, len(items))
	for i := range items {
		names[i] = items[i].Name
	}

	return names
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
