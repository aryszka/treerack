package main

import (
	"fmt"
	"os"
)

func stderr(a ...interface{}) {
	fmt.Fprintln(os.Stderr, a...)
}

func stdout(a ...interface{}) {
	fmt.Fprintln(os.Stderr, a...)
}
