package treerack

var bootSyntaxDefs = [][]string{{
	"chars", "space", "alias", " ",
}, {
	"chars", "tab", "alias", "\\t",
}, {
	"chars", "nl", "alias", "\\n",
}, {
	"chars", "backspace", "alias", "\\b",
}, {
	"chars", "formfeed", "alias", "\\f",
}, {
	"chars", "carryreturn", "alias", "\\r",
}, {
	"chars", "verticaltab", "alias", "\\v",
}, {
	"choice",
	"ws",
	"alias",
	"space",
	"tab",
	"nl",
	"backspace",
	"formfeed",
	"carryreturn",
	"verticaltab",
}, {
	"chars", "open-block-comment", "alias", "/*",
}, {
	"chars", "close-block-comment", "alias", "*/",
}, {
	"chars", "star", "alias", "*",
}, {
	"class", "not-slash", "alias", "^/",
}, {
	"class", "not-star", "alias", "^*",
}, {
	"chars", "double-slash", "alias", "//",
}, {
	"class", "not-nl", "alias", "^\\n",
}, {
	"sequence", "not-block-close", "alias", "star", "not-slash",
}, {
	"choice", "block-comment-char", "alias", "not-block-close", "not-star",
}, {
	"sequence", "block-comment-body", "alias", "block-comment-char:0:-1",
}, {
	"sequence",
	"block-comment",
	"alias",
	"open-block-comment",
	"block-comment-body",
	"close-block-comment",
}, {
	"sequence", "not-nls", "alias", "not-nl:0:-1",
}, {
	"sequence", "line-comment", "alias", "double-slash", "not-nls",
}, {
	"choice", "comment-segment", "alias", "block-comment", "line-comment",
}, {
	"sequence", "wss", "alias", "ws:0:-1",
}, {
	"sequence", "optional-nl", "alias", "nl:0:1",
}, {
	"choice",
	"ws-no-nl",
	"alias",
	"space",
	"tab",
	"backspace",
	"formfeed",
	"carryreturn",
	"verticaltab",
}, {
	"sequence",
	"continue-comment-segment",
	"alias",
	"ws-no-nl",
	"optional-nl",
	"ws-no-nl",
	"comment-segment",
}, {
	"sequence", "continue-comment", "alias", "continue-comment-segment:0:-1",
}, {
	"sequence",
	"comment",
	"none",
	"comment-segment",
	"continue-comment",
}, {
	"choice", "wsc", "alias", "ws", "comment",
}, {
	"sequence", "wscs", "alias", "wsc:0:-1",
}, {
	"anything", "anything", "alias",
}, {
	"chars", "any-char", "none", ".",
}, {
	"chars", "open-square", "alias", "[",
}, {
	"chars", "close-square", "alias", "]",
}, {
	"chars", "class-not", "none", "^",
}, {
	"chars", "dash", "alias", "-",
}, {
	"sequence", "optional-class-not", "alias", "class-not:0:1",
}, {
	"class", "not-class-control", "alias", "^\\\\\\[\\]\\^\\-",
}, {
	"chars", "escape", "alias", "\\\\",
}, {
	"sequence", "escaped-char", "alias", "escape", "anything",
}, {
	"choice", "class-char", "none", "not-class-control", "escaped-char",
}, {
	"sequence", "char-range", "none", "class-char", "dash", "class-char",
}, {
	"choice", "char-or-range", "alias", "class-char", "char-range",
}, {
	"sequence", "chars-or-ranges", "alias", "char-or-range:0:-1",
}, {
	"sequence", "char-class", "none", "open-square", "optional-class-not", "chars-or-ranges", "close-square",
}, {
	"chars", "double-quote", "alias", "\\\"",
}, {
	"class", "not-char-sequence-control", "alias", "^\\\\\"",
}, {
	"choice", "sequence-char", "none", "not-char-sequence-control", "escaped-char",
}, {
	"sequence", "char-sequence-chars", "alias", "sequence-char:0:-1",
}, {
	"sequence", "char-sequence", "none", "double-quote", "char-sequence-chars", "double-quote",
}, {
	"choice", "terminal", "alias", "any-char", "char-class", "char-sequence",
}, {
	"class", "symbol-char", "alias", "^\\\\ \\n\\t\\b\\f\\r\\v\\b/.\\[\\]\\\"{}\\^+*?|():=;",
}, {
	"sequence", "symbol-chars", "alias", "symbol-char:1:-1",
}, {
	"sequence", "symbol", "none", "symbol-chars",
}, {
	"chars", "open-paren", "alias", "(",
}, {
	"chars", "close-paren", "alias", ")",
}, {
	"sequence", "group", "alias", "open-paren", "wscs", "expression", "wscs", "close-paren",
}, {
	"chars", "open-brace", "alias", "{",
}, {
	"chars", "close-brace", "alias", "}",
}, {
	"class", "digit", "alias", "0-9",
}, {
	"sequence", "number", "alias", "digit:1:-1",
}, {
	"sequence", "count", "none", "number",
}, {
	"sequence", "count-quantifier", "none", "open-brace", "wscs", "count", "wscs", "close-brace",
}, {
	"sequence", "range-from", "none", "number",
}, {
	"sequence", "range-to", "none", "number",
}, {
	"chars", "comma", "alias", ",",
}, {
	"sequence",
	"range-quantifier",
	"none",
	"open-brace",
	"wscs",
	"range-from",
	"wscs",
	"comma",
	"wscs",
	"range-to",
	"close-brace",
}, {
	"chars", "one-or-more", "none", "+",
}, {
	"chars", "zero-or-more", "none", "*",
}, {
	"chars", "zero-or-one", "none", "?",
}, {
	"choice",
	"quantity",
	"alias",
	"count-quantifier",
	"range-quantifier",
	"one-or-more",
	"zero-or-more",
	"zero-or-one",
}, {
	"choice", "quantifiable", "alias", "terminal", "symbol", "group",
}, {
	"choice", "item-choice", "alias", "terminal", "symbol", "group",
}, {
	"sequence", "item", "none", "item-choice", "quantity:0:1",
}, {
	"sequence", "item-continue", "alias", "wscs", "item",
}, {
	"sequence", "items-continue", "alias", "item-continue:0:-1",
}, {
	"sequence", "sequence", "none", "item", "items-continue",
}, {
	"choice", "element", "alias", "terminal", "symbol", "group", "sequence",
}, {
	"chars", "pipe", "alias", "|",
}, {
	"sequence", "element-continue", "alias", "wscs", "pipe", "wscs", "element",
}, {
	"sequence", "elements-continue", "alias", "element-continue:1:-1",
}, {
	"sequence", "choice", "none", "element", "elements-continue",
}, {
	"choice",
	"expression",
	"alias",
	"terminal",
	"symbol",
	"group",
	"sequence",
	"choice",
}, {
	"chars", "alias", "none", "alias",
}, {
	"chars", "doc", "none", "doc",
}, {
	"chars", "root", "none", "root",
}, {
	"choice", "flag", "alias", "alias", "doc", "root",
}, {
	"chars", "colon", "alias", ":",
}, {
	"sequence", "flag-tag", "alias", "colon", "flag",
}, {
	"sequence", "flags", "alias", "flag-tag:0:-1",
}, {
	"chars", "equal", "alias", "=",
}, {
	"sequence", "definition", "none", "symbol", "flags", "wscs", "equal", "wscs", "expression",
}, {
	"chars", "semicolon", "alias", ";",
}, {
	"choice", "wsc-or-semicolon", "alias", "wsc", "semicolon",
}, {
	"sequence", "wsc-or-semicolons", "alias", "wsc-or-semicolon:0:-1",
}, {
	"sequence",
	"subsequent-definition",
	"alias",
	"wscs",
	"semicolon",
	"wsc-or-semicolons",
	"definition",
}, {
	"sequence",
	"subsequent-definitions",
	"alias",
	"subsequent-definition:0:-1",
}, {
	"sequence",
	"definitions",
	"alias",
	"definition",
	"subsequent-definitions",
}, {
	"sequence",
	"opt-definitions",
	"alias",
	"definitions:0:1",
}, {
	"sequence",
	"syntax",
	"root",
	"wsc-or-semicolons",
	"opt-definitions",
	"wsc-or-semicolons",
}}
