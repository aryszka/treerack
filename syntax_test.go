package treerack

import (
	"bytes"
	"testing"
)

func TestDefinitionProperties(t *testing.T) {
	testProperties := func(t *testing.T, d definition, withCommit bool) {
		d.setName("foo")
		if d.nodeName() != "foo" {
			t.Error("name failed")
			return
		}

		d.setID(42)
		if d.nodeID() != 42 {
			t.Error("id failed")
			return
		}

		if !withCommit {
			return
		}

		d.setCommitType(Alias | NoWhitespace)
		if d.commitType() != Alias|NoWhitespace {
			t.Error("commit type failed")
			return
		}

		d.init(newRegistry())

		if p := d.parser(); p.nodeName() != "foo" || p.nodeID() != 42 {
			t.Error("parser failed")
		}

		if b := d.builder(); b.nodeName() != "foo" || b.nodeID() != 42 {
			t.Error("parser failed")
		}
	}

	t.Run("char", func(t *testing.T) {
		testProperties(t, newChar("", false, nil, nil), false)
	})

	t.Run("choice", func(t *testing.T) {
		testProperties(t, newChoice("", None, nil), true)
	})

	t.Run("sequence", func(t *testing.T) {
		testProperties(t, newSequence("", None, nil), true)
	})
}

func TestValidation(t *testing.T) {
	t.Run("undefined parser", func(t *testing.T) {
		t.Run("sequence", func(t *testing.T) {
			if _, err := openSyntaxString("a = b"); err == nil {
				t.Error("failed to fail")
			}
		})

		t.Run("sequence in sequence", func(t *testing.T) {
			if _, err := openSyntaxString("a:root = b; b = c"); err == nil {
				t.Error("failed to fail")
			}
		})

		t.Run("choice", func(t *testing.T) {
			if _, err := openSyntaxString("a = a | b"); err == nil {
				t.Error("failed to fail")
			}
		})
	})

	t.Run("choice item", func(t *testing.T) {
		if _, err := openSyntaxString("b = c; a = a | b"); err == nil {
			t.Error("failed to fail")
		}
	})
}

func TestInit(t *testing.T) {
	t.Run("add generalizations", func(t *testing.T) {
		t.Run("choice containing itself", func(t *testing.T) {
			s, err := openSyntaxString(`c = "c"; d = "d"; b = a | c; a = b | d`)
			if err != nil {
				t.Error(err)
				return
			}

			s.Init()
			if len(s.root.(*choiceDefinition).generalizations) != 2 {
				t.Error("invalid number of generalizations")
			}
		})

		t.Run("choice containing a sequence two times", func(t *testing.T) {
			s, err := openSyntaxString(`a = "a"; b = a | a`)
			if err != nil {
				t.Error(err)
				return
			}

			s.Init()
			if len(s.registry.definitions["a"].(*sequenceDefinition).generalizations) != 1 {
				t.Error("invalid number of generalizations")
			}
		})
	})

	t.Run("reinit after failed", func(t *testing.T) {
		s := &Syntax{}
		if err := s.Choice("a", None, "b"); err != nil {
			t.Error(err)
			return
		}

		if err := s.Init(); err == nil {
			t.Error("failed to fail")
			return
		}

		if err := s.Init(); err == nil {
			t.Error("failed to fail")
			return
		}
	})

	t.Run("init without definitions", func(t *testing.T) {
		s := &Syntax{}
		if err := s.Init(); err == nil {
			t.Error("failed to fail")
		}
	})

	t.Run("root is an alias", func(t *testing.T) {
		s := &Syntax{}
		if err := s.AnyChar("a", Root|Alias); err != nil {
			t.Error(err)
			return
		}

		if err := s.Init(); err == nil {
			t.Error("failed to fail")
		}
	})

	t.Run("root is whitespace", func(t *testing.T) {
		s := &Syntax{}
		if err := s.AnyChar("a", Root|Whitespace); err != nil {
			t.Error(err)
			return
		}

		if err := s.Init(); err == nil {
			t.Error("failed to fail")
		}
	})

	t.Run("init fails during call to parse", func(t *testing.T) {
		s := &Syntax{}
		if _, err := s.Parse(bytes.NewBuffer(nil)); err == nil {
			t.Error("failed to fail")
		}
	})
}

func TestTooBigNumber(t *testing.T) {
	t.Run("range to", func(t *testing.T) {
		if _, err := openSyntaxString(`A = "a"{0,123456789012345678901234567890}`); err == nil {
			t.Error("failed to fail")
		}
	})

	t.Run("range from", func(t *testing.T) {
		if _, err := openSyntaxString(`A = "a"{123456789012345678901234567890,0}`); err == nil {
			t.Error("failed to fail")
		}
	})

	t.Run("fixed count", func(t *testing.T) {
		if _, err := openSyntaxString(`A = "a"{123456789012345678901234567890}`); err == nil {
			t.Error("failed to fail")
		}
	})

	t.Run("error in sequence item", func(t *testing.T) {
		if _, err := openSyntaxString(`A = ("a"{123456789012345678901234567890})*`); err == nil {
			t.Error("failed to fail")
		}
	})

	t.Run("error in choice option", func(t *testing.T) {
		if _, err := openSyntaxString(`A = "42" | "a"{123456789012345678901234567890}`); err == nil {
			t.Error("failed to fail")
		}
	})
}

