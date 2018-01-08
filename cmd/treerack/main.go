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

	switch os.Args[1] {
	case "check":
		code := check(os.Args[2:])
		exit(code)
	case "generate":
		code := generate(os.Args[2:])
		exit(code)
	case "help", "-help":
		mainHelp()
	default:
		stderr("invalid command")
		stderr()
		stderr(commandsHelp)
		exit(-1)
	}
}
