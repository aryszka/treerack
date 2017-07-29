package treerack

// TODO:
// - store it similarly to the excluded ones? sorted by offset?
// - use a helper field for the last accessed position to walk from there? for every offset?
// - use a helper field to store the largest value and its index, too? for an offset?

type store struct {
	noMatch []*idSet
	match   [][]int
}

func (s *store) hasNoMatch(offset, id int) bool {
	if len(s.noMatch) <= offset || s.noMatch[offset] == nil {
		return false
	}

	return s.noMatch[offset].has(id)
}

func (s *store) getMatch(offset, id int) (int, bool, bool) {
	if s.hasNoMatch(offset, id) {
		return 0, false, true
	}

	if len(s.match) <= offset {
		return 0, false, false
	}

	var (
		found  bool
		length int
	)

	for i := 0; i < len(s.match[offset]); i++ {
		if s.match[offset][i] != id {
			continue
		}

		found = true
		if s.match[offset][i+1] > length {
			length = s.match[offset][i+1]
		}
	}

	return length, found, found
}

func (s *store) ensureOffset(offset int) {
	if len(s.match) > offset {
		return
	}

	if cap(s.match) > offset {
		s.match = s.match[:offset+1]
		return
	}

	s.match = s.match[:cap(s.match)]
	for i := cap(s.match); i <= offset; i++ {
		s.match = append(s.match, nil)
	}
}

func (s *store) setMatch(offset, id, to int) {
	s.ensureOffset(offset)
	s.match[offset] = append(s.match[offset], id, to)
}

func (s *store) setNoMatch(offset, id int) {
	if len(s.noMatch) <= offset {
		if cap(s.noMatch) > offset {
			s.noMatch = s.noMatch[:offset+1]
		} else {
			s.noMatch = s.noMatch[:cap(s.noMatch)]
			for i := cap(s.noMatch); i <= offset; i++ {
				s.noMatch = append(s.noMatch, nil)
			}
		}
	}

	if s.noMatch[offset] == nil {
		s.noMatch[offset] = &idSet{}
	}

	s.noMatch[offset].set(id)
}
