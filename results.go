package treerack

type results struct {
	noMatch []*idSet
	match   [][]int
}

func (s *results) getMatch(offset, id int) (int, bool, bool) {
	if len(s.noMatch) > offset && s.noMatch[offset] != nil && s.noMatch[offset].has(id) {
		return 0, false, true
	}

	if len(s.match) <= offset {
		return 0, false, false
	}

	var (
		found bool
		to    int
	)

	for i := 0; i < len(s.match[offset]); i += 2 {
		if s.match[offset][i] != id {
			continue
		}

		found = true
		if s.match[offset][i+1] > to {
			to = s.match[offset][i+1]
		}
	}

	return to, found, found
}

func (s *results) hasMatchTo(offset, id, to int) bool {
	if len(s.noMatch) > offset && s.noMatch[offset] != nil && s.noMatch[offset].has(id) {
		return false
	}

	if len(s.match) <= offset {
		return false
	}

	for i := 0; i < len(s.match[offset]); i += 2 {
		if s.match[offset][i] != id {
			continue
		}

		if s.match[offset][i+1] == to {
			return true
		}
	}

	return false
}

func (s *results) takeMatch(offset, id int, includedBy *idSet) (int, bool) {
	if len(s.match) <= offset {
		return 0, false
	}

	var (
		found bool
		to    int
		index int
	)

	for i := 0; i < len(s.match[offset]); i += 2 {
		if s.match[offset][i] != id {
			continue
		}

		if s.match[offset][i+1] > to || !found {
			to = s.match[offset][i+1]
			index = i
		}

		found = true
	}

	if found && to-offset > 0 {
		s.match[offset][index] = -1
		for i := 0; i < len(s.match[offset]); i += 2 {
			if includedBy.has(s.match[offset][i]) && s.match[offset][i+1] == to {
				s.match[offset][i] = -1
			}
		}
	}

	return to, found
}

func (s *results) takeMatchLength(offset, id, to int) {
	if len(s.match) <= offset {
		return
	}

	for i := 0; i < len(s.match[offset]); i += 2 {
		if s.match[offset][i] == id && s.match[offset][i+1] == to {
			s.match[offset][i] = -1
			return
		}
	}
}

func (s *results) ensureOffset(offset int) {
	if len(s.match) > offset {
		return
	}

	if cap(s.match) > offset {
		s.match = s.match[:offset+1]
		return
	}

	s.match = s.match[:cap(s.match)]
	for i := len(s.match); i <= offset; i++ {
		s.match = append(s.match, nil)
	}
}

func (s *results) setMatch(offset, id, to int) {
	s.ensureOffset(offset)
	for i := 0; i < len(s.match[offset]); i += 2 {
		if s.match[offset][i] != id || s.match[offset][i+1] != to {
			continue
		}

		return
	}

	s.match[offset] = append(s.match[offset], id, to)
}

func (s *results) setNoMatch(offset, id int) {
	if len(s.match) > offset {
		for i := 0; i < len(s.match[offset]); i += 2 {
			if s.match[offset][i] != id {
				continue
			}

			return
		}
	}

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
