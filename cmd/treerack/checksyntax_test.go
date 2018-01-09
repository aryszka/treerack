package main

import "testing"

var checkSyntaxFailureTests = []mainTest{
	{
		title: "invalid flag",
		args: []string{
			"treerack", "check-syntax", "-foo",
		},
		exit: -1,
		stderr: []string{
			"-syntax",
			"-syntax-string",
			joinLines(positionalSyntaxUsage),
		},
	},

	{
		title: "multiple inputs",
		args: []string{
			"treerack", "check-syntax", "-syntax", "foo.treerack", "-syntax-string", `foo = "bar"`,
		},
		exit: -1,
		stderr: []string{
			"only one syntax",
			"-syntax",
			"-syntax-string",
			joinLines(positionalSyntaxUsage),
		},
	},

	{
		title: "multiple inputs, positional",
		args: []string{
			"treerack", "check-syntax", "foo.treerack", "bar.treerack",
		},
		exit: -1,
		stderr: []string{
			"only one syntax",
			"-syntax",
			"-syntax-string",
			joinLines(positionalSyntaxUsage),
		},
	},

	{
		title: "multiple inputs, positional and explicit file",
		args: []string{
			"treerack", "check-syntax", "-syntax", "foo.treerack", "bar.treerack",
		},
		exit: -1,
		stderr: []string{
			"only one syntax",
			"-syntax",
			"-syntax-string",
			joinLines(positionalSyntaxUsage),
		},
	},

	{
		title: "no input",
		args: []string{
			"treerack", "check-syntax",
		},
		exit: -1,
		stderr: []string{
			"missing syntax",
			"-syntax",
			"-syntax-string",
			joinLines(positionalSyntaxUsage),
		},
	},

	{
		title: "invalid input",
		args: []string{
			"treerack", "check-syntax", "-syntax-string", "foo",
		},
		exit: -1,
		stderr: []string{
			"parse failed",
		},
	},

	{
		title: "file open fails",
		args: []string{
			"treerack", "check-syntax", "-syntax", "noexist.treerack",
		},
		exit: -1,
		stderr: []string{
			"file",
		},
	},
}

var checkSyntaxTests = []mainTest{
	{
		title: "syntax as stdin",
		args: []string{
			"treerack", "check-syntax",
		},
		stdin: `foo = "bar"`,
	},

	{
		title: "syntax as file",
		args: []string{
			"treerack", "check-syntax", "-syntax", "foo_test.treerack",
		},
	},

	{
		title: "syntax as string",
		args: []string{
			"treerack", "check-syntax", "-syntax-string", `foo = "bar"`,
		},
	},

	{
		title: "explicit over stdin",
		args: []string{
			"treerack", "check-syntax", "-syntax", "foo_test.treerack",
		},
		stdin: "invalid",
	},
}

func TestCheckSyntax(t *testing.T) {
	runMainTest(t, mainTest{
		title: "help",
		args: []string{
			"treerack", "check-syntax", "-help",
		},
		stdout: []string{
			joinLines(checkSyntaxUsage),
			"-syntax",
			"-syntax-string",
			joinLines(positionalSyntaxUsage),
			joinLines(checkSyntaxExample),
			joinLines(docRef),
		},
	})

	runMainTest(t, checkSyntaxFailureTests...)
	runMainTest(t, checkSyntaxTests...)
}
