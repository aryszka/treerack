package treerack

func runesContain(rs []rune, r rune) bool {
	for _, ri := range rs {
		if ri == r {
			return true
		}
	}

	return false
}

func escapeChar(escape, c rune) []rune {
	switch c {
	case '\b':
		return []rune{escape, 'b'}
	case '\f':
		return []rune{escape, 'f'}
	case '\n':
		return []rune{escape, 'n'}
	case '\r':
		return []rune{escape, 'r'}
	case '\t':
		return []rune{escape, 't'}
	case '\v':
		return []rune{escape, 'v'}
	default:
		return []rune{escape, c}
	}
}

func escape(escape rune, banned, chars []rune) []rune {
	var escaped []rune
	for i := range chars {
		if runesContain(banned, chars[i]) {
			escaped = append(escaped, escapeChar(escape, chars[i])...)
			continue
		}

		escaped = append(escaped, chars[i])
	}

	return escaped
}

func unescapeChar(c rune) rune {
	switch c {
	case 'b':
		return '\b'
	case 'f':
		return '\f'
	case 'n':
		return '\n'
	case 'r':
		return '\r'
	case 't':
		return '\t'
	case 'v':
		return '\v'
	default:
		return c
	}
}

func unescape(escape rune, banned, chars []rune) ([]rune, error) {
	var (
		unescaped []rune
		escaped   bool
	)

	for _, ci := range chars {
		if escaped {
			unescaped = append(unescaped, unescapeChar(ci))
			escaped = false
			continue
		}

		switch {
		case ci == escape:
			escaped = true
		case runesContain(banned, ci):
			return nil, ErrInvalidEscapeCharacter
		default:
			unescaped = append(unescaped, ci)
		}
	}

	if escaped {
		return nil, ErrInvalidEscapeCharacter
	}

	return unescaped, nil
}

func unescapeCharSequence(s string) ([]rune, error) {
	return unescape('\\', []rune{'"', '\\'}, []rune(s))
}
