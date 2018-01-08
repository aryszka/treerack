package main

import "flag"

type syntaxOptions struct {
	usage      string
	example    string
	args       []string
	positional []string
	syntax     string
	syntaxFile string
	flagSet    *flag.FlagSet
}

func initOptions(usage, example string, args []string) *syntaxOptions {
	var o syntaxOptions
	o.usage = usage
	o.example = example
	o.args = args
	o.flagSet = flag.NewFlagSet("", flag.ContinueOnError)
	o.flagSet.Usage = func() {}
	o.flagSet.SetOutput(werr)
	o.flagSet.StringVar(&o.syntax, "syntax-string", "", syntaxStringUsage)
	o.flagSet.StringVar(&o.syntaxFile, "syntax", "", syntaxFileUsage)
	return &o
}

func flagError(fs *flag.FlagSet) {
	stderr()
	stderr("Options:")
	fs.PrintDefaults()
}

func (o *syntaxOptions) parse() (exit int) {
	if err := o.flagSet.Parse(o.args); err != nil {
		flagError(o.flagSet)
		exit = -1
	}

	o.positional = o.flagSet.Args()
	return
}

func (o *syntaxOptions) help() {
	stdout(o.usage)
	stdout()
	stdout("Options:")
	o.flagSet.SetOutput(wout)
	o.flagSet.PrintDefaults()
	stdout()
	stdout(o.example)
	stdout()
	stdout(docRef)
}

func (o *syntaxOptions) checkHelp() bool {
	if len(o.args) == 0 || o.args[0] != "-help" {
		return false
	}

	o.help()
	return true
}
