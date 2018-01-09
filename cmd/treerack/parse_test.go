package main

import "testing"

var parseFailureTests = convertTests("parse", checkFailureTests)

var parseTests = []mainTest{
	{
		title: "syntax as file",
		args: []string{
			"treerack", "parse", "-syntax", "foo_test.treerack", "-input-string", "bar",
		},
		stdout: []string{
			`"name":"foo"`,
		},
	},

	{
		title: "syntax as string",
		args: []string{
			"treerack", "parse", "-syntax-string", `foo = "bar"`, "-input-string", "bar",
		},
		stdout: []string{
			`"name":"foo"`,
		},
	},

	{
		title: "input as stdin",
		args: []string{
			"treerack", "parse", "-syntax-string", `foo = "bar"`,
		},
		stdin: "bar",
		stdout: []string{
			`"name":"foo"`,
		},
	},

	{
		title: "input as file",
		args: []string{
			"treerack", "parse", "-syntax-string", `foo = "bar"`, "-input", "bar_test.txt",
		},
		stdout: []string{
			`"name":"foo"`,
		},
	},

	{
		title: "input as string",
		args: []string{
			"treerack", "parse", "-syntax-string", `foo = "bar"`, "-input-string", "bar",
		},
		stdout: []string{
			`"name":"foo"`,
		},
	},

	{
		title: "explicit over stdin",
		args: []string{
			"treerack", "parse", "-syntax", "foo_test.treerack", "-input-string", "bar",
		},
		stdin: "invalid",
		stdout: []string{
			`"name":"foo"`,
		},
	},

	{
		title: "pretty",
		args: []string{
			"treerack", "parse", "-syntax-string", `foo = "bar"`, "-input-string", "bar", "-pretty",
		},
		stdout: []string{
			`    "name": "foo"`,
		},
	},

	{
		title: "pretty and indent",
		args: []string{
			"treerack", "parse", "-syntax-string", `foo = "bar"`, "-input-string", "bar", "-pretty", "-indent", "xx",
		},
		stdout: []string{
			`xx"name": "foo"`,
		},
	},

	{
		title: "indent without pretty",
		args: []string{
			"treerack", "parse", "-syntax-string", `foo = "bar"`, "-input-string", "bar", "-pretty", "-indent", "xx",
		},
		stdout: []string{
			`xx"name": "foo"`,
		},
	},

	{
		title: "with child nodes",
		args: []string{
			"treerack", "parse", "-syntax-string", `foo = "bar"; doc = foo`, "-input-string", "bar",
		},
		stdout: []string{
			`"nodes":[`,
			`"text":"bar"`,
		},
	},
}

func TestParse(t *testing.T) {
	runMainTest(t, mainTest{
		title: "help",
		args: []string{
			"treerack", "parse", "-help",
		},
		stdout: []string{
			joinLines(parseUsage),
			"-syntax",
			"-syntax-string",
			"-input",
			"-input-string",
			"-pretty",
			"-indent",
			joinLines(positionalInputUsage),
			joinLines(parseExample),
			joinLines(docRef),
		},
	})

	runMainTest(t, parseFailureTests...)
	runMainTest(t, parseTests...)
}
