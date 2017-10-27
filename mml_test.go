package treerack

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"
)

func TestMML(t *testing.T) {
	runTestsFile(t, "mml.parser", []testItem{{
		title: "empty",
		node:  &Node{Name: "mml"},
	}, {
		title: "single line comment",
		text:  "// foo bar baz",
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
		title: "multiple line comments",
		text:  "// foo bar\n// baz qux",
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
		title: "block comment",
		text:  "/* foo bar baz */",
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
		title: "block comments",
		text:  "/* foo bar */\n/* baz qux */",
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
		title: "mixed comments",
		text:  "// foo\n/* bar */\n// baz",
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
		title: "int",
		text:  "42",
		nodes: []*Node{{
			Name: "int",
			To:   2,
		}},
	}, {
		title: "ints",
		text:  "1; 2; 3",
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
		title: "int, octal",
		text:  "052",
		nodes: []*Node{{
			Name: "int",
			To:   3,
		}},
	}, {
		title: "int, hexa",
		text:  "0x2a",
		nodes: []*Node{{
			Name: "int",
			To:   4,
		}},
	}, {
		title: "float, 0.",
		text:  "0.",
		nodes: []*Node{{
			Name: "float",
			To:   2,
		}},
	}, {
		title: "float, 72.40",
		text:  "72.40",
		nodes: []*Node{{
			Name: "float",
			To:   5,
		}},
	}, {
		title: "float, 072.40",
		text:  "072.40",
		nodes: []*Node{{
			Name: "float",
			To:   6,
		}},
	}, {
		title: "float, 2.71828",
		text:  "2.71828",
		nodes: []*Node{{
			Name: "float",
			To:   7,
		}},
	}, {
		title: "float, 6.67428e-11",
		text:  "6.67428e-11",
		nodes: []*Node{{
			Name: "float",
			To:   11,
		}},
	}, {
		title: "float, 1E6",
		text:  "1E6",
		nodes: []*Node{{
			Name: "float",
			To:   3,
		}},
	}, {
		title: "float, .25",
		text:  ".25",
		nodes: []*Node{{
			Name: "float",
			To:   3,
		}},
	}, {
		title: "float, .12345E+5",
		text:  ".12345E+5",
		nodes: []*Node{{
			Name: "float",
			To:   9,
		}},
	}, {
		title: "string, empty",
		text:  "\"\"",
		nodes: []*Node{{
			Name: "string",
			To:   2,
		}},
	}, {
		title: "string",
		text:  "\"foo\"",
		nodes: []*Node{{
			Name: "string",
			To:   5,
		}},
	}, {
		title: "string, with new line",
		text:  "\"foo\nbar\"",
		nodes: []*Node{{
			Name: "string",
			To:   9,
		}},
	}, {
		title: "string, with escaped new line",
		text:  "\"foo\\nbar\"",
		nodes: []*Node{{
			Name: "string",
			To:   10,
		}},
	}, {
		title: "string, with quotes",
		text:  "\"foo \\\"bar\\\" baz\"",
		nodes: []*Node{{
			Name: "string",
			To:   17,
		}},
	}, {
		title: "bool, true",
		text:  "true",
		nodes: []*Node{{
			Name: "true",
			To:   4,
		}},
	}, {
		title: "bool, false",
		text:  "false",
		nodes: []*Node{{
			Name: "false",
			To:   5,
		}},
	}, {
		title: "symbol",
		text:  "foo",
		nodes: []*Node{{
			Name: "symbol",
			To:   3,
		}},
	}, {
		title: "dynamic-symbol",
		text:  "symbol(a)",
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
		title: "empty list",
		text:  "[]",
		nodes: []*Node{{
			Name: "list",
			To:   2,
		}},
	}, {
		title: "list",
		text:  "[a, b, c]",
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
		title: "list, new lines",
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
		title: "list, complex",
		text:  "[a, b, c..., [d, e], [f, [g]]...]",
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
		title: "mutable list",
		text:  "~[a, b, c]",
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
		title: "empty struct",
		text:  "{}",
		nodes: []*Node{{
			Name: "struct",
			To:   2,
		}},
	}, {
		title: "struct",
		text:  "{foo: 1, \"bar\": 2, symbol(baz): 3, [qux]: 4}",
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
		title: "struct, complex",
		text:  "{foo: 1, {bar: 2}..., {baz: {}}...}",
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
		title: "struct with indexer key",
		text:  "{[a]: b}",
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
		title: "mutable struct",
		text:  "~{foo: 1}",
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
		title: "channel",
		text:  "<>",
		nodes: []*Node{{
			Name: "channel",
			To:   2,
		}},
	}, {
		title: "buffered channel",
		text:  "<42>",
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
		title: "and expression",
		text:  "and(a, b, c)",
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
		title: "or expression",
		text:  "or(a, b, c)",
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
		title: "function",
		text:  "fn () 42",
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
		title: "function, noop",
		text:  "fn () {;}",
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
		title: "function with args",
		text:  "fn (a, b, c) [a, b, c]",
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
		title: "function with args in new lines",
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
		title: "function with spread arg",
		text:  "fn (a, b, ...c) [a, b, c]",
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
		title: "effect",
		text:  "fn ~ () 42",
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
		title: "indexer",
		text:  "a[42]",
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
		title: "range indexer",
		text:  "a[3:9]",
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
		title: "range indexer, lower unbound",
		text:  "a[:9]",
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
		title: "range indexer, upper unbound",
		text:  "a[3:]",
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
		title: "indexer, chained",
		text:  "a[b][c][d]",
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
		title: "symbol indexer",
		text:  "a.b",
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
		title: "symbol indexer, with string",
		text:  "a.\"b\"",
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
		title: "symbol indexer, with dynamic symbol",
		text:  "a.symbol(b)",
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
		title: "chained symbol indexer",
		text:  "a.b.c.d",
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
		title: "chained symbol indexer on new line",
		text:  "a\n.b\n.c",
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
		title: "chained symbol indexer on new line after dot",
		text:  "a.\nb.\nc",
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
		title: "float on a new line",
		text:  "f()\n.9",
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
		title: "function application",
		text:  "f()",
		nodes: []*Node{{
			Name: "function-application",
			To:   3,
			Nodes: []*Node{{
				Name: "symbol",
				To:   1,
			}},
		}},
	}, {
		title: "function application, single arg",
		text:  "f(a)",
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
		title: "function application, multiple args",
		text:  "f(a, b, c)",
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
		title: "function application, multiple args, new line",
		text:  "f(a\nb\nc\n)",
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
		title: "function application, spread",
		text:  "f(a, b..., c, d...)",
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
		title: "chained function application",
		text:  "f(a)(b)(c)",
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
		title: "embedded function application",
		text:  "f(g(h(a)))",
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
		title: "if",
		text:  "if a { b() }",
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
		title: "if, else",
		text:  "if a { b } else { c }",
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
		title: "if, else if, else if, else",
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
		title: "switch, empty",
		text:  "switch {default:}",
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
		title: "switch, empty cases",
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
		title: "switch, single case",
		text:  "switch a {case b: c}",
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
		title: "switch",
		text:  "switch a {case b: c; case d: e; default: f}",
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
		title: "switch, all new lines",
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
		title: "match expression, empty",
		text:  "match a {}",
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
		title: "match expression",
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
		title: "match expression, multiple cases",
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
		title: "match function",
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
		title: "match expression, combined",
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
		title: "match expression, complex",
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
		title: "receive op",
		text:  "<-chan",
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
		title: "send op",
		text:  "chan <- a",
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
		title: "select, empty",
		text: `select {
		}`,
		nodes: []*Node{{
			Name: "select",
			To:   12,
		}},
	}, {
		title: "select",
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
		title: "select, call",
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
		title:          "block",
		ignorePosition: true,
		text:           "{ f() }",
		nodes: []*Node{{
			Name: "block",
			Nodes: []*Node{{
				Name: "function-application",
				Nodes: []*Node{{
					Name: "symbol",
				}},
			}},
		}},
	}, {
		title: "go",
		text:  "go f()",
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
		title: "go, block",
		text:  "go { for { f() } }",
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
		title: "require, dot, equal",
		text:  "require . = \"mml/foo\"",
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
		title: "require, symbol, equal",
		text:  "require bar = \"mml/foo\"",
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
		title: "require, symbol",
		text:  "require bar \"mml/foo\"",
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
		title: "require",
		text:  "require \"mml/foo\"",
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
		title: "require, group",
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
		title: "expression group",
		text:  "(fn (a) a)(a)",
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
		title: "unary operator",
		text:  "!foo",
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
		title: "binary 0",
		text:  "a * b",
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
		title: "binary 1",
		text:  "a * b + c * d",
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
		title: "binary 2",
		text:  "a * b + c * d == e * f",
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
		title: "binary 3, 4, 5",
		text:  "a * b + c * d == e * f && g || h -> f()",
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
		title: "ternary expression",
		text:  "a ? b : c",
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
		title: "multiple ternary expressions, consequence",
		text:  "a ? b ? c : d : e",
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
		title: "multiple ternary expressions, alternative",
		text:  "a ? b : c ? d : e",
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
		title: "infinite loop",
		text:  "for {}",
		nodes: []*Node{{
			Name: "loop",
			Nodes: []*Node{{
				Name: "block",
			}},
		}},
		ignorePosition: true,
	}, {
		title: "conditional loop",
		text:  "for foo {}",
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
		title: "in list loop",
		text:  "for i in [1, 2, 3] {}",
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
		title: "in range loop",
		text:  "for i in -3:42 {}",
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
		title: "loop control",
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
		title: "assign, eq",
		text:  "a = b",
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
		title: "assign, set, eq",
		text:  "set a = b",
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
		title: "assign, set",
		text:  "set a b",
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
		title: "assign, group",
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
		title: "define, eq",
		text:  "let a = b",
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
		title: "define",
		text:  "let a b",
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
		title: "define mutable, eq",
		text:  "let ~ a = b",
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
		title: "define mutable",
		text:  "let ~ a b",
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
		title: "mixed define group",
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
		title: "mutable define group",
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
		title: "define function",
		text:  "fn a() b",
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
		title: "define effect",
		text:  "fn ~ a() b",
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
		title: "define function group",
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
		title: "define effect group",
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
		title: "type constraint",
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
		title: "type alias",
		text:  "type alias a int|(fn () int|string)|string",
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
		title: "statement group",
		text:  "(for {})",
		nodes: []*Node{{
			Name: "loop",
			Nodes: []*Node{{
				Name: "block",
			}},
		}},
		ignorePosition: true,
	}})
}

func TestMMLFile(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	const n = 180

	s, err := openSyntaxFile("mml.parser")
	if err != nil {
		t.Error(err)
		return
	}

	s.Init()

	f, err := os.Open("test.mml")
	if err != nil {
		t.Error(err)
		return
	}

	defer f.Close()

	var d time.Duration
	for i := 0; i < n && !t.Failed(); i++ {
		func() {
			if _, err := f.Seek(0, 0); err != nil {
				t.Error(err)
				return
			}

			b := bytes.NewBuffer(nil)
			if _, err := io.Copy(b, f); err != nil {
				t.Error(err)
				return
			}

			start := time.Now()
			_, err = s.Parse(b)
			d += time.Now().Sub(start)

			if err != nil {
				t.Error(err)
			}
		}()
	}

	t.Log("average duration:", d/n)
}
