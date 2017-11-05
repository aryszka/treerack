package treerack

import (
	"bufio"
	"bytes"
	"testing"
)

func TestCharBuildNoop(t *testing.T) {
	c := newChar("foo", false, nil, nil)
	c.init(newRegistry())
	b := c.builder()
	ctx := newContext(bufio.NewReader(bytes.NewBuffer(nil)))
	if n, ok := b.build(ctx); len(n) != 0 || ok {
		t.Error("char build not noop")
	}
}
