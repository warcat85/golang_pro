package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	if len(cmd) == 0 {
		return 1
	}

	// Build environment slice for child process
	baseEnv := os.Environ()
	newEnv := make([]string, 0, len(baseEnv)+len(env))

	// Copy existing environment, excluding variables that need removal
	for _, baseEntry := range baseEnv {
		base := strings.SplitN(baseEntry, "=", 2)
		if len(base) == 1 {
			// malformed entry, keep as is
			newEnv = append(newEnv, baseEntry)
			continue
		}

		name := base[0]
		if envValue, ok := env[name]; ok && envValue.NeedRemove {
			// skip this variable (remove)
			continue
		}
		newEnv = append(newEnv, baseEntry)
	}

	// Add or replace variables from env
	for name, envValue := range env {
		if envValue.NeedRemove {
			// already removed above, skip
			continue
		}

		// construct entry "name=value"
		entry := strings.Join([]string{name, envValue.Value}, "=")
		newEnv = append(newEnv, entry)
	}

	// Create command
	command := exec.Command(cmd[0], cmd[1:]...) //nolint:gosec
	command.Env = newEnv
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	// Run and get exit code
	err := command.Run()
	if err != nil {
		var exitErr *exec.ExitError
		switch {
		case errors.Is(err, exec.ErrNotFound):
			// 127 for not found like the shell
			return 127
		case errors.As(err, &exitErr):
			// error from command
			return exitErr.ExitCode()
		default:
			// 126 like other execution errors in the shell
			return 126
		}
	}
	return 0
}
