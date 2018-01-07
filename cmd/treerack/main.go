package main

import (
	"os"
)

func mainHelp() {
	stdout(summary)
	stdout()
	stdout(commandsHelp)
	stdout()
	stdout(docRef)
}

func main() {
	if len(os.Args) == 1 {
		stderr("missing command")
		stderr()
		stderr(commandsHelp)
		stdout()
		stdout(docRef)
		os.Exit(-1)
	}

	switch os.Args[1] {
	case "generate":
		exit := generate(os.Args[2:])
		os.Exit(exit)
	case "help", "-help":
		mainHelp()
	default:
		stderr("invalid command")
		stderr()
		stderr(commandsHelp)
		os.Exit(-1)
	}
}
