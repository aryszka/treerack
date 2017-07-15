package parse

import "testing"

func TestMML(t *testing.T) {
	test(t, "mml.parser", "mml", []testItem{{
		msg:  "empty",
		node: &Node{Name: "mml"},
	}, {
		msg:  "single line comment",
		text: "// foo bar baz",
		nodes: []*Node{{
			Name: "comment",
			To:   14,
			Nodes: []*Node{{
				Name: "line-comment-content",
				From: 2,
				To:   14,
			}},
		}},
	}, {
		msg:  "multiple line comments",
		text: "// foo bar\n// baz qux",
		nodes: []*Node{{
			Name: "comment",
			To:   21,
			Nodes: []*Node{{
				Name: "line-comment-content",
				From: 2,
				To:   10,
			}, {
				Name: "line-comment-content",
				From: 13,
				To:   21,
			}},
		}},
	}, {
		msg:  "block comment",
		text: "/* foo bar baz */",
		nodes: []*Node{{
			Name: "comment",
			To:   17,
			Nodes: []*Node{{
				Name: "block-comment-content",
				From: 2,
				To:   15,
			}},
		}},
	}, {
		msg:  "block comments",
		text: "/* foo bar */\n/* baz qux */",
		nodes: []*Node{{
			Name: "comment",
			To:   27,
			Nodes: []*Node{{
				Name: "block-comment-content",
				From: 2,
				To:   11,
			}, {
				Name: "block-comment-content",
				From: 16,
				To:   25,
			}},
		}},
	}, {
		msg:  "mixed comments",
		text: "// foo\n/* bar */\n// baz",
		nodes: []*Node{{
			Name: "comment",
			To:   23,
			Nodes: []*Node{{
				Name: "line-comment-content",
				From: 2,
				To:   6,
			}, {
				Name: "block-comment-content",
				From: 9,
				To:   14,
			}, {
				Name: "line-comment-content",
				From: 19,
				To:   23,
			}},
		}},
	}, {
		msg:  "int",
		text: "42",
		nodes: []*Node{{
			Name: "int",
			To:   2,
		}},
	}, {
		msg:  "ints",
		text: "1; 2; 3",
		nodes: []*Node{{
			Name: "int",
			To:   1,
		}, {
			Name: "int",
			From: 3,
			To:   4,
		}, {
			Name: "int",
			From: 6,
			To:   7,
		}},
	}, {
		msg:  "int, octal",
		text: "052",
		nodes: []*Node{{
			Name: "int",
			To:   3,
		}},
	}, {
		msg:  "int, hexa",
		text: "0x2a",
		nodes: []*Node{{
			Name: "int",
			To:   4,
		}},
	}, {
		msg:  "float, 0.",
		text: "0.",
		nodes: []*Node{{
			Name: "float",
			To:   2,
		}},
	}, {
		msg:  "float, 72.40",
		text: "72.40",
		nodes: []*Node{{
			Name: "float",
			To:   5,
		}},
	}, {
		msg:  "float, 072.40",
		text: "072.40",
		nodes: []*Node{{
			Name: "float",
			To:   6,
		}},
	}, {
		msg:  "float, 2.71828",
		text: "2.71828",
		nodes: []*Node{{
			Name: "float",
			To:   7,
		}},
	}, {
		msg:  "float, 6.67428e-11",
		text: "6.67428e-11",
		nodes: []*Node{{
			Name: "float",
			To:   11,
		}},
	}, {
		msg:  "float, 1E6",
		text: "1E6",
		nodes: []*Node{{
			Name: "float",
			To:   3,
		}},
	}, {
		msg:  "float, .25",
		text: ".25",
		nodes: []*Node{{
			Name: "float",
			To:   3,
		}},
	}, {
		msg:  "float, .12345E+5",
		text: ".12345E+5",
		nodes: []*Node{{
			Name: "float",
			To:   9,
		}},
	}, {
		msg:  "string, empty",
		text: "\"\"",
		nodes: []*Node{{
			Name: "string",
			To:   2,
		}},
	}, {
		msg:  "string",
		text: "\"foo\"",
		nodes: []*Node{{
			Name: "string",
			To:   5,
		}},
	}, {
		msg:  "string, with new line",
		text: "\"foo\nbar\"",
		nodes: []*Node{{
			Name: "string",
			To:   9,
		}},
	}, {
		msg:  "string, with escaped new line",
		text: "\"foo\\nbar\"",
		nodes: []*Node{{
			Name: "string",
			To:   10,
		}},
	}, {
		msg:  "string, with quotes",
		text: "\"foo \\\"bar\\\" baz\"",
		nodes: []*Node{{
			Name: "string",
			To:   17,
		}},
	}, {
		msg:  "bool, true",
		text: "true",
		nodes: []*Node{{
			Name: "true",
			To:   4,
		}},
	}, {
		msg:  "bool, false",
		text: "false",
		nodes: []*Node{{
			Name: "false",
			To:   5,
		}},
	}, {
		msg:  "symbol",
		text: "foo",
		nodes: []*Node{{
			Name: "symbol",
			To:   3,
		}},
	}, {
		msg:  "dynamic-symbol",
		text: "symbol(a)",
		nodes: []*Node{{
			Name: "dynamic-symbol",
			To:   9,
			Nodes: []*Node{{
				Name: "symbol",
				From: 7,
				To:   8,
			}},
		}},
	}, {
		msg:  "empty list",
		text: "[]",
		nodes: []*Node{{
			Name: "list",
			To:   2,
		}},
	}, {
		msg:  "list",
		text: "[a, b, c]",
		nodes: []*Node{{
			Name: "list",
			To:   9,
			Nodes: []*Node{{
				Name: "symbol",
				From: 1,
				To:   2,
			}, {
				Name: "symbol",
				From: 4,
				To:   5,
			}, {
				Name: "symbol",
				From: 7,
				To:   8,
			}},
		}},
	}, {
		msg: "list, new lines",
		text: `[
			a
			b
			c
		]`,
		nodes: []*Node{{
			Name: "list",
			To:   20,
			Nodes: []*Node{{
				Name: "symbol",
				From: 5,
				To:   6,
			}, {
				Name: "symbol",
				From: 10,
				To:   11,
			}, {
				Name: "symbol",
				From: 15,
				To:   16,
			}},
		}},
	}, {
		msg:  "list, complex",
		text: "[a, b, c..., [d, e], [f, [g]]...]",
		nodes: []*Node{{
			Name: "list",
			To:   33,
			Nodes: []*Node{{
				Name: "symbol",
				From: 1,
				To:   2,
			}, {
				Name: "symbol",
				From: 4,
				To:   5,
			}, {
				Name: "spread-expression",
				From: 7,
				To:   11,
				Nodes: []*Node{{
					Name: "symbol",
					From: 7,
					To:   8,
				}},
			}, {
				Name: "list",
				From: 13,
				To:   19,
				Nodes: []*Node{{
					Name: "symbol",
					From: 14,
					To:   15,
				}, {
					Name: "symbol",
					From: 17,
					To:   18,
				}},
			}, {
				Name: "spread-expression",
				From: 21,
				To:   32,
				Nodes: []*Node{{
					Name: "list",
					From: 21,
					To:   29,
					Nodes: []*Node{{
						Name: "symbol",
						From: 22,
						To:   23,
					}, {
						Name: "list",
						From: 25,
						To:   28,
						Nodes: []*Node{{
							Name: "symbol",
							From: 26,
							To:   27,
						}},
					}},
				}},
			}},
		}},
	}, {
		msg:  "mutable list",
		text: "~[a, b, c]",
		nodes: []*Node{{
			Name: "mutable-list",
			To:   10,
			Nodes: []*Node{{
				Name: "symbol",
				From: 2,
				To:   3,
			}, {
				Name: "symbol",
				From: 5,
				To:   6,
			}, {
				Name: "symbol",
				From: 8,
				To:   9,
			}},
		}},
	}, {
		msg:  "empty struct",
		text: "{}",
		nodes: []*Node{{
			Name: "struct",
			To:   2,
		}},
	}, {
		msg:  "struct",
		text: "{foo: 1, \"bar\": 2, symbol(baz): 3, [qux]: 4}",
		nodes: []*Node{{
			Name: "struct",
			To:   44,
			Nodes: []*Node{{
				Name: "entry",
				From: 1,
				To:   7,
				Nodes: []*Node{{
					Name: "symbol",
					From: 1,
					To:   4,
				}, {
					Name: "int",
					From: 6,
					To:   7,
				}},
			}, {
				Name: "entry",
				From: 9,
				To:   17,
				Nodes: []*Node{{
					Name: "string",
					From: 9,
					To:   14,
				}, {
					Name: "int",
					From: 16,
					To:   17,
				}},
			}, {
				Name: "entry",
				From: 19,
				To:   33,
				Nodes: []*Node{{
					Name: "dynamic-symbol",
					From: 19,
					To:   30,
					Nodes: []*Node{{
						Name: "symbol",
						From: 26,
						To:   29,
					}},
				}, {
					Name: "int",
					From: 32,
					To:   33,
				}},
			}, {
				Name: "entry",
				From: 35,
				To:   43,
				Nodes: []*Node{{
					Name: "indexer-symbol",
					From: 35,
					To:   40,
					Nodes: []*Node{{
						Name: "symbol",
						From: 36,
						To:   39,
					}},
				}, {
					Name: "int",
					From: 42,
					To:   43,
				}},
			}},
		}},
	}, {
		msg:  "struct, complex",
		text: "{foo: 1, {bar: 2}..., {baz: {}}...}",
		nodes: []*Node{{
			Name: "struct",
			To:   35,
			Nodes: []*Node{{
				Name: "entry",
				From: 1,
				To:   7,
				Nodes: []*Node{{
					Name: "symbol",
					From: 1,
					To:   4,
				}, {
					Name: "int",
					From: 6,
					To:   7,
				}},
			}, {
				Name: "spread-expression",
				From: 9,
				To:   20,
				Nodes: []*Node{{
					Name: "struct",
					From: 9,
					To:   17,
					Nodes: []*Node{{
						Name: "entry",
						From: 10,
						To:   16,
						Nodes: []*Node{{
							Name: "symbol",
							From: 10,
							To:   13,
						}, {
							Name: "int",
							From: 15,
							To:   16,
						}},
					}},
				}},
			}, {
				Name: "spread-expression",
				From: 22,
				To:   34,
				Nodes: []*Node{{
					Name: "struct",
					From: 22,
					To:   31,
					Nodes: []*Node{{
						Name: "entry",
						From: 23,
						To:   30,
						Nodes: []*Node{{
							Name: "symbol",
							From: 23,
							To:   26,
						}, {
							Name: "struct",
							From: 28,
							To:   30,
						}},
					}},
				}},
			}},
		}},
	}, {
		msg:  "struct with indexer key",
		text: "{[a]: b}",
		nodes: []*Node{{
			Name: "struct",
			To:   8,
			Nodes: []*Node{{
				Name: "entry",
				From: 1,
				To:   7,
				Nodes: []*Node{{
					Name: "indexer-symbol",
					From: 1,
					To:   4,
					Nodes: []*Node{{
						Name: "symbol",
						From: 2,
						To:   3,
					}},
				}, {
					Name: "symbol",
					From: 6,
					To:   7,
				}},
			}},
		}},
	}, {
		msg:  "mutable struct",
		text: "~{foo: 1}",
		nodes: []*Node{{
			Name: "mutable-struct",
			To:   9,
			Nodes: []*Node{{
				Name: "entry",
				From: 2,
				To:   8,
				Nodes: []*Node{{
					Name: "symbol",
					From: 2,
					To:   5,
				}, {
					Name: "int",
					From: 7,
					To:   8,
				}},
			}},
		}},
	}, {
		msg:  "channel",
		text: "<>",
		nodes: []*Node{{
			Name: "channel",
			To:   2,
		}},
	}, {
		msg:  "buffered channel",
		text: "<42>",
		nodes: []*Node{{
			Name: "channel",
			To:   4,
			Nodes: []*Node{{
				Name: "int",
				From: 1,
				To:   3,
			}},
		}},
	}, {
		msg:  "and expression",
		text: "and(a, b, c)",
		nodes: []*Node{{
			Name: "function-application",
			To:   12,
			Nodes: []*Node{{
				Name: "symbol",
				To:   3,
			}, {
				Name: "symbol",
				From: 4,
				To:   5,
			}, {
				Name: "symbol",
				From: 7,
				To:   8,
			}, {
				Name: "symbol",
				From: 10,
				To:   11,
			}},
		}},
	}, {
		msg:  "or expression",
		text: "or(a, b, c)",
		nodes: []*Node{{
			Name: "function-application",
			To:   11,
			Nodes: []*Node{{
				Name: "symbol",
				To:   2,
			}, {
				Name: "symbol",
				From: 3,
				To:   4,
			}, {
				Name: "symbol",
				From: 6,
				To:   7,
			}, {
				Name: "symbol",
				From: 9,
				To:   10,
			}},
		}},
	}, {
		msg:  "function",
		text: "fn () 42",
		nodes: []*Node{{
			Name: "function",
			To:   8,
			Nodes: []*Node{{
				Name: "int",
				From: 6,
				To:   8,
			}},
		}},
	}, {
		msg:  "function, noop",
		text: "fn () {;}",
		nodes: []*Node{{
			Name: "function",
			To:   9,
			Nodes: []*Node{{
				Name: "block",
				From: 6,
				To:   9,
			}},
		}},
	}, {
		msg:  "function with args",
		text: "fn (a, b, c) [a, b, c]",
		nodes: []*Node{{
			Name: "function",
			To:   22,
			Nodes: []*Node{{
				Name: "symbol",
				From: 4,
				To:   5,
			}, {
				Name: "symbol",
				From: 7,
				To:   8,
			}, {
				Name: "symbol",
				From: 10,
				To:   11,
			}, {
				Name: "list",
				From: 13,
				To:   22,
				Nodes: []*Node{{
					Name: "symbol",
					From: 14,
					To:   15,
				}, {
					Name: "symbol",
					From: 17,
					To:   18,
				}, {
					Name: "symbol",
					From: 20,
					To:   21,
				}},
			}},
		}},
	}, {
		msg: "function with args in new lines",
		text: `fn (
			a
			b
			c
		) [a, b, c]`,
		nodes: []*Node{{
			Name: "function",
			To:   33,
			Nodes: []*Node{{
				Name: "symbol",
				From: 8,
				To:   9,
			}, {
				Name: "symbol",
				From: 13,
				To:   14,
			}, {
				Name: "symbol",
				From: 18,
				To:   19,
			}, {
				Name: "list",
				From: 24,
				To:   33,
				Nodes: []*Node{{
					Name: "symbol",
					From: 25,
					To:   26,
				}, {
					Name: "symbol",
					From: 28,
					To:   29,
				}, {
					Name: "symbol",
					From: 31,
					To:   32,
				}},
			}},
		}},
	}, {
		msg:  "function with spread arg",
		text: "fn (a, b, ...c) [a, b, c]",
		nodes: []*Node{{
			Name: "function",
			To:   25,
			Nodes: []*Node{{
				Name: "symbol",
				From: 4,
				To:   5,
			}, {
				Name: "symbol",
				From: 7,
				To:   8,
			}, {
				Name: "collect-symbol",
				From: 10,
				To:   14,
				Nodes: []*Node{{
					Name: "symbol",
					From: 13,
					To:   14,
				}},
			}, {
				Name: "list",
				From: 16,
				To:   25,
				Nodes: []*Node{{
					Name: "symbol",
					From: 17,
					To:   18,
				}, {
					Name: "symbol",
					From: 20,
					To:   21,
				}, {
					Name: "symbol",
					From: 23,
					To:   24,
				}},
			}},
		}},
	}, {
		msg:  "effect",
		text: "fn ~ () 42",
		nodes: []*Node{{
			Name: "effect",
			To:   10,
			Nodes: []*Node{{
				Name: "int",
				From: 8,
				To:   10,
			}},
		}},
	}, {
		msg:  "indexer",
		text: "a[42]",
		nodes: []*Node{{
			Name: "indexer",
			To:   5,
			Nodes: []*Node{{
				Name: "symbol",
				To:   1,
			}, {
				Name: "int",
				From: 2,
				To:   4,
			}},
		}},
	}, {
		msg:  "range indexer",
		text: "a[3:9]",
		nodes: []*Node{{
			Name: "indexer",
			To:   6,
			Nodes: []*Node{{
				Name: "symbol",
				To:   1,
			}, {
				Name: "range-from",
				From: 2,
				To:   3,
				Nodes: []*Node{{
					Name: "int",
					From: 2,
					To:   3,
				}},
			}, {
				Name: "range-to",
				From: 4,
				To:   5,
				Nodes: []*Node{{
					Name: "int",
					From: 4,
					To:   5,
				}},
			}},
		}},
	}, {
		msg:  "range indexer, lower unbound",
		text: "a[:9]",
		nodes: []*Node{{
			Name: "indexer",
			To:   5,
			Nodes: []*Node{{
				Name: "symbol",
				To:   1,
			}, {
				Name: "range-to",
				From: 3,
				To:   4,
				Nodes: []*Node{{
					Name: "int",
					From: 3,
					To:   4,
				}},
			}},
		}},
	}, {
		msg:  "range indexer, upper unbound",
		text: "a[3:]",
		nodes: []*Node{{
			Name: "indexer",
			To:   5,
			Nodes: []*Node{{
				Name: "symbol",
				To:   1,
			}, {
				Name: "range-from",
				From: 2,
				To:   3,
				Nodes: []*Node{{
					Name: "int",
					From: 2,
					To:   3,
				}},
			}},
		}},
	}, {
		msg:  "indexer, chained",
		text: "a[b][c][d]",
		nodes: []*Node{{
			Name: "indexer",
			To:   10,
			Nodes: []*Node{{
				Name: "indexer",
				To:   7,
				Nodes: []*Node{{
					Name: "indexer",
					To:   4,
					Nodes: []*Node{{
						Name: "symbol",
						To:   1,
					}, {
						Name: "symbol",
						From: 2,
						To:   3,
					}},
				}, {
					Name: "symbol",
					From: 5,
					To:   6,
				}},
			}, {
				Name: "symbol",
				From: 8,
				To:   9,
			}},
		}},
	}, {
		msg:  "symbol indexer",
		text: "a.b",
		nodes: []*Node{{
			Name: "indexer",
			To:   3,
			Nodes: []*Node{{
				Name: "symbol",
				To:   1,
			}, {
				Name: "symbol",
				From: 2,
				To:   3,
			}},
		}},
	}, {
		msg:  "symbol indexer, with string",
		text: "a.\"b\"",
		nodes: []*Node{{
			Name: "indexer",
			To:   5,
			Nodes: []*Node{{
				Name: "symbol",
				To:   1,
			}, {
				Name: "string",
				From: 2,
				To:   5,
			}},
		}},
	}, {
		msg:  "symbol indexer, with dynamic symbol",
		text: "a.symbol(b)",
		nodes: []*Node{{
			Name: "indexer",
			To:   11,
			Nodes: []*Node{{
				Name: "symbol",
				To:   1,
			}, {
				Name: "dynamic-symbol",
				From: 2,
				To:   11,
				Nodes: []*Node{{
					Name: "symbol",
					From: 9,
					To:   10,
				}},
			}},
		}},
	}, {
		msg:  "chained symbol indexer",
		text: "a.b.c.d",
		nodes: []*Node{{
			Name: "indexer",
			To:   7,
			Nodes: []*Node{{
				Name: "indexer",
				To:   5,
				Nodes: []*Node{{
					Name: "indexer",
					To:   3,
					Nodes: []*Node{{
						Name: "symbol",
						To:   1,
					}, {
						Name: "symbol",
						From: 2,
						To:   3,
					}},
				}, {
					Name: "symbol",
					From: 4,
					To:   5,
				}},
			}, {
				Name: "symbol",
				From: 6,
				To:   7,
			}},
		}},
	}, {
		msg:  "chained symbol indexer on new line",
		text: "a\n.b\n.c",
		nodes: []*Node{{
			Name: "indexer",
			To:   7,
			Nodes: []*Node{{
				Name: "indexer",
				To:   4,
				Nodes: []*Node{{
					Name: "symbol",
					To:   1,
				}, {
					Name: "symbol",
					From: 3,
					To:   4,
				}},
			}, {
				Name: "symbol",
				From: 6,
				To:   7,
			}},
		}},
	}, {
		msg:  "chained symbol indexer on new line after dot",
		text: "a.\nb.\nc",
		nodes: []*Node{{
			Name: "indexer",
			To:   7,
			Nodes: []*Node{{
				Name: "indexer",
				To:   4,
				Nodes: []*Node{{
					Name: "symbol",
					To:   1,
				}, {
					Name: "symbol",
					From: 3,
					To:   4,
				}},
			}, {
				Name: "symbol",
				From: 6,
				To:   7,
			}},
		}},
	}, {
		msg:  "float on a new line",
		text: "f()\n.9",
		nodes: []*Node{{
			Name: "function-application",
			Nodes: []*Node{{
				Name: "symbol",
			}},
		}, {
			Name: "float",
		}},
		ignorePosition: true,
	}, {
		msg:  "function application",
		text: "f()",
		nodes: []*Node{{
			Name: "function-application",
			To:   3,
			Nodes: []*Node{{
				Name: "symbol",
				To:   1,
			}},
		}},
	}, {
		msg:  "function application, single arg",
		text: "f(a)",
		nodes: []*Node{{
			Name: "function-application",
			To:   4,
			Nodes: []*Node{{
				Name: "symbol",
				To:   1,
			}, {
				Name: "symbol",
				From: 2,
				To:   3,
			}},
		}},
	}, {
		msg:  "function application, multiple args",
		text: "f(a, b, c)",
		nodes: []*Node{{
			Name: "function-application",
			To:   10,
			Nodes: []*Node{{
				Name: "symbol",
				To:   1,
			}, {
				Name: "symbol",
				From: 2,
				To:   3,
			}, {
				Name: "symbol",
				From: 5,
				To:   6,
			}, {
				Name: "symbol",
				From: 8,
				To:   9,
			}},
		}},
	}, {
		msg:  "function application, multiple args, new line",
		text: "f(a\nb\nc\n)",
		nodes: []*Node{{
			Name: "function-application",
			To:   9,
			Nodes: []*Node{{
				Name: "symbol",
				To:   1,
			}, {
				Name: "symbol",
				From: 2,
				To:   3,
			}, {
				Name: "symbol",
				From: 4,
				To:   5,
			}, {
				Name: "symbol",
				From: 6,
				To:   7,
			}},
		}},
	}, {
		msg:  "function application, spread",
		text: "f(a, b..., c, d...)",
		nodes: []*Node{{
			Name: "function-application",
			To:   19,
			Nodes: []*Node{{
				Name: "symbol",
				To:   1,
			}, {
				Name: "symbol",
				From: 2,
				To:   3,
			}, {
				Name: "spread-expression",
				From: 5,
				To:   9,
				Nodes: []*Node{{
					Name: "symbol",
					From: 5,
					To:   6,
				}},
			}, {
				Name: "symbol",
				From: 11,
				To:   12,
			}, {
				Name: "spread-expression",
				From: 14,
				To:   18,
				Nodes: []*Node{{
					Name: "symbol",
					From: 14,
					To:   15,
				}},
			}},
		}},
	}, {
		msg:  "chained function application",
		text: "f(a)(b)(c)",
		nodes: []*Node{{
			Name: "function-application",
			To:   10,
			Nodes: []*Node{{
				Name: "function-application",
				To:   7,
				Nodes: []*Node{{
					Name: "function-application",
					To:   4,
					Nodes: []*Node{{
						Name: "symbol",
						To:   1,
					}, {
						Name: "symbol",
						From: 2,
						To:   3,
					}},
				}, {
					Name: "symbol",
					From: 5,
					To:   6,
				}},
			}, {
				Name: "symbol",
				From: 8,
				To:   9,
			}},
		}},
	}, {
		msg:  "embedded function application",
		text: "f(g(h(a)))",
		nodes: []*Node{{
			Name: "function-application",
			To:   10,
			Nodes: []*Node{{
				Name: "symbol",
				To:   1,
			}, {
				Name: "function-application",
				From: 2,
				To:   9,
				Nodes: []*Node{{
					Name: "symbol",
					From: 2,
					To:   3,
				}, {
					Name: "function-application",
					From: 4,
					To:   8,
					Nodes: []*Node{{
						Name: "symbol",
						From: 4,
						To:   5,
					}, {
						Name: "symbol",
						From: 6,
						To:   7,
					}},
				}},
			}},
		}},
	}, {
		msg:  "if",
		text: "if a { b() }",
		nodes: []*Node{{
			Name: "if",
			To:   12,
			Nodes: []*Node{{
				Name: "symbol",
				From: 3,
				To:   4,
			}, {
				Name: "block",
				From: 5,
				To:   12,
				Nodes: []*Node{{
					Name: "function-application",
					From: 7,
					To:   10,
					Nodes: []*Node{{
						Name: "symbol",
						From: 7,
						To:   8,
					}},
				}},
			}},
		}},
	}, {
		msg:  "if, else",
		text: "if a { b } else { c }",
		nodes: []*Node{{
			Name: "if",
			To:   21,
			Nodes: []*Node{{
				Name: "symbol",
				From: 3,
				To:   4,
			}, {
				Name: "block",
				From: 5,
				To:   10,
				Nodes: []*Node{{
					Name: "symbol",
					From: 7,
					To:   8,
				}},
			}, {
				Name: "block",
				From: 16,
				To:   21,
				Nodes: []*Node{{
					Name: "symbol",
					From: 18,
					To:   19,
				}},
			}},
		}},
	}, {
		msg: "if, else if, else if, else",
		text: `
			if a { b }
			else if c { d }
			else if e { f }
			else { g }
		`,
		nodes: []*Node{{
			Name: "if",
			From: 4,
			To:   66,
			Nodes: []*Node{{
				Name: "symbol",
				From: 7,
				To:   8,
			}, {
				Name: "block",
				From: 9,
				To:   14,
				Nodes: []*Node{{
					Name: "symbol",
					From: 11,
					To:   12,
				}},
			}, {
				Name: "symbol",
				From: 26,
				To:   27,
			}, {
				Name: "block",
				From: 28,
				To:   33,
				Nodes: []*Node{{
					Name: "symbol",
					From: 30,
					To:   31,
				}},
			}, {
				Name: "symbol",
				From: 45,
				To:   46,
			}, {
				Name: "block",
				From: 47,
				To:   52,
				Nodes: []*Node{{
					Name: "symbol",
					From: 49,
					To:   50,
				}},
			}, {
				Name: "block",
				From: 61,
				To:   66,
				Nodes: []*Node{{
					Name: "symbol",
					From: 63,
					To:   64,
				}},
			}},
		}},
	}, {
		msg:  "switch, empty",
		text: "switch {default:}",
		nodes: []*Node{{
			Name: "switch",
			To:   17,
			Nodes: []*Node{{
				Name: "default",
				From: 8,
				To:   16,
			}},
		}},
	}, {
		msg: "switch, empty cases",
		text: `
			switch {
			case a:
			case b:
			default:
				f()
			}
		`,
		nodes: []*Node{{
			Name: "switch",
			Nodes: []*Node{{
				Name: "case",
				Nodes: []*Node{{
					Name: "symbol",
				}},
			}, {
				Name: "case",
				Nodes: []*Node{{
					Name: "symbol",
				}},
			}, {
				Name: "default",
			}, {
				Name: "function-application",
				Nodes: []*Node{{
					Name: "symbol",
				}},
			}},
		}},
		ignorePosition: true,
	}, {
		msg:  "switch, single case",
		text: "switch a {case b: c}",
		nodes: []*Node{{
			Name: "switch",
			To:   20,
			Nodes: []*Node{{
				Name: "symbol",
				From: 7,
				To:   8,
			}, {
				Name: "case",
				From: 10,
				To:   17,
				Nodes: []*Node{{
					Name: "symbol",
					From: 15,
					To:   16,
				}},
			}, {
				Name: "symbol",
				From: 18,
				To:   19,
			}},
		}},
	}, {
		msg:  "switch",
		text: "switch a {case b: c; case d: e; default: f}",
		nodes: []*Node{{
			Name: "switch",
			To:   43,
			Nodes: []*Node{{
				Name: "symbol",
				From: 7,
				To:   8,
			}, {
				Name: "case",
				From: 10,
				To:   17,
				Nodes: []*Node{{
					Name: "symbol",
					From: 15,
					To:   16,
				}},
			}, {
				Name: "symbol",
				From: 18,
				To:   19,
			}, {
				Name: "case",
				From: 21,
				To:   28,
				Nodes: []*Node{{
					Name: "symbol",
					From: 26,
					To:   27,
				}},
			}, {
				Name: "symbol",
				From: 29,
				To:   30,
			}, {
				Name: "default",
				From: 32,
				To:   40,
			}, {
				Name: "symbol",
				From: 41,
				To:   42,
			}},
		}},
	}, {
		msg: "switch, all new lines",
		text: `switch
			a
			{
			case
			b
			:
			c
			case
			d
			:
			e
			default
			:
			f
		}`,
		nodes: []*Node{{
			Name: "switch",
			To:   87,
			Nodes: []*Node{{
				Name: "symbol",
				From: 10,
				To:   11,
			}, {
				Name: "case",
				From: 20,
				To:   34,
				Nodes: []*Node{{
					Name: "symbol",
					From: 28,
					To:   29,
				}},
			}, {
				Name: "symbol",
				From: 38,
				To:   39,
			}, {
				Name: "case",
				From: 43,
				To:   57,
				Nodes: []*Node{{
					Name: "symbol",
					From: 51,
					To:   52,
				}},
			}, {
				Name: "symbol",
				From: 61,
				To:   62,
			}, {
				Name: "default",
				From: 66,
				To:   78,
			}, {
				Name: "symbol",
				From: 82,
				To:   83,
			}},
		}},
	}, {
		msg:  "match expression, empty",
		text: "match a {}",
		nodes: []*Node{{
			Name: "match",
			To:   10,
			Nodes: []*Node{{
				Name: "symbol",
				From: 6,
				To:   7,
			}},
		}},
	}, {
		msg: "match expression",
		text: `match a {
			case [first, ...rest]: first
		}`,
		nodes: []*Node{{
			Name: "match",
			To:   45,
			Nodes: []*Node{{
				Name: "symbol",
				From: 6,
				To:   7,
			}, {
				Name: "match-case",
				From: 13,
				To:   35,
				Nodes: []*Node{{
					Name: "list-type",
					From: 18,
					To:   34,
					Nodes: []*Node{{
						Name: "list-destructure-type",
						From: 19,
						To:   33,
						Nodes: []*Node{{
							Name: "destructure-item",
							From: 19,
							To:   24,
							Nodes: []*Node{{
								Name: "symbol",
								From: 19,
								To:   24,
							}},
						}, {
							Name: "collect-destructure-item",
							From: 26,
							To:   33,
							Nodes: []*Node{{
								Name: "destructure-item",
								From: 29,
								To:   33,
								Nodes: []*Node{{
									Name: "symbol",
									From: 29,
									To:   33,
								}},
							}},
						}},
					}},
				}},
			}, {
				Name: "symbol",
				From: 36,
				To:   41,
			}},
		}},
	}, {
		msg: "match expression, multiple cases",
		text: `match a {
			case [0]: []
			case [2:]: a[2:]
			default: error("invalid length")
		}`,
		nodes: []*Node{{
			Name: "match",
			Nodes: []*Node{{
				Name: "symbol",
			}, {
				Name: "match-case",
				Nodes: []*Node{{
					Name: "list-type",
					Nodes: []*Node{{
						Name: "items-type",
						Nodes: []*Node{{
							Name: "items-quantifier",
							Nodes: []*Node{{
								Name: "int",
							}},
						}},
					}},
				}},
			}, {
				Name: "list",
			}, {
				Name: "match-case",
				Nodes: []*Node{{
					Name: "list-type",
					Nodes: []*Node{{
						Name: "items-type",
						Nodes: []*Node{{
							Name: "items-quantifier",
							Nodes: []*Node{{
								Name: "static-range-from",
								Nodes: []*Node{{
									Name: "int",
								}},
							}},
						}},
					}},
				}},
			}, {
				Name: "indexer",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "range-from",
					Nodes: []*Node{{
						Name: "int",
					}},
				}},
			}, {
				Name: "default",
			}, {
				Name: "function-application",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "string",
				}},
			}},
		}},
		ignorePosition: true,
	}, {
		msg: "match function",
		text: `match a {
			case fn () int: a()
			default: 42
		}`,
		nodes: []*Node{{
			Name: "match",
			Nodes: []*Node{{
				Name: "symbol",
			}, {
				Name: "match-case",
				Nodes: []*Node{{
					Name: "function-type",
					Nodes: []*Node{{
						Name: "int-type",
					}},
				}},
			}, {
				Name: "function-application",
				Nodes: []*Node{{
					Name: "symbol",
				}},
			}, {
				Name: "default",
			}, {
				Name: "int",
			}},
		}},
		ignorePosition: true,
	}, {
		msg: "match expression, combined",
		text: `match a {
			case [fn (int)]: a[0]()
			default: 42
		}`,
		nodes: []*Node{{
			Name: "match",
			Nodes: []*Node{{
				Name: "symbol",
			}, {
				Name: "match-case",
				Nodes: []*Node{{
					Name: "list-type",
					Nodes: []*Node{{
						Name: "items-type",
						Nodes: []*Node{{
							Name: "function-type",
							Nodes: []*Node{{
								Name: "arg-type",
								Nodes: []*Node{{
									Name: "int-type",
								}},
							}},
						}},
					}},
				}},
			}, {
				Name: "function-application",
				Nodes: []*Node{{
					Name: "indexer",
					Nodes: []*Node{{
						Name: "symbol",
					}, {
						Name: "int",
					}},
				}},
			}, {
				Name: "default",
			}, {
				Name: "int",
			}},
		}},
		ignorePosition: true,
	}, {
		msg: "match expression, complex",
		text: `match a {
				case [first T int|string, op fn ([T, int, ...T]) int, ...rest T]:
					op([first, now(), rest...])
				default:
					error("invalid list")
			}`,
		nodes: []*Node{{
			Name: "match",
			Nodes: []*Node{{
				Name: "symbol",
			}, {
				Name: "match-case",
				Nodes: []*Node{{
					Name: "list-match",
					Nodes: []*Node{{
						Name: "list-destructure-match",
						Nodes: []*Node{{
							Name: "destructure-match-item",
							Nodes: []*Node{{
								Name: "symbol",
							}, {
								Name: "symbol",
							}, {
								Name: "int-type",
							}, {
								Name: "string-type",
							}},
						}, {
							Name: "destructure-match-item",
							Nodes: []*Node{{
								Name: "symbol",
							}, {
								Name: "function-type",
								Nodes: []*Node{{
									Name: "arg-type",
									Nodes: []*Node{{
										Name: "list-type",
										Nodes: []*Node{{
											Name: "list-destructure-type",
											Nodes: []*Node{{
												Name: "destructure-item",
												Nodes: []*Node{{
													Name: "symbol",
												}},
											}, {
												Name: "destructure-item",
												Nodes: []*Node{{
													Name: "int-type",
												}},
											}, {
												Name: "collect-destructure-item",
												Nodes: []*Node{{
													Name: "destructure-item",
													Nodes: []*Node{{
														Name: "symbol",
													}},
												}},
											}},
										}},
									}},
								}, {
									Name: "int-type",
								}},
							}},
						}, {
							Name: "collect-destructure-match-item",
							Nodes: []*Node{{
								Name: "destructure-match-item",
								Nodes: []*Node{{
									Name: "symbol",
								}, {
									Name: "symbol",
								}},
							}},
						}},
					}},
				}},
			}, {
				Name: "function-application",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "list",
					Nodes: []*Node{{
						Name: "symbol",
					}, {
						Name: "function-application",
						Nodes: []*Node{{
							Name: "symbol",
						}},
					}, {
						Name: "spread-expression",
						Nodes: []*Node{{
							Name: "symbol",
						}},
					}},
				}},
			}, {
				Name: "default",
			}, {
				Name: "function-application",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "string",
				}},
			}},
		}},
		ignorePosition: true,
	}, {
		msg:  "receive op",
		text: "<-chan",
		nodes: []*Node{{
			Name: "unary-expression",
			Nodes: []*Node{{
				Name: "receive-op",
				Nodes: []*Node{{
					Name: "symbol",
				}},
			}},
		}},
		ignorePosition: true,
	}, {
		msg:  "send op",
		text: "chan <- a",
		nodes: []*Node{{
			Name: "send",
			Nodes: []*Node{{
				Name: "symbol",
			}, {
				Name: "symbol",
			}},
		}},
		ignorePosition: true,
	}, {
		msg: "select, empty",
		text: `select {
		}`,
		nodes: []*Node{{
			Name: "select",
			To:   12,
		}},
	}, {
		msg: "select",
		text: `select {
			case let a <-r: s <- a
			case s <- f(): g()
			default: h()
		}`,
		nodes: []*Node{{
			Name: "select",
			Nodes: []*Node{{
				Name: "select-case",
				Nodes: []*Node{{
					Name: "receive-definition",
					Nodes: []*Node{{
						Name: "symbol",
					}, {
						Name: "receive-op",
						Nodes: []*Node{{
							Name: "symbol",
						}},
					}},
				}},
			}, {
				Name: "send",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "symbol",
				}},
			}, {
				Name: "select-case",
				Nodes: []*Node{{
					Name: "send",
					Nodes: []*Node{{
						Name: "symbol",
					}, {
						Name: "function-application",
						Nodes: []*Node{{
							Name: "symbol",
						}},
					}},
				}},
			}, {
				Name: "function-application",
				Nodes: []*Node{{
					Name: "symbol",
				}},
			}, {
				Name: "default",
			}, {
				Name: "function-application",
				Nodes: []*Node{{
					Name: "symbol",
				}},
			}},
		}},
		ignorePosition: true,
	}, {
		msg: "select, call",
		text: `select {
			case let a receive(r): f()
			case send(s, g()): h()
			default: i()
		}`,
		nodes: []*Node{{
			Name: "select",
			Nodes: []*Node{{
				Name: "select-case",
				Nodes: []*Node{{
					Name: "receive-definition",
					Nodes: []*Node{{
						Name: "symbol",
					}, {
						Name: "receive-call",
						Nodes: []*Node{{
							Name: "symbol",
						}},
					}},
				}},
			}, {
				Name: "function-application",
				Nodes: []*Node{{
					Name: "symbol",
				}},
			}, {
				Name: "select-case",
				Nodes: []*Node{{
					Name: "send",
					Nodes: []*Node{{
						Name: "symbol",
					}, {
						Name: "function-application",
						Nodes: []*Node{{
							Name: "symbol",
						}},
					}},
				}},
			}, {
				Name: "function-application",
				Nodes: []*Node{{
					Name: "symbol",
				}},
			}, {
				Name: "default",
			}, {
				Name: "function-application",
				Nodes: []*Node{{
					Name: "symbol",
				}},
			}},
		}},
		ignorePosition: true,
	}, {
		msg:  "go",
		text: "go f()",
		nodes: []*Node{{
			Name: "go",
			Nodes: []*Node{{
				Name: "function-application",
				Nodes: []*Node{{
					Name: "symbol",
				}},
			}},
		}},
		ignorePosition: true,
	}, {
		msg:  "go, block",
		text: "go { for { f() } }",
		nodes: []*Node{{
			Name: "go",
			Nodes: []*Node{{
				Name: "block",
				Nodes: []*Node{{
					Name: "loop",
					Nodes: []*Node{{
						Name: "block",
						Nodes: []*Node{{
							Name: "function-application",
							Nodes: []*Node{{
								Name: "symbol",
							}},
						}},
					}},
				}},
			}},
		}},
		ignorePosition: true,
	}, {
		msg:  "require, dot, equal",
		text: "require . = \"mml/foo\"",
		nodes: []*Node{{
			Name: "require",
			Nodes: []*Node{{
				Name: "require-fact",
				Nodes: []*Node{{
					Name: "require-inline",
				}, {
					Name: "string",
				}},
			}},
		}},
		ignorePosition: true,
	}, {
		msg:  "require, symbol, equal",
		text: "require bar = \"mml/foo\"",
		nodes: []*Node{{
			Name: "require",
			Nodes: []*Node{{
				Name: "require-fact",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "string",
				}},
			}},
		}},
		ignorePosition: true,
	}, {
		msg:  "require, symbol",
		text: "require bar \"mml/foo\"",
		nodes: []*Node{{
			Name: "require",
			Nodes: []*Node{{
				Name: "require-fact",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "string",
				}},
			}},
		}},
		ignorePosition: true,
	}, {
		msg:  "require",
		text: "require \"mml/foo\"",
		nodes: []*Node{{
			Name: "require",
			Nodes: []*Node{{
				Name: "require-fact",
				Nodes: []*Node{{
					Name: "string",
				}},
			}},
		}},
		ignorePosition: true,
	}, {
		msg: "require, group",
		text: `require (
			. = "mml/foo"
			bar = "mml/foo"
			. "mml/foo"
			bar "mml/foo"
			"mml/foo"
		)`,
		nodes: []*Node{{
			Name: "require",
			Nodes: []*Node{{
				Name: "require-fact",
				Nodes: []*Node{{
					Name: "require-inline",
				}, {
					Name: "string",
				}},
			}, {
				Name: "require-fact",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "string",
				}},
			}, {
				Name: "require-fact",
				Nodes: []*Node{{
					Name: "require-inline",
				}, {
					Name: "string",
				}},
			}, {
				Name: "require-fact",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "string",
				}},
			}, {
				Name: "require-fact",
				Nodes: []*Node{{
					Name: "string",
				}},
			}},
		}},
		ignorePosition: true,
	}, {
		msg:  "expression group",
		text: "(fn (a) a)(a)",
		nodes: []*Node{{
			Name: "function-application",
			Nodes: []*Node{{
				Name: "function",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "symbol",
				}},
			}, {
				Name: "symbol",
			}},
		}},
		ignorePosition: true,
	}, {
		msg:  "unary operator",
		text: "!foo",
		nodes: []*Node{{
			Name: "unary-expression",
			Nodes: []*Node{{
				Name: "logical-not",
			}, {
				Name: "symbol",
			}},
		}},
		ignorePosition: true,
	}, {
		msg:  "binary 0",
		text: "a * b",
		nodes: []*Node{{
			Name: "binary0",
			Nodes: []*Node{{
				Name: "symbol",
			}, {
				Name: "mul",
			}, {
				Name: "symbol",
			}},
		}},
		ignorePosition: true,
	}, {
		msg:  "binary 1",
		text: "a * b + c * d",
		nodes: []*Node{{
			Name: "binary1",
			Nodes: []*Node{{
				Name: "binary0",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "mul",
				}, {
					Name: "symbol",
				}},
			}, {
				Name: "add",
			}, {
				Name: "binary0",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "mul",
				}, {
					Name: "symbol",
				}},
			}},
		}},
		ignorePosition: true,
	}, {
		msg:  "binary 2",
		text: "a * b + c * d == e * f",
		nodes: []*Node{{
			Name: "binary2",
			Nodes: []*Node{{
				Name: "binary1",
				Nodes: []*Node{{
					Name: "binary0",
					Nodes: []*Node{{
						Name: "symbol",
					}, {
						Name: "mul",
					}, {
						Name: "symbol",
					}},
				}, {
					Name: "add",
				}, {
					Name: "binary0",
					Nodes: []*Node{{
						Name: "symbol",
					}, {
						Name: "mul",
					}, {
						Name: "symbol",
					}},
				}},
			}, {
				Name: "eq",
			}, {
				Name: "binary0",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "mul",
				}, {
					Name: "symbol",
				}},
			}},
		}},
		ignorePosition: true,
	}, {
		msg:  "binary 3, 4, 5",
		text: "a * b + c * d == e * f && g || h -> f()",
		nodes: []*Node{{
			Name: "binary5",
			Nodes: []*Node{{
				Name: "binary4",
				Nodes: []*Node{{
					Name: "binary3",
					Nodes: []*Node{{
						Name: "binary2",
						Nodes: []*Node{{
							Name: "binary1",
							Nodes: []*Node{{
								Name: "binary0",
								Nodes: []*Node{{
									Name: "symbol",
								}, {
									Name: "mul",
								}, {
									Name: "symbol",
								}},
							}, {
								Name: "add",
							}, {
								Name: "binary0",
								Nodes: []*Node{{
									Name: "symbol",
								}, {
									Name: "mul",
								}, {
									Name: "symbol",
								}},
							}},
						}, {
							Name: "eq",
						}, {
							Name: "binary0",
							Nodes: []*Node{{
								Name: "symbol",
							}, {
								Name: "mul",
							}, {
								Name: "symbol",
							}},
						}},
					}, {
						Name: "logical-and",
					}, {
						Name: "symbol",
					}},
				}, {
					Name: "logical-or",
				}, {
					Name: "symbol",
				}},
			}, {
				Name: "chain",
			}, {
				Name: "function-application",
				Nodes: []*Node{{
					Name: "symbol",
				}},
			}},
		}},
		ignorePosition: true,
	}, {
		msg:  "ternary expression",
		text: "a ? b : c",
		nodes: []*Node{{
			Name: "ternary-expression",
			To:   9,
			Nodes: []*Node{{
				Name: "symbol",
				To:   1,
			}, {
				Name: "symbol",
				From: 4,
				To:   5,
			}, {
				Name: "symbol",
				From: 8,
				To:   9,
			}},
		}},
	}, {
		msg:  "multiple ternary expressions, consequence",
		text: "a ? b ? c : d : e",
		nodes: []*Node{{
			Name: "ternary-expression",
			To:   17,
			Nodes: []*Node{{
				Name: "symbol",
				To:   1,
			}, {
				Name: "ternary-expression",
				From: 4,
				To:   13,
				Nodes: []*Node{{
					Name: "symbol",
					From: 4,
					To:   5,
				}, {
					Name: "symbol",
					From: 8,
					To:   9,
				}, {
					Name: "symbol",
					From: 12,
					To:   13,
				}},
			}, {
				Name: "symbol",
				From: 16,
				To:   17,
			}},
		}},
	}, {
		msg:  "multiple ternary expressions, alternative",
		text: "a ? b : c ? d : e",
		nodes: []*Node{{
			Name: "ternary-expression",
			To:   17,
			Nodes: []*Node{{
				Name: "symbol",
				To:   1,
			}, {
				Name: "symbol",
				From: 4,
				To:   5,
			}, {
				Name: "ternary-expression",
				From: 8,
				To:   17,
				Nodes: []*Node{{
					Name: "symbol",
					From: 8,
					To:   9,
				}, {
					Name: "symbol",
					From: 12,
					To:   13,
				}, {
					Name: "symbol",
					From: 16,
					To:   17,
				}},
			}},
		}},
	}, {
		msg:  "infinite loop",
		text: "for {}",
		nodes: []*Node{{
			Name: "loop",
			Nodes: []*Node{{
				Name: "block",
			}},
		}},
		ignorePosition: true,
	}, {
		msg:  "conditional loop",
		text: "for foo {}",
		nodes: []*Node{{
			Name: "loop",
			Nodes: []*Node{{
				Name: "loop-expression",
				Nodes: []*Node{{
					Name: "symbol",
				}},
			}, {
				Name: "block",
			}},
		}},
		ignorePosition: true,
	}, {
		msg:  "in list loop",
		text: "for i in [1, 2, 3] {}",
		nodes: []*Node{{
			Name: "loop",
			Nodes: []*Node{{
				Name: "loop-expression",
				Nodes: []*Node{{
					Name: "in-expression",
					Nodes: []*Node{{
						Name: "symbol",
					}, {
						Name: "list",
						Nodes: []*Node{{
							Name: "int",
						}, {
							Name: "int",
						}, {
							Name: "int",
						}},
					}},
				}},
			}, {
				Name: "block",
			}},
		}},
		ignorePosition: true,
	}, {
		msg:  "in range loop",
		text: "for i in -3:42 {}",
		nodes: []*Node{{
			Name: "loop",
			Nodes: []*Node{{
				Name: "loop-expression",
				Nodes: []*Node{{
					Name: "in-expression",
					Nodes: []*Node{{
						Name: "symbol",
					}, {
						Name: "range-from",
						Nodes: []*Node{{
							Name: "unary-expression",
							Nodes: []*Node{{
								Name: "minus",
							}, {
								Name: "int",
							}},
						}},
					}, {
						Name: "range-to",
						Nodes: []*Node{{
							Name: "int",
						}},
					}},
				}},
			}, {
				Name: "block",
			}},
		}},
		ignorePosition: true,
	}, {
		msg: "loop control",
		text: `for i in l {
			if i % 2 == 0 {
				break
			}
		}`,
		nodes: []*Node{{
			Name: "loop",
			Nodes: []*Node{{
				Name: "loop-expression",
				Nodes: []*Node{{
					Name: "in-expression",
					Nodes: []*Node{{
						Name: "symbol",
					}, {
						Name: "symbol",
					}},
				}},
			}, {
				Name: "block",
				Nodes: []*Node{{
					Name: "if",
					Nodes: []*Node{{
						Name: "binary2",
						Nodes: []*Node{{
							Name: "binary0",
							Nodes: []*Node{{
								Name: "symbol",
							}, {
								Name: "mod",
							}, {
								Name: "int",
							}},
						}, {
							Name: "eq",
						}, {
							Name: "int",
						}},
					}, {
						Name: "block",
						Nodes: []*Node{{
							Name: "break",
						}},
					}},
				}},
			}},
		}},
		ignorePosition: true,
	}, {
		msg:  "assign, eq",
		text: "a = b",
		nodes: []*Node{{
			Name: "assignment",
			Nodes: []*Node{{
				Name: "assign-equal",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "symbol",
				}},
			}},
		}},
		ignorePosition: true,
	}, {
		msg:  "assign, set, eq",
		text: "set a = b",
		nodes: []*Node{{
			Name: "assignment",
			Nodes: []*Node{{
				Name: "assign-capture",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "symbol",
				}},
			}},
		}},
		ignorePosition: true,
	}, {
		msg:  "assign, set",
		text: "set a b",
		nodes: []*Node{{
			Name: "assignment",
			Nodes: []*Node{{
				Name: "assign-capture",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "symbol",
				}},
			}},
		}},
		ignorePosition: true,
	}, {
		msg: "assign, group",
		text: `set (
			a = b
			c d
		)`,
		nodes: []*Node{{
			Name: "assignment",
			Nodes: []*Node{{
				Name: "assign-capture",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "symbol",
				}},
			}, {
				Name: "assign-capture",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "symbol",
				}},
			}},
		}},
		ignorePosition: true,
	}, {
		msg:  "define, eq",
		text: "let a = b",
		nodes: []*Node{{
			Name: "value-definition",
			Nodes: []*Node{{
				Name: "value-capture",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "symbol",
				}},
			}},
		}},
		ignorePosition: true,
	}, {
		msg:  "define",
		text: "let a b",
		nodes: []*Node{{
			Name: "value-definition",
			Nodes: []*Node{{
				Name: "value-capture",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "symbol",
				}},
			}},
		}},
		ignorePosition: true,
	}, {
		msg:  "define mutable, eq",
		text: "let ~ a = b",
		nodes: []*Node{{
			Name: "value-definition",
			Nodes: []*Node{{
				Name: "mutable-capture",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "symbol",
				}},
			}},
		}},
		ignorePosition: true,
	}, {
		msg:  "define mutable",
		text: "let ~ a b",
		nodes: []*Node{{
			Name: "value-definition",
			Nodes: []*Node{{
				Name: "mutable-capture",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "symbol",
				}},
			}},
		}},
		ignorePosition: true,
	}, {
		msg: "mixed define group",
		text: `let (
			a = b
			c d
			~ e f
			~ g h
		)`,
		nodes: []*Node{{
			Name: "value-definition-group",
			Nodes: []*Node{{
				Name: "value-capture",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "symbol",
				}},
			}, {
				Name: "value-capture",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "symbol",
				}},
			}, {
				Name: "mutable-capture",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "symbol",
				}},
			}, {
				Name: "mutable-capture",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "symbol",
				}},
			}},
		}},
		ignorePosition: true,
	}, {
		msg: "mutable define group",
		text: `let ~ (
			a = b
			c d
		)`,
		nodes: []*Node{{
			Name: "mutable-definition-group",
			Nodes: []*Node{{
				Name: "value-capture",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "symbol",
				}},
			}, {
				Name: "value-capture",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "symbol",
				}},
			}},
		}},
		ignorePosition: true,
	}, {
		msg:  "define function",
		text: "fn a() b",
		nodes: []*Node{{
			Name: "function-definition",
			Nodes: []*Node{{
				Name: "function-capture",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "symbol",
				}},
			}},
		}},
		ignorePosition: true,
	}, {
		msg:  "define effect",
		text: "fn ~ a() b",
		nodes: []*Node{{
			Name: "function-definition",
			Nodes: []*Node{{
				Name: "effect-capture",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "symbol",
				}},
			}},
		}},
		ignorePosition: true,
	}, {
		msg: "define function group",
		text: `fn (
			a() b
			~ c() d
		)`,
		nodes: []*Node{{
			Name: "function-definition-group",
			Nodes: []*Node{{
				Name: "function-capture",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "symbol",
				}},
			}, {
				Name: "effect-capture",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "symbol",
				}},
			}},
		}},
		ignorePosition: true,
	}, {
		msg: "define effect group",
		text: `fn ~ (
			a() b
			c() d
		)`,
		nodes: []*Node{{
			Name: "effect-definition-group",
			Nodes: []*Node{{
				Name: "function-capture",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "symbol",
				}},
			}, {
				Name: "function-capture",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "symbol",
				}},
			}},
		}},
		ignorePosition: true,
	}, {
		msg: "type constraint",
		text: `
			type a fn ([]) int
			fn a(l) len(l)
		`,
		nodes: []*Node{{
			Name: "type-constraint",
			Nodes: []*Node{{
				Name: "symbol",
			}, {
				Name: "function-type",
				Nodes: []*Node{{
					Name: "arg-type",
					Nodes: []*Node{{
						Name: "list-type",
					}},
				}, {
					Name: "int-type",
				}},
			}},
		}, {
			Name: "function-definition",
			Nodes: []*Node{{
				Name: "function-capture",
				Nodes: []*Node{{
					Name: "symbol",
				}, {
					Name: "symbol",
				}, {
					Name: "function-application",
					Nodes: []*Node{{
						Name: "symbol",
					}, {
						Name: "symbol",
					}},
				}},
			}},
		}},
		ignorePosition: true,
	}, {
		msg:  "type alias",
		text: "type alias a int|(fn () int|string)|string",
		nodes: []*Node{{
			Name: "type-alias",
			Nodes: []*Node{{
				Name: "symbol",
			}, {
				Name: "int-type",
			}, {
				Name: "function-type",
				Nodes: []*Node{{
					Name: "int-type",
				}, {
					Name: "string-type",
				}},
			}, {
				Name: "string-type",
			}},
		}},
		ignorePosition: true,
	}, {
		msg:  "statement group",
		text: "(for {})",
		nodes: []*Node{{
			Name: "loop",
			Nodes: []*Node{{
				Name: "block",
			}},
		}},
		ignorePosition: true,
	}})
}
