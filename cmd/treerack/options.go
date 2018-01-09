package main

import "flag"

type commandOptions struct {
	usage         string
	example       string
	args          []string
	flagSet       *flag.FlagSet
	positionalDoc string
}

func initOptions(usage, example, positionalDoc string, args []string) *commandOptions {
	var o commandOptions

	o.usage = wrapLines(usage)
	o.example = wrapLines(example)
	o.positionalDoc = wrapLines(positionalDoc)
	o.args = args

	o.flagSet = flag.NewFlagSet("", flag.ContinueOnError)
	o.flagSet.Usage = func() {}
	o.flagSet.SetOutput(werr)

	return &o
}

func (o *commandOptions) boolFlag(v *bool, name, usage string) {
	usage = wrapLines(usage)
	o.flagSet.BoolVar(v, name, *v, usage)
}

func (o *commandOptions) stringFlag(v *string, name, usage string) {
	usage = wrapLines(usage)
	o.flagSet.StringVar(v, name, *v, usage)
}

func (o *commandOptions) flagError() {
	stderr()
	stderr("Options:")
	o.flagSet.PrintDefaults()
	stderr()
	stderr(o.positionalDoc)
}

func (o *commandOptions) parseArgs() (exit int) {
	if err := o.flagSet.Parse(o.args); err != nil {
		o.flagError()
		exit = -1
	}

	return
}

func (o *commandOptions) printHelp() {
	stdout(o.usage)
	stdout()

	stdout("Options:")
	o.flagSet.SetOutput(wout)
	o.flagSet.PrintDefaults()
	stdout()
	stdout(o.positionalDoc)

	stdout()
	stdout(o.example)
	stdout()
	stdout(wrapLines(docRef))
}

func (o *commandOptions) help() bool {
	if len(o.args) == 0 || o.args[0] != "-help" {
		return false
	}

	o.printHelp()
	return true
}
