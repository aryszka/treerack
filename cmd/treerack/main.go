package main

import "os"

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
		stderr()
		stderr(docRef)
		exit(-1)
		return
	}

	var cmd func([]string) int

	switch os.Args[1] {
	case "check-syntax":
		cmd = checkSyntax
	case "generate":
		cmd = generate
	case "check":
		cmd = check
	case "show":
		cmd = show
	case "help", "-help":
		mainHelp()
		return
	default:
		stderr("invalid command")
		stderr()
		stderr(commandsHelp)
		exit(-1)
		return
	}

	exit(cmd(os.Args[2:]))
}
