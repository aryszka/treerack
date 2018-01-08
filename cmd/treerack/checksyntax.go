package main

func checkSyntax(args []string) int {
	options := initOptions(checkSyntaxUsage, checkSyntaxExample, args)
	if options.checkHelp() {
		return 0
	}

	if code := options.parse(); code != 0 {
		return code
	}

	_, code := openSyntax(options)
	return code
}
