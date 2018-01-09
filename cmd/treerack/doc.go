package main

import (
	"strings"
	"unicode/utf8"
)

const summary = `treerack - parser generator - https://github.com/aryszka/treerack`

const commandsHelp = `Available commands:
check          validates arbitrary input against a syntax definition
parse          parses arbitrary input with a syntax definition into an abstract syntax tree
check-syntax   validates a syntax definition
generate       generates a parser from a syntax definition
help           prints the current help

See more details about a particular command by calling:
treerack <command> -help`

const docRef = `See more documentation about the definition syntax and the parser output at
https://github.com/aryszka/treerack.`

const positionalSyntaxUsage = "The path to the syntax file is accepted as a positional argument."

const positionalInputUsage = "The path to the input file is accepted as a positional argument."

const syntaxFileUsage = "path to the syntax file in treerack format"

const syntaxStringUsage = "inline syntax in treerack format"

const inputFileUsage = "path to the input to be parsed"

const inputStringUsage = "inline input string to be parsed"

const packageNameUsage = `package name of the generated Go code`

const exportUsage = `when the export flag is set, the generated code will have exported symbols to allow using
it as a separate package.`

const prettyUsage = `when the pretty flag is set, the AST will be pretty printed`

const indentUsage = `string used for indentation of the printed AST`

const checkUsage = `'treerack check' takes a syntax description from a file or inline string, an arbitrary piece
of text from the standard input, or a file, or inline string, and parses the input text with the defined syntax.
It returns non-zero exit code and prints the problem if the provided syntax is not valid or the input cannot be
parsed with it.`

const checkExample = `Example:
treerack check -syntax example.treerack foo.example`

const parseUsage = `'treerack parse' takes a syntax description from a file or inline string, an arbitrary piece
of text from the standard input, or a file, or inline string, and parses the input text with the defined syntax.
If it was successfully parsed, it prints the resulting abstract syntax tree (AST) in JSON format.`

const parseExample = `Example:
treerack parse -syntax example.treerack foo.example`

const checkSyntaxUsage = `'treerack check-syntax' takes a syntax description from the standard input, or a file,
or inline string, and validates it to check whether it represents a valid syntax. It returns with non-zero exit
code and prints the problem if the syntax is not valid.`

const checkSyntaxExample = `Example:
treerack check-syntax example.treerack`

const generateUsage = `'treerack generate' takes a syntax description from the standard input, or a file, or
inline string, and generates parser code implementing the described syntax. It prints the parser code to the
standard output.`

const generateExample = `Example:
treerack generate example.treerack > parser.go`

const wrap = 72

func wrapLines(s string) string {
	s = strings.Replace(s, "\n", " ", -1)
	w := strings.Split(s, " ")

	var l, ll []string
	for i := 0; i < len(w); i++ {
		ll = append(ll, w[i])
		lineLength := utf8.RuneCount([]byte(strings.Join(ll, " ")))
		if lineLength < wrap {
			continue
		}

		if lineLength > wrap {
			ll = ll[:len(ll)-1]
			i--
		}

		if len(ll) == 0 {
			l = append(l, w[i])
			i++
		} else {
			l = append(l, strings.Join(ll, " "))
			ll = nil
		}
	}

	if len(ll) > 0 {
		l = append(l, strings.Join(ll, " "))
	}

	return strings.Join(l, "\n")
}
