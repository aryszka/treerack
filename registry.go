package treerack

type registry struct {
	idSeed      int
	definition  map[string]definition
	definitions []definition
}

func newRegistry(defs ...definition) *registry {
	r := &registry{
		definition: make(map[string]definition),
	}

	for _, def := range defs {
		r.setDefinition(def)
	}

	return r
}

func (r *registry) setDefinition(d definition) error {
	if _, ok := r.definition[d.nodeName()]; ok {
		return duplicateDefinition(d.nodeName())
	}

	r.idSeed++
	id := r.idSeed
	d.setID(id)

	r.definition[d.nodeName()] = d
	r.definitions = append(r.definitions, d)
	return nil
}
