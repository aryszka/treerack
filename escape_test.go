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

func TestEscape(t *testing.T) {
	const (
		banned    = "\b\f\n\r\t\v"
		unescaped = "\b\f\n\r\t\v"
		expected  = "\\b\\f\\n\\r\\t\\v"
	)

	e := escape('\\', []rune(banned), []rune(unescaped))
	if string(e) != expected {
		t.Error("failed to escape", string(e), expected)
	}
}
