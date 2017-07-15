package treerack

import "strconv"

type idSet struct {
	ids []uint
}

func divModBits(id int) (int, int) {
	return id / strconv.IntSize, id % strconv.IntSize
}

func (s *idSet) set(id int) {
	d, m := divModBits(id)
	if d >= len(s.ids) {
		if d < cap(s.ids) {
			s.ids = s.ids[:d+1]
		} else {
			s.ids = s.ids[:cap(s.ids)]
			for i := cap(s.ids); i <= d; i++ {
				s.ids = append(s.ids, 0)
			}
		}
	}

	s.ids[d] |= 1 << uint(m)
}

func (s *idSet) unset(id int) {
	d, m := divModBits(id)
	if d >= len(s.ids) {
		return
	}

	s.ids[d] &^= 1 << uint(m)
}

func (s *idSet) has(id int) bool {
	d, m := divModBits(id)
	if d >= len(s.ids) {
		return false
	}

	return s.ids[d]&(1<<uint(m)) != 0
}
