package treerack

type storeEntry struct {
	match   *idSet
	noMatch *idSet
	matches   []int
	all []int
}

type store struct {
	entries []*storeEntry
}

func (s *store) getEntry(offset int) *storeEntry {
	if len(s.entries) <= offset {
		return nil
	}

	return s.entries[offset]
}

func (s *store) hasNoMatch(offset, id int) bool {
	e := s.getEntry(offset)
	if e == nil {
		return false
	}

	return e.noMatch.has(id)
}

func (s *store) getMatch(offset, id int) (int, bool, bool) {
	e := s.getEntry(offset)
	if e == nil {
		return 0, false, false
	}

	if e.noMatch.has(id) {
		return 0, false, true
	}

	if !e.match.has(id) {
		return 0, false, false
	}

	for i := 0; i < len(e.matches); i += 2 {
		if e.matches[i] == id {
			return e.matches[i + 1], true, true
		}
	}

	return 0, false, false
}

func (s *store) ensureOffset(offset int) {
	if len(s.entries) > offset {
		return
	}

	if cap(s.entries) > offset {
		s.entries = s.entries[:offset+1]
		return
	}

	s.entries = s.entries[:cap(s.entries)]
	for len(s.entries) <= offset {
		s.entries = append(s.entries, nil)
	}
}

func (s *store) ensureEntry(offset int) *storeEntry {
	s.ensureOffset(offset)
	e := s.entries[offset]
	if e != nil {
		return e
	}

	e = &storeEntry{
		match:   &idSet{},
		noMatch: &idSet{},
	}

	s.entries[offset] = e
	return e
}

func (s *store) setMatch(offset, id, to int) {
	e := s.ensureEntry(offset)

	e.match.set(id)
	for i := 0; i < len(e.matches); i += 2 {
		if e.matches[i] == id {
			if to > e.matches[i + 1] {
				e.matches[i + 1] = to
			}

			if to != e.matches[i + 1] {
				e.all = append(e.all, id, to)
			}

			return
		}
	}

	e.matches = append(e.matches, id, to)
}

func (s *store) setNoMatch(offset, id int) {
	e := s.ensureEntry(offset)

	if e.match.has(id) {
		return
	}

	e.noMatch.set(id)
}

func (s *store) add(offset, id, to int) {
	e := s.ensureEntry(offset)
	e.all = append(e.all, id, to)
}