func TestDefinition(t *testing.T) {
	t.Run("duplicate definition", func(t *testing.T) {
		s := &Syntax{}

		if err := s.AnyChar("a", None); err != nil {
			t.Error(err)
			return
		}

		if err := s.AnyChar("a", None); err == nil {
			t.Error("failed to fail")
		}
	})

	t.Run("invalid symbol", func(t *testing.T) {
		s := &Syntax{}

		t.Run("any char", func(t *testing.T) {
			if err := s.AnyChar("foo[]", None); err == nil {
				t.Error("failed to fail")
				return
			}
		})

		t.Run("class", func(t *testing.T) {
			if err := s.Class("foo[]", None, false, []rune("a"), nil); err == nil {
				t.Error("failed to fail")
				return
			}
		})

		t.Run("char sequence", func(t *testing.T) {
			if err := s.CharSequence("foo[]", None, []rune("a")); err == nil {
				t.Error("failed to fail")
				return
			}
		})

		t.Run("sequence", func(t *testing.T) {
			if err := s.Sequence("foo[]", None, SequenceItem{Name: "bar"}); err == nil {
				t.Error("failed to fail")
				return
			}
		})

		t.Run("choice", func(t *testing.T) {
			if err := s.Choice("foo[]", None, "bar"); err == nil {
				t.Error("failed to fail")
				return
			}
		})
	})

	t.Run("multiple roots", func(t *testing.T) {
		s := &Syntax{}

		if err := s.AnyChar("foo", Root); err != nil {
			t.Error(err)
			return
		}

		if err := s.AnyChar("bar", Root); err == nil {
			t.Error("failed to fail")
		}
	})

	t.Run("define after init", func(t *testing.T) {
		s := &Syntax{}

		if err := s.AnyChar("foo", None); err != nil {
			t.Error(err)
			return
		}

		if err := s.Init(); err != nil {
			t.Error(err)
			return
		}

		if err := s.CharSequence("bar", None, []rune("bar")); err == nil {
			t.Error("failed to fail")
		}
	})

	t.Run("define", func(t *testing.T) {
		s := &Syntax{}

		t.Run("any char", func(t *testing.T) {
			if err := s.AnyChar("a", None); err != nil {
				t.Error(err)
			}

			if _, ok := s.registry.definition("a"); !ok {
				t.Error("definition failed")
			}
		})

		t.Run("class", func(t *testing.T) {
			if err := s.Class("b", None, false, []rune("b"), nil); err != nil {
				t.Error(err)
			}

			if _, ok := s.registry.definition("b"); !ok {
				t.Error("definition failed")
			}
		})

		t.Run("char sequence", func(t *testing.T) {
			if err := s.CharSequence("c", None, []rune("b")); err != nil {
				t.Error(err)
			}

			if _, ok := s.registry.definition("c"); !ok {
				t.Error("definition failed")
			}
		})

		t.Run("sequence", func(t *testing.T) {
			if err := s.Sequence("d", None, SequenceItem{Name: "d"}); err != nil {
				t.Error(err)
			}

			if _, ok := s.registry.definition("d"); !ok {
				t.Error("definition failed")
			}
		})

		t.Run("choice", func(t *testing.T) {
			if err := s.Choice("e", None, "e"); err != nil {
				t.Error(err)
			}

			if _, ok := s.registry.definition("e"); !ok {
				t.Error("definition failed")
			}
		})
	})
}

func TestReadSyntax(t *testing.T) {
	t.Skip()

	t.Run("already initialized", func(t *testing.T) {
		s := &Syntax{}
		if err := s.AnyChar("a", None); err != nil {
			t.Error(err)
			return
		}

		if err := s.Init(); err != nil {
			t.Error(err)
			return
		}

		if err := s.ReadSyntax(bytes.NewBuffer(nil)); err == nil {
			t.Error(err)
		}
	})

	t.Run("not implemented", func(t *testing.T) {
		s := &Syntax{}
		if err := s.ReadSyntax(bytes.NewBuffer(nil)); err == nil {
			t.Error(err)
		}
	})
}

func TestGenerateSyntax(t *testing.T) {
	t.Skip()

	t.Run("init fails", func(t *testing.T) {
		s := &Syntax{}
		if err := s.Choice("a", None, "b"); err != nil {
			t.Error(err)
			return
		}

		if err := s.Generate(bytes.NewBuffer(nil)); err == nil {
			t.Error(err)
		}
	})

	t.Run("not implemented", func(t *testing.T) {
		s := &Syntax{}
		if err := s.AnyChar("a", None); err != nil {
			t.Error(err)
			return
		}

		if err := s.Generate(bytes.NewBuffer(nil)); err == nil {
			t.Error(err)
		}
	})
}
