package main

import (
	"encoding/json"

	"github.com/aryszka/treerack"
)

type parseOptions struct {
	command *commandOptions
	syntax  *fileOptions
	input   *fileOptions
	pretty  bool
	indent  string
}

type node struct {
	Name  string  `json:"name"`
	From  int     `json:"from"`
	To    int     `json:"to"`
	Text  string  `json:"text,omitempty"`
	Nodes []*node `json:"nodes,omitempty"`
}

func mapNode(n *treerack.Node) *node {
	var nn node
	nn.Name = n.Name
	nn.From = n.From
	nn.To = n.To

	if len(n.Nodes) == 0 {
		nn.Text = n.Text()
		return &nn
	}

	for i := range n.Nodes {
		nn.Nodes = append(nn.Nodes, mapNode(n.Nodes[i]))
	}

	return &nn
}

func parse(args []string) int {
	var o parseOptions
	o.command = initOptions(parseUsage, parseExample, positionalInputUsage, args)
	o.syntax = &fileOptions{typ: "syntax", flagSet: o.command.flagSet, positionalDoc: positionalInputUsage}
	o.input = &fileOptions{typ: "input", flagSet: o.command.flagSet, positionalDoc: positionalInputUsage}

	o.command.stringFlag(&o.syntax.inline, "syntax-string", syntaxStringUsage)
	o.command.stringFlag(&o.syntax.fileName, "syntax", syntaxFileUsage)

	o.command.stringFlag(&o.input.inline, "input-string", inputStringUsage)
	o.command.stringFlag(&o.input.fileName, "input", inputFileUsage)

	o.command.boolFlag(&o.pretty, "pretty", prettyUsage)
	o.command.stringFlag(&o.indent, "indent", indentUsage)

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

	n, err := s.Parse(input)
	if err != nil {
		stderr(err)
		return -1
	}

	nn := mapNode(n)

	marshal := json.Marshal
	if o.pretty || o.indent != "" {
		if o.indent == "" {
			o.indent = "    "
		}

		marshal = func(n interface{}) ([]byte, error) {
			return json.MarshalIndent(n, "", o.indent)
		}
	}

	b, err := marshal(nn)
	if err != nil {
		stderr(err)
	}

	stdout(string(b))
	return 0
}
