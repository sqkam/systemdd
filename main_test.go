package main

import (
	"fmt"
	"path/filepath"
	"slices"
	"testing"
)

func TestSplit(t *testing.T) {

	cmdRaw1 := "       ls         -al"
	cmd1, args1 := splitCmd(cmdRaw1)
	abs, _ := filepath.Abs(".")
	fmt.Printf("%v\n", abs)

	Assert(t, cmd1 == "ls")
	Assert(t, slices.Equal(args1, []string{"-al"}))

}

// Assert asserts cond is true, otherwise fails the test.
func Assert(t *testing.T, cond bool) {
	t.Helper()
	if !cond {
		t.Fatal("assertion failed")
	}
}
