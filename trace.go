package parse

import (
	"fmt"
	"os"
)

type Trace interface {
	OutN(int, ...interface{})
	Out(...interface{})
	Out1(...interface{})
	Out2(...interface{})
	Out3(...interface{})
	Extend(string) Trace
}

type DefaultTrace struct {
	level int
	path  string
}

type NopTrace struct{}

func NewTrace(level int) *DefaultTrace {
	return &DefaultTrace{
		level: level,
		path:  "/",
	}
}

func (t *DefaultTrace) OutN(level int, a ...interface{}) {
	if level > t.level {
		return
	}

	fmt.Fprintln(os.Stderr, append([]interface{}{t.path}, a...)...)
}

func (t *DefaultTrace) Out(a ...interface{}) {
	t.OutN(0, a...)
}

func (t *DefaultTrace) Out1(a ...interface{}) {
	t.OutN(1, a...)
}

func (t *DefaultTrace) Out2(a ...interface{}) {
	t.OutN(2, a...)
}

func (t *DefaultTrace) Out3(a ...interface{}) {
	t.OutN(3, a...)
}

func (t *DefaultTrace) Extend(name string) Trace {
	var p string
	if t.path == "/" {
		p = t.path + name
	} else {
		p = t.path + "/" + name
	}

	return &DefaultTrace{
		level: t.level,
		path:  p,
	}
}

func (NopTrace) OutN(int, ...interface{}) {}
func (NopTrace) Out(...interface{})       {}
func (NopTrace) Out1(...interface{})      {}
func (NopTrace) Out2(...interface{})      {}
func (NopTrace) Out3(...interface{})      {}
func (t NopTrace) Extend(string) Trace    { return t }
