package main

type checkOptions struct {
	command *commandOptions
	syntax  *fileOptions
	input   *fileOptions
}

func check(args []string) int {
	var o checkOptions
	o.command = initOptions(checkUsage, checkExample, positionalInputUsage, args)
	o.syntax = &fileOptions{typ: "syntax", flagSet: o.command.flagSet, positionalDoc: positionalInputUsage}
	o.input = &fileOptions{typ: "input", flagSet: o.command.flagSet, positionalDoc: positionalInputUsage}

	o.command.stringFlag(&o.syntax.inline, "syntax-string", syntaxStringUsage)
	o.command.stringFlag(&o.syntax.fileName, "syntax", syntaxFileUsage)

	o.command.stringFlag(&o.input.inline, "input-string", inputStringUsage)
	o.command.stringFlag(&o.input.fileName, "input", inputFileUsage)

	if o.command.help() {
		return 0
	}

	if code := o.command.parseArgs(); code != 0 {
		return code
	}

	s, code := o.syntax.openSyntax()
	if code != 0 {
		return code
	}

	o.input.positional = o.command.flagSet.Args()
	input, code := o.input.open()
	if code != 0 {
		return code
	}

	defer input.Close()

	_, err := s.Parse(input)
	if err != nil {
		stderr(err)
		return -1
	}

	return 0
}
