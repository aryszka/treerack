package treerack

import "fmt"

const whitespaceName = ":ws"

func brokenRegistryError(err error) error {
	return fmt.Errorf("broken registry: %v", err)
}

func splitWhitespaceDefs(all map[string]definition) ([]definition, []definition) {
	var whitespaceDefs, nonWhitespaceDefs []definition
	for _, def := range all {
		if def.commitType()&Whitespace != 0 {
			def.setCommitType(def.commitType() | Alias)
			whitespaceDefs = append(whitespaceDefs, def)
			continue
		}

		nonWhitespaceDefs = append(nonWhitespaceDefs, def)
	}

	return whitespaceDefs, nonWhitespaceDefs
}

func splitRoot(defs []definition) (definition, []definition) {
	var (
		root definition
		rest []definition
	)

	for _, def := range defs {
		if def.commitType()&Root != 0 {
			root = def
			continue
		}

		rest = append(rest, def)
	}

	return root, rest
}

func mergeWhitespaceDefs(ws []definition) definition {
	var names []string
	for _, def := range ws {
		names = append(names, def.nodeName())
	}

	return newChoice(whitespaceName, Alias, names)
}

// TODO: validate min and max

func applyWhitespaceToSeq(s *sequenceDefinition) []definition {
	var (
		defs  []definition
		items []SequenceItem
	)

	for i, item := range s.items {
		if i > 0 {
			items = append(items, SequenceItem{Name: whitespaceName, Min: 0, Max: -1})
		}

		if item.Max >= 0 && item.Max <= 1 {
			items = append(items, item)
			continue
		}

		singleItem := SequenceItem{Name: item.Name, Min: 1, Max: 1}

		restName := item.Name + ":wsrest"
		restDef := newSequence(restName, Alias, []SequenceItem{{Name: whitespaceName, Min: 0, Max: -1}, singleItem})
		defs = append(defs, restDef)

		restItems := SequenceItem{Name: restName, Min: 0, Max: -1}
		if item.Min > 0 {
			restItems.Min = item.Min - 1
		}
		if item.Max > 0 {
			restItems.Min = item.Max - 1
		}

		if item.Min > 0 {
			items = append(items, singleItem, restItems)
			continue
		}

		optName := item.Name + ":wsopt"
		optDef := newSequence(optName, Alias, []SequenceItem{singleItem, restItems})
		defs = append(defs, optDef)
		items = append(items, SequenceItem{Name: optName, Min: 0, Max: 1})
	}

	s = newSequence(s.nodeName(), s.commitType(), items)
	defs = append(defs, s)
	return defs
}

func applyWhitespace(defs []definition) []definition {
	var defsWS []definition
	for _, def := range defs {
		if def.commitType()&NoWhitespace != 0 {
			defsWS = append(defsWS, def)
			continue
		}

		seq, ok := def.(*sequenceDefinition)
		if !ok {
			defsWS = append(defsWS, def)
			continue
		}

		defsWS = append(defsWS, applyWhitespaceToSeq(seq)...)
	}

	return defsWS
}

func applyWhitespaceRoot(root definition) (definition, definition) {
	original, name := root, root.nodeName()
	wsName := ":wsroot:" + name

	original.setNodeName(wsName)
	original.setCommitType(original.commitType() &^ Root)
	original.setCommitType(original.commitType() | Alias)

	root = newSequence(name, Root, []SequenceItem{{
		Name: whitespaceName,
		Min:  0,
		Max:  -1,
	}, {
		Name: wsName,
		Min:  1,
		Max:  1,
	}, {
		Name: whitespaceName,
		Min:  0,
		Max:  -1,
	}})

	return original, root
}

func registerPatched(r *registry, defs ...definition) {
	for _, def := range defs {
		if err := r.setDefinition(def); err != nil {
			panic(brokenRegistryError(err))
		}
	}
}

func initWhitespace(r *registry) *registry {
	whitespaceDefs, defs := splitWhitespaceDefs(r.definitions)
	if len(whitespaceDefs) == 0 {
		return r
	}

	whitespace := mergeWhitespaceDefs(whitespaceDefs)
	defs = applyWhitespace(defs)

	root, defs := splitRoot(defs)
	originalRoot, root := applyWhitespaceRoot(root)

	r = newRegistry()
	registerPatched(r, whitespace)
	registerPatched(r, whitespaceDefs...)
	registerPatched(r, defs...)
	registerPatched(r, originalRoot, root)
	return r
}
