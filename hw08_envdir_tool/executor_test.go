package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunCmd(t *testing.T) {
	// Create environment manually
	env := Environment{
		"FOO":   EnvValue{Value: "123", NeedRemove: false},
		"BAR":   EnvValue{Value: "value", NeedRemove: false},
		"UNSET": EnvValue{NeedRemove: true},
		"ADDED": EnvValue{Value: "added=true", NeedRemove: false},
		"EMPTY": EnvValue{Value: "", NeedRemove: false},
	}
	vars := []struct {
		name  string
		value string
	}{
		{"HELLO", "hello"},
		{"FOO", "456"},
		{"BAR", "not value"},
		{"EMPTY", "123"},
		{"UNSET", "first and second"},
		{"NULL", "first and second"},
	}

	for _, variable := range vars {
		os.Setenv(variable.name, variable.value)
		defer os.Unsetenv(variable.name)
	}

	scriptPath := filepath.Join("testdata", "echo.sh")
	cmd := []string{"bash", scriptPath, "arg1", "arg2"}

	// Capture output
	outputFile := filepath.Join(t.TempDir(), "output.txt")
	out, err := os.Create(outputFile)
	if err != nil {
		t.Fatalf("Failed to create output file: %v", err)
	}
	defer out.Close()

	// Temporarily redirect stdout and stderr to output.txt
	stdout, stderr := overrideOutput(t, out, out)
	defer overrideOutput(t, stdout, stderr)

	code := RunCmd(cmd, env)

	out.Close()

	output, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	// Verify environment variables are set correctly
	// The script prints HELLO, BAR, FOO, UNSET, ADDED, EMPTY
	// HELLO stays the same
	// We set FOO=123, BAR=value, EMPTY should be set to empty
	// UNSET should be removed
	// ADDED should be added
	expected := strings.Join([]string{
		"HELLO is (hello)",
		"BAR is (value)",
		"FOO is (123)",
		"UNSET is ()",
		"ADDED is (added=true)",
		"EMPTY is ()",
		"arguments are arg1 arg2",
	}, "\n") + "\n"

	require.Equal(t, 0, code)
	require.Equal(t, expected, string(output))
}

func TestRunCmdRemovesVariable(t *testing.T) {
	// Set a variable in the parent environment
	os.Setenv("TO_BE_REMOVED", "somevalue")
	defer os.Unsetenv("TO_BE_REMOVED")

	env := Environment{
		"TO_BE_REMOVED": EnvValue{NeedRemove: true},
	}

	outputFile := filepath.Join(t.TempDir(), "output.txt")
	out, err := os.Create(outputFile)
	if err != nil {
		t.Fatalf("Failed to create output file: %v", err)
	}
	defer out.Close()

	// Temporarily redirect stdout and stderr to output.txt
	stdout, stderr := overrideOutput(t, out, out)
	defer overrideOutput(t, stdout, stderr)

	// Run bash command that prints if variable is set
	cmd := []string{"bash", "-c", "echo \"${TO_BE_REMOVED-unset}\""}
	code := RunCmd(cmd, env)

	out.Close()

	output, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	require.Equal(t, 0, code)
	// Verify that TO_BE_REMOVED printed "unset"
	require.Equal(t, "unset\n", string(output), "TO_BE_REMOVED is not unset")
}

func TestRunCmdErrors(t *testing.T) {
	tests := []struct {
		name     string
		cmd      []string
		env      Environment
		expected int
	}{
		{
			name:     "empty command",
			cmd:      []string{},
			env:      Environment{},
			expected: 1,
		},
		{
			name:     "exit code",
			cmd:      []string{"bash", "-c", "exit 42"},
			env:      Environment{},
			expected: 42,
		},
		{
			name:     "permission denied",
			cmd:      []string{"/dev/null"},
			env:      Environment{},
			expected: 126,
		},
		{
			name:     "command not found",
			cmd:      []string{"nonexistentcommand"},
			env:      Environment{},
			expected: 127,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			code := RunCmd(testCase.cmd, testCase.env)
			require.Equal(t, testCase.expected, code)
		})
	}
}

func overrideOutput(t *testing.T, stdout, stderr *os.File) (*os.File, *os.File) {
	t.Helper()

	prevStdout := os.Stdout
	os.Stdout = stdout

	prevStderr := os.Stderr
	os.Stderr = stderr
	return prevStdout, prevStderr
}
