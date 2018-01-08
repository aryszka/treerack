package main

import "flag"

type commandOptions struct {
	usage   string
	example string
	args    []string
	flagSet *flag.FlagSet
}

func initOptions(usage, example string, args []string) *commandOptions {
	var o commandOptions

	o.usage = usage
	o.example = example
	o.args = args

	o.flagSet = flag.NewFlagSet("", flag.ContinueOnError)
	o.flagSet.Usage = func() {}
	o.flagSet.SetOutput(werr)

	return &o
}

func (o *commandOptions) flagError() {
	stderr()
	stderr("Options:")
	o.flagSet.PrintDefaults()
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
	stdout(o.example)
	stdout()
	stdout(docRef)
}

func (o *commandOptions) help() bool {
	if len(o.args) == 0 || o.args[0] != "-help" {
		return false
	}

	o.printHelp()
	return true
}
