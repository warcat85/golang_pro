package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("cannot read directory %s: %w", dir, err)
	}

	env := make(Environment)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.Contains(name, "=") {
			// return error for filenames containing '='
			return nil, fmt.Errorf("invalid file name %s: contains '='", name)
		}

		fileName := filepath.Join(dir, name)
		file, err := os.Open(fileName)
		if err != nil {
			// if cannot open, return error
			return nil, fmt.Errorf("cannot open file %s: %w", fileName, err)
		}
		defer file.Close()

		reader := bufio.NewReader(file)
		// read first line (up to newline)
		line, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			// if cannot read the file, return error
			return nil, fmt.Errorf("cannot read file %s: %w", fileName, err)
		}

		if len(line) == 0 {
			env[name] = EnvValue{NeedRemove: true}
			continue
		}

		// trim trailing newline, spaces and tabs
		line = strings.TrimRight(line, " \t\n")
		// replace nulls with newline
		line = strings.ReplaceAll(line, "\x00", "\n")
		env[name] = EnvValue{Value: line}
	}

	return env, nil
}
