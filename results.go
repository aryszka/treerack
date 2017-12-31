package treerack

type results struct {
	noMatch   []*idSet
	match     [][]int
	isPending [][]int
}

func ensureOffsetInts(ints [][]int, offset int) [][]int {
	if len(ints) > offset {
		return ints
	}

	if cap(ints) > offset {
		ints = ints[:offset+1]
		return ints
	}

	ints = ints[:cap(ints)]
	for i := len(ints); i <= offset; i++ {
		ints = append(ints, nil)
	}

	return ints
}

func ensureOffsetIDs(ids []*idSet, offset int) []*idSet {
	if len(ids) > offset {
		return ids
	}

	if cap(ids) > offset {
		ids = ids[:offset+1]
		return ids
	}

	ids = ids[:cap(ids)]
	for i := len(ids); i <= offset; i++ {
		ids = append(ids, nil)
	}

	return ids
}

func (r *results) setMatch(offset, id, to int) {
	r.match = ensureOffsetInts(r.match, offset)

	for i := 0; i < len(r.match[offset]); i += 2 {
		if r.match[offset][i] != id || r.match[offset][i+1] != to {
			continue
		}

		return
	}

	r.match[offset] = append(r.match[offset], id, to)
}

func (r *results) setNoMatch(offset, id int) {
	if len(r.match) > offset {
		for i := 0; i < len(r.match[offset]); i += 2 {
			if r.match[offset][i] != id {
				continue
			}

			return
		}
	}

	r.noMatch = ensureOffsetIDs(r.noMatch, offset)
	if r.noMatch[offset] == nil {
		r.noMatch[offset] = &idSet{}
	}

	r.noMatch[offset].set(id)
}

func (r *results) hasMatchTo(offset, id, to int) bool {
	if len(r.match) <= offset {
		return false
	}

	for i := 0; i < len(r.match[offset]); i += 2 {
		if r.match[offset][i] != id {
			continue
		}

		if r.match[offset][i+1] == to {
			return true
		}
	}

	return false
}

func (r *results) longestMatch(offset, id int) (int, bool) {
	if len(r.match) <= offset {
		return 0, false
	}

	var found bool
	to := -1
	for i := 0; i < len(r.match[offset]); i += 2 {
		if r.match[offset][i] != id {
			continue
		}

		if r.match[offset][i+1] > to {
			to = r.match[offset][i+1]
		}

		found = true
	}

	return to, found
}

func (r *results) longestResult(offset, id int) (int, bool, bool) {
	if len(r.noMatch) > offset && r.noMatch[offset] != nil && r.noMatch[offset].has(id) {
		return 0, false, true
	}

	to, ok := r.longestMatch(offset, id)
	return to, ok, ok
}

func (r *results) dropMatchTo(offset, id, to int) {
	for i := 0; i < len(r.match[offset]); i += 2 {
		if r.match[offset][i] != id {
			continue
		}

		if r.match[offset][i+1] == to {
			r.match[offset][i] = -1
			return
		}
	}
}

func (r *results) resetPending() {
	r.isPending = nil
}

func (r *results) pending(offset, id int) bool {
	if len(r.isPending) <= id {
		return false
	}

	for i := range r.isPending[id] {
		if r.isPending[id][i] == offset {
			return true
		}
	}

	return false
}

func (r *results) markPending(offset, id int) {
	r.isPending = ensureOffsetInts(r.isPending, id)
	for i := range r.isPending[id] {
		if r.isPending[id][i] == -1 {
			r.isPending[id][i] = offset
			return
		}
	}

	r.isPending[id] = append(r.isPending[id], offset)
}

func (r *results) unmarkPending(offset, id int) {
	for i := range r.isPending[id] {
		if r.isPending[id][i] == offset {
			r.isPending[id][i] = -1
			break
		}
	}
}
