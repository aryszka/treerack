package treerack

import "testing"

func TestUnescape(t *testing.T) {
	t.Run("char should be escaped", func(t *testing.T) {
		if _, err := unescape('\\', []rune{'a'}, []rune{'a'}); err == nil {
			t.Error("failed to fail")
		}
	})

	t.Run("finished with escape char", func(t *testing.T) {
		if _, err := unescape('\\', []rune{'a'}, []rune{'b', '\\'}); err == nil {
			t.Error("failed to fail")
		}
	})

	t.Run("unescapes", func(t *testing.T) {
		u, err := unescape('\\', []rune{'a'}, []rune{'b', '\\', 'a'})
		if err != nil {
			t.Error(err)
			return
		}

		if string(u) != "ba" {
			t.Error("unescape failed")
		}
	})
}
