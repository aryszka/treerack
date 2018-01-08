package main

import "testing"

func TestGenerate(t *testing.T) {
	runMainTest(t,
		mainTest{
			title: "help",
			args: []string{
				"treerack", "generate", "-help",
			},
			stdout: []string{
				generateUsage,
				"-export",
				"-package-name",
				"-syntax",
				"-syntax-string",
				generateExample,
				docRef,
			},
		},

		mainTest{
			title: "invalid flag",
			args: []string{
				"treerack", "generate", "-foo",
			},
			exit: -1,
			stderr: []string{
				"-export",
				"-package-name",
				"-syntax",
				"-syntax-string",
			},
		},

		mainTest{
			title: "multiple inputs",
			args: []string{
				"treerack", "generate", "-syntax", "foo.treerack", "-syntax-string", `foo = "bar"`,
			},
			exit: -1,
			stderr: []string{
				"only one",
				"-export",
				"-package-name",
				"-syntax",
				"-syntax-string",
			},
		},

		mainTest{
			title: "no input",
			args: []string{
				"treerack", "generate",
			},
			exit: -1,
			stderr: []string{
				"missing syntax input",
				"-export",
				"-package-name",
				"-syntax",
				"-syntax-string",
			},
		},

		mainTest{
			title: "invalid input",
			args: []string{
				"treerack", "generate", "-syntax-string", "foo",
			},
			exit: -1,
			stderr: []string{
				"parse failed",
			},
		},

		mainTest{
			title: "file open fails",
			args: []string{
				"treerack", "generate", "-syntax", "noexist.treerack",
			},
			exit: -1,
			stderr: []string{
				"file",
			},
		},

		mainTest{
			title: "failing output",
			args: []string{
				"treerack", "generate", "-syntax-string", `foo = "bar"`,
			},
			failingOutput: true,
			exit:          -1,
		},

		mainTest{
			title: "syntax as stdin",
			args: []string{
				"treerack", "generate", "-export", "-package-name", "foo",
			},
			stdin: `foo = "bar"`,
			stdout: []string{
				"package foo",
				"func Parse",
			},
		},

		mainTest{
			title: "syntax as file",
			args: []string{
				"treerack", "generate", "-export", "-package-name", "foo", "-syntax", "foo_test.treerack",
			},
			stdout: []string{
				"package foo",
				"func Parse",
			},
		},

		mainTest{
			title: "syntax as string",
			args: []string{
				"treerack", "generate", "-export", "-package-name", "foo", "-syntax-string", `foo = "bar"`,
			},
			stdout: []string{
				"package foo",
				"func Parse",
			},
		},
	)
}
