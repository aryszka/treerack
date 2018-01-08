package main

type checkOptions struct {
	command *commandOptions
	syntax  *fileOptions
	input   *fileOptions
}

func check(args []string) int {
	var o checkOptions
	o.command = initOptions(checkUsage, checkExample, args)
	o.syntax = &fileOptions{typ: "syntax", flagSet: o.command.flagSet}
	o.input = &fileOptions{typ: "input", flagSet: o.command.flagSet}

	o.command.flagSet.StringVar(&o.syntax.inline, "syntax-string", "", syntaxStringUsage)
	o.command.flagSet.StringVar(&o.syntax.fileName, "syntax", "", syntaxFileUsage)

	o.command.flagSet.StringVar(&o.input.inline, "input-string", "", inputStringUsage)
	o.command.flagSet.StringVar(&o.input.fileName, "input", "", inputFileUsage)

	if o.command.checkHelp() {
		return 0
	}

	if code := o.command.parseArgs(); code != 0 {
		return code
	}

	o.input.positional = o.command.flagSet.Args()
	input, code := o.input.open()
	if code != 0 {
		return code
	}

	s, code := o.syntax.openSyntax()
	if code != 0 {
		return code
	}

	_, err := s.Parse(input)
	if err != nil {
		stderr(err)
		return -1
	}

	return 0
}
