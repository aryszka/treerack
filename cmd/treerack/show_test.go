package main

import "testing"

var showFailureTests = convertTests("show", checkFailureTests)

var showTests = []mainTest{
	{
		title: "syntax as file",
		args: []string{
			"treerack", "show", "-syntax", "foo_test.treerack", "-input-string", "bar",
		},
		stdout: []string{
			`"name":"foo"`,
		},
	},

	{
		title: "syntax as string",
		args: []string{
			"treerack", "show", "-syntax-string", `foo = "bar"`, "-input-string", "bar",
		},
		stdout: []string{
			`"name":"foo"`,
		},
	},

	{
		title: "input as stdin",
		args: []string{
			"treerack", "show", "-syntax-string", `foo = "bar"`,
		},
		stdin: "bar",
		stdout: []string{
			`"name":"foo"`,
		},
	},

	{
		title: "input as file",
		args: []string{
			"treerack", "show", "-syntax-string", `foo = "bar"`, "-input", "bar_test.txt",
		},
		stdout: []string{
			`"name":"foo"`,
		},
	},

	{
		title: "input as positional",
		args: []string{
			"treerack", "show", "-syntax-string", `foo = "bar"`, "bar_test.txt",
		},
		stdout: []string{
			`"name":"foo"`,
		},
	},

	{
		title: "input as string",
		args: []string{
			"treerack", "show", "-syntax-string", `foo = "bar"`, "-input-string", "bar",
		},
		stdout: []string{
			`"name":"foo"`,
		},
	},

	{
		title: "explicit over stdin",
		args: []string{
			"treerack", "show", "-syntax", "foo_test.treerack", "-input-string", "bar",
		},
		stdin: "invalid",
		stdout: []string{
			`"name":"foo"`,
		},
	},

	{
		title: "pretty",
		args: []string{
			"treerack", "show", "-syntax-string", `foo = "bar"`, "-input-string", "bar", "-pretty",
		},
		stdout: []string{
			`    "name": "foo"`,
		},
	},

	{
		title: "pretty and indent",
		args: []string{
			"treerack", "show", "-syntax-string", `foo = "bar"`, "-input-string", "bar", "-pretty", "-indent", "xx",
		},
		stdout: []string{
			`xx"name": "foo"`,
		},
	},

	{
		title: "indent without pretty",
		args: []string{
			"treerack", "show", "-syntax-string", `foo = "bar"`, "-input-string", "bar", "-pretty", "-indent", "xx",
		},
		stdout: []string{
			`xx"name": "foo"`,
		},
	},

	{
		title: "with child nodes",
		args: []string{
			"treerack", "show", "-syntax-string", `foo = "bar"; doc = foo`, "-input-string", "bar",
		},
		stdout: []string{
			`"nodes":[`,
			`"text":"bar"`,
		},
	},
}

func TestShow(t *testing.T) {
	runMainTest(t, mainTest{
		title: "help",
		args: []string{
			"treerack", "show", "-help",
		},
		stdout: []string{
			wrapLines(showUsage),
			"-syntax",
			"-syntax-string",
			"-input",
			"-input-string",
			"-pretty",
			"-indent",
			wrapLines(positionalInputUsage),
			wrapLines(showExample),
			wrapLines(docRef),
		},
	})

	runMainTest(t, showFailureTests...)
	runMainTest(t, showTests...)
}
