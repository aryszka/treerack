package main

import "testing"

func TestCheck(t *testing.T) {
	runMainTest(t,
		mainTest{
			title: "help",
			args: []string{
				"treerack", "check", "-help",
			},
			stdout: []string{
				checkUsage,
				"-syntax",
				"-syntax-string",
				"-input",
				"-input-string",
				checkExample,
				docRef,
			},
		},

		mainTest{
			title: "invalid flag",
			args: []string{
				"treerack", "check", "-foo",
			},
			exit: -1,
			stderr: []string{
				"-syntax",
				"-syntax-string",
				"-input",
				"-input-string",
			},
		},

		mainTest{
			title: "multiple syntaxes",
			args: []string{
				"treerack", "check", "-syntax", "foo.treerack", "-syntax-string", `foo = "bar"`, "-input-string", "bar",
			},
			exit: -1,
			stderr: []string{
				"only one syntax",
				"-syntax",
				"-syntax-string",
				"-input",
				"-input-string",
			},
		},

		mainTest{
			title: "multiple inputs",
			args: []string{
				"treerack", "check", "-syntax-string", `foo = "bar"`, "-input", "foo.txt", "-input-string", "bar",
			},
			exit: -1,
			stderr: []string{
				"only one input",
				"-syntax",
				"-syntax-string",
				"-input",
				"-input-string",
			},
		},

		mainTest{
			title: "multiple inputs, positional",
			args: []string{
				"treerack", "check", "-syntax-string", `foo = "bar"`, "foo.txt", "bar.txt",
			},
			exit: -1,
			stderr: []string{
				"only one input",
				"-syntax",
				"-syntax-string",
				"-input",
				"-input-string",
			},
		},

		mainTest{
			title: "multiple inputs, positional and explicit file",
			args: []string{
				"treerack", "check", "-syntax-string", `foo = "bar"`, "-input", "foo.txt", "bar.txt",
			},
			exit: -1,
			stderr: []string{
				"only one input",
				"-syntax",
				"-syntax-string",
				"-input",
				"-input-string",
			},
		},

		mainTest{
			title: "no syntax",
			args: []string{
				"treerack", "check", "-input-string", "foo",
			},
			exit: -1,
			stderr: []string{
				"missing syntax",
				"-syntax",
				"-syntax-string",
				"-input",
				"-input-string",
			},
		},

		mainTest{
			title: "no input",
			args: []string{
				"treerack", "check", "-syntax-string", `foo = "bar"`,
			},
			exit: -1,
			stderr: []string{
				"missing input",
				"-syntax",
				"-syntax-string",
				"-input",
				"-input-string",
			},
		},

		mainTest{
			title: "invalid syntax",
			args: []string{
				"treerack", "check", "-syntax-string", "foo", "-input-string", "foo",
			},
			exit: -1,
			stderr: []string{
				"parse failed",
			},
		},

		mainTest{
			title: "syntax file open fails",
			args: []string{
				"treerack", "check", "-syntax", "noexist.treerack", "-input-string", "foo",
			},
			exit: -1,
			stderr: []string{
				"file",
			},
		},

		mainTest{
			title: "input file open fails",
			args: []string{
				"treerack", "check", "-syntax-string", `foo = "bar"`, "-input", "noexist.txt",
			},
			exit: -1,
			stderr: []string{
				"file",
			},
		},

		mainTest{
			title: "syntax as file",
			args: []string{
				"treerack", "check", "-syntax", "foo_test.treerack", "-input-string", "bar",
			},
		},

		mainTest{
			title: "syntax as string",
			args: []string{
				"treerack", "check", "-syntax-string", `foo = "bar"`, "-input-string", "bar",
			},
		},

		mainTest{
			title: "input as stdin",
			args: []string{
				"treerack", "check", "-syntax-string", `foo = "bar"`,
			},
			stdin: "bar",
		},

		mainTest{
			title: "input as file",
			args: []string{
				"treerack", "check", "-syntax-string", `foo = "bar"`, "-input", "bar_test.txt",
			},
		},

		mainTest{
			title: "input as string",
			args: []string{
				"treerack", "check", "-syntax-string", `foo = "bar"`, "-input-string", "bar",
			},
		},
	)
}
