package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <envdir> <command> [args...]\n", os.Args[0])
		os.Exit(1)
	}

	dir := os.Args[1]
	cmdArgs := os.Args[2:]

	env, err := ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading environment directory: %v\n", err)
		os.Exit(111)
	}

	code := RunCmd(cmdArgs, env)
	os.Exit(code)
}
