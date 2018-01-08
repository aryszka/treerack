package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

type mainTest struct {
	title         string
	args          []string
	failingOutput bool
	exit          int
	stdin         string
	stdout        []string
	stderr        []string
}

type failingWriter struct{}

var errWriteFailed = errors.New("write failed")

func (w failingWriter) Write([]byte) (int, error) {
	return 0, errWriteFailed
}

func init() {
	isTest = true
}

func convertTest(cmd string, t mainTest) mainTest {
	args := make([]string, len(t.args))
	copy(args, t.args)
	args[1] = cmd
	t.args = args
	return t
}

func convertTests(cmd string, t []mainTest) []mainTest {
	tt := make([]mainTest, len(t))
	for i := range t {
		tt[i] = convertTest(cmd, t[i])
	}

	return tt
}

func mockArgs(args ...string) (reset func()) {
	original := os.Args
	os.Args = args
	reset = func() {
		os.Args = original
	}

	return
}

func mockStdin(in string) (reset func()) {
	original := rin

	if in == "" {
		rin = nil
	} else {
		rin = bytes.NewBufferString(in)
	}

	reset = func() {
		rin = original
	}

	return
}

func mockOutput(w *io.Writer, failing bool) (out fmt.Stringer, reset func()) {
	original := *w
	reset = func() { *w = original }

	if failing {
		*w = failingWriter{}
		return
	}

	var buf bytes.Buffer
	*w = &buf
	out = &buf
	return
}

func mockStdout() (out fmt.Stringer, reset func()) {
	return mockOutput(&wout, false)
}

func mockStderr() (out fmt.Stringer, reset func()) {
	return mockOutput(&werr, false)
}

func mockFailingOutput() (reset func()) {
	_, reset = mockOutput(&wout, true)
	return
}

func mockExit() (code *int, reset func()) {
	var exitCode int
	code = &exitCode
	original := exit
	exit = func(c int) { exitCode = c }
	reset = func() { exit = original }
	return
}

func (mt mainTest) run(t *testing.T) {
	test := func(t *testing.T) {
		defer mockArgs(mt.args...)()

		defer mockStdin(mt.stdin)()

		var stdout fmt.Stringer
		if mt.failingOutput {
			defer mockFailingOutput()()
		} else {
			var reset func()
			stdout, reset = mockStdout()
			defer reset()
		}

		stderr, resetStderr := mockStderr()
		defer resetStderr()

		code, resetExit := mockExit()
		defer resetExit()

		main()

		if *code != mt.exit {
			t.Error("invalid exit code")
		}

		if stdout != nil {
			var failed bool
			for i := range mt.stdout {
				if !strings.Contains(stdout.String(), mt.stdout[i]) {
					t.Error("invalid output")
					failed = true
				}
			}

			if failed {
				t.Log(stdout.String())
			}
		}

		var failed bool
		for i := range mt.stderr {
			if !strings.Contains(stderr.String(), mt.stderr[i]) {
				t.Error("invalid error output")
				failed = true
			}
		}

		if failed {
			t.Log(stderr.String())
		}
	}

	if mt.title == "" {
		test(t)
	} else {
		t.Run(mt.title, test)
	}
}

func runMainTest(t *testing.T, mt ...mainTest) {
	for i := range mt {
		mt[i].run(t)
	}
}

func TestMissingCommand(t *testing.T) {
	runMainTest(t,
		mainTest{
			args: []string{"treerack"},
			exit: -1,
			stderr: []string{
				"missing command",
				commandsHelp,
				docRef,
			},
		},
	)
}

func TestInvalidCommand(t *testing.T) {
	runMainTest(t,
		mainTest{
			args: []string{
				"treerack", "foo",
			},
			exit: -1,
			stderr: []string{
				"invalid command",
				commandsHelp,
			},
		},
	)
}

func TestHelp(t *testing.T) {
	runMainTest(t,
		mainTest{
			title: "without dash",
			args: []string{
				"treerack", "help",
			},
			stdout: []string{
				summary, commandsHelp, docRef,
			},
		},
		mainTest{
			title: "with dash",
			args: []string{
				"treerack", "-help",
			},
			stdout: []string{
				summary, commandsHelp, docRef,
			},
		},
	)
}
