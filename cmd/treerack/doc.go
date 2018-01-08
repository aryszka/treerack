package main

const summary = `treerack - parser generator - https://github.com/aryszka/treerack`

const commandsHelp = `Available commands:
generate     generates a parser from a syntax definition
help         prints the current help

See more details about a particular command by calling:
treerack <command> -help`

const docRef = "See more documentation about the definition syntax and the parser output at https://github.com/aryszka/treerack."

const syntaxFileUsage = "path to the syntax file in treerack format"

const syntaxStringUsage = "inline syntax in treerack format"

const packageNameUsage = `package name of the generated Go code`

const exportUsage = `when the export flag is set, the generated code will have exported symbols to allow using it as a separate package`

const generateUsage = `treerack generate takes a syntax description from the standard input, or a file, or inline string, and generates parser code implementing the described syntax. It prints the parser code to the standard output.`

const generateExample = `Example:
treerack generate -syntax syntax.treerack > parser.go`

const checkUsage = `treerack check takes a syntax description from the standard input, or a file, or inline string, and validates it to check whether it represents a valid syntax. It returns with non-zero exit code and prints the problem if the syntax is not valid.`

const checkExample = `Example:
treerack check -syntax syntax.treerack`