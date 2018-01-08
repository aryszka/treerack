package main

import "testing"

var generateFailureTests = convertTests("generate", checkSyntaxFailureTests)

var generateTests = []mainTest{
	{
		title: "failing output",
		args: []string{
			"treerack", "generate", "-syntax-string", `foo = "bar"`,
		},
		failingOutput: true,
		exit:          -1,
	},

	{
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

	{
		title: "syntax as file",
		args: []string{
			"treerack", "generate", "-export", "-package-name", "foo", "-syntax", "foo_test.treerack",
		},
		stdout: []string{
			"package foo",
			"func Parse",
		},
	},

	{
		title: "syntax as string",
		args: []string{
			"treerack", "generate", "-export", "-package-name", "foo", "-syntax-string", `foo = "bar"`,
		},
		stdout: []string{
			"package foo",
			"func Parse",
		},
	},
}

func TestGenerate(t *testing.T) {
	runMainTest(t, mainTest{
		title: "help",
		args: []string{
			"treerack", "generate", "-help",
		},
		stdout: []string{
			generateUsage,
			"-syntax",
			"-syntax-string",
			"-export",
			"-package-name",
			generateExample,
			docRef,
		},
	})

	runMainTest(t, generateFailureTests...)
	runMainTest(t, generateTests...)
}
