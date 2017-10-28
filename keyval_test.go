package treerack

import "testing"

func TestKeyVal(t *testing.T) {
	runTestsFile(t, "keyval.parser", []testItem{{
		title: "empty",
	}, {
		title: "a comment",
		text:  "# a comment",
	}, {
		title: "a key",
		text:  "a key",
		nodes: []*Node{{
			Name: "key-val",
			To:   5,
			Nodes: []*Node{{
				Name: "key",
				To:   5,
				Nodes: []*Node{{
					Name: "symbol",
					To:   5,
				}},
			}},
		}},
	}, {
		title: "a key with a preceeding whitespace",
		text:  " a key",
		nodes: []*Node{{
			Name: "key-val",
			From: 1,
			To:   6,
			Nodes: []*Node{{
				Name: "key",
				From: 1,
				To:   6,
				Nodes: []*Node{{
					Name: "symbol",
					From: 1,
					To:   6,
				}},
			}},
		}},
	}, {
		title: "a key and a comment",
		text: `
			# a comment

			a key
		`,
		nodes: []*Node{{
			Name: "key-val",
			From: 20,
			To:   25,
			Nodes: []*Node{{
				Name: "key",
				From: 20,
				To:   25,
				Nodes: []*Node{{
					Name: "symbol",
					From: 20,
					To:   25,
				}},
			}},
		}},
	}, {
		title: "a key value pair",
		text:  "a key = a value",
		nodes: []*Node{{
			Name: "key-val",
			To:   15,
			Nodes: []*Node{{
				Name: "key",
				To:   5,
				Nodes: []*Node{{
					Name: "symbol",
					To:   5,
				}},
			}, {
				Name: "value",
				From: 8,
				To:   15,
			}},
		}},
	}, {
		title: "key value pairs with a comment at the end of line",
		text: "a key       = a value       # a comment\n" +
			"another key = another value # another comment",
		nodes: []*Node{{
			Name: "key-val",
			From: 0,
			To:   39,
			Nodes: []*Node{{
				Name: "key",
				From: 0,
				To:   5,
				Nodes: []*Node{{
					Name: "symbol",
					From: 0,
					To:   5,
				}},
			}, {
				Name: "value",
				From: 14,
				To:   21,
			}},
		}, {
			Name: "key-val",
			From: 40,
			To:   85,
			Nodes: []*Node{{
				Name: "key",
				From: 40,
				To:   51,
				Nodes: []*Node{{
					Name: "symbol",
					From: 40,
					To:   51,
				}},
			}, {
				Name: "value",
				From: 54,
				To:   67,
			}},
		}},
	}, {
		title: "value without a key",
		text:  "= a value",
		nodes: []*Node{{
			Name: "key-val",
			To:   9,
			Nodes: []*Node{{
				Name: "value",
				From: 2,
				To:   9,
			}},
		}},
	}, {
		title: "a key value pair with comment",
		text: `
			# a comment
			a key = a value
		`,
		nodes: []*Node{{
			Name: "key-val",
			From: 4,
			To:   34,
			Nodes: []*Node{{
				Name: "comment",
				From: 4,
				To:   15,
			}, {
				Name: "key",
				From: 19,
				To:   24,
				Nodes: []*Node{{
					Name: "symbol",
					From: 19,
					To:   24,
				}},
			}, {
				Name: "value",
				From: 27,
				To:   34,
			}},
		}},
	}, {
		title: "a key with multiple symbols",
		text:  "a key . with.multiple.symbols=a value",
		nodes: []*Node{{
			Name: "key-val",
			To:   37,
			Nodes: []*Node{{
				Name: "key",
				From: 0,
				To:   29,
				Nodes: []*Node{{
					Name: "symbol",
					From: 0,
					To:   5,
				}, {
					Name: "symbol",
					From: 8,
					To:   12,
				}, {
					Name: "symbol",
					From: 13,
					To:   21,
				}, {
					Name: "symbol",
					From: 22,
					To:   29,
				}},
			}, {
				Name: "value",
				From: 30,
				To:   37,
			}},
		}},
	}, {
		title: "a group key",
		text: `
			# a comment
			[a group key.empty]
		`,
		nodes: []*Node{{
			Name: "group-key",
			From: 4,
			To:   38,
			Nodes: []*Node{{
				Name: "comment",
				From: 4,
				To:   15,
			}, {
				Name: "symbol",
				From: 20,
				To:   31,
			}, {
				Name: "symbol",
				From: 32,
				To:   37,
			}},
		}},
	}, {
		title: "a group key with multiple values",
		text: `
			[foo.bar.baz]
			= one
			= two
			= three
		`,
		nodes: []*Node{{
			Name: "group-key",
			Nodes: []*Node{{
				Name: "symbol",
			}, {
				Name: "symbol",
			}, {
				Name: "symbol",
			}},
		}, {
			Name: "key-val",
			Nodes: []*Node{{
				Name: "value",
			}},
		}, {
			Name: "key-val",
			Nodes: []*Node{{
				Name: "value",
			}},
		}, {
			Name: "key-val",
			Nodes: []*Node{{
				Name: "value",
			}},
		}},
		ignorePosition: true,
	}, {
		title: "a group key with multiple values, in a single line",
		text:  "[foo.bar.baz] = one = two = three",
		nodes: []*Node{{
			Name: "group-key",
			Nodes: []*Node{{
				Name: "symbol",
			}, {
				Name: "symbol",
			}, {
				Name: "symbol",
			}},
		}, {
			Name: "key-val",
			Nodes: []*Node{{
				Name: "value",
			}},
		}, {
			Name: "key-val",
			Nodes: []*Node{{
				Name: "value",
			}},
		}, {
			Name: "key-val",
			Nodes: []*Node{{
				Name: "value",
			}},
		}},
		ignorePosition: true,
	}, {
		title: "full example",
		text: `
			# a keyval document

			key1 = foo
			key1.a = bar
			key1.b = baz

			key2 = qux

			# foo bar baz values
			[foo.bar.baz]
			a = 1
			b = 2 # even
			c = 3
		`,
		nodes: []*Node{{
			Name: "key-val",
			Nodes: []*Node{{
				Name: "key",
				Nodes: []*Node{{
					Name: "symbol",
				}},
			}, {
				Name: "value",
			}},
		}, {
			Name: "key-val",
			Nodes: []*Node{{
				Name: "key",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "symbol",
				}},
			}, {
				Name: "value",
			}},
		}, {
			Name: "key-val",
			Nodes: []*Node{{
				Name: "key",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "symbol",
				}},
			}, {
				Name: "value",
			}},
		}, {
			Name: "key-val",
			Nodes: []*Node{{
				Name: "key",
				Nodes: []*Node{{
					Name: "symbol",
				}},
			}, {
				Name: "value",
			}},
		}, {
			Name: "group-key",
			Nodes: []*Node{{
				Name: "comment",
			}, {
				Name: "symbol",
			}, {
				Name: "symbol",
			}, {
				Name: "symbol",
			}},
		}, {
			Name: "key-val",
			Nodes: []*Node{{
				Name: "key",
				Nodes: []*Node{{
					Name: "symbol",
				}},
			}, {
				Name: "value",
			}},
		}, {
			Name: "key-val",
			Nodes: []*Node{{
				Name: "key",
				Nodes: []*Node{{
					Name: "symbol",
				}},
			}, {
				Name: "value",
			}},
		}, {
			Name: "key-val",
			Nodes: []*Node{{
				Name: "key",
				Nodes: []*Node{{
					Name: "symbol",
				}},
			}, {
				Name: "value",
			}},
		}},
		ignorePosition: true,
	}})
}
