package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadDir(t *testing.T) {
	testCases := map[string]EnvValue{
		"FOO":                  {Value: "   foo\nwith new line", NeedRemove: false},
		"BAR":                  {Value: "bar", NeedRemove: false},
		"EMPTY":                {Value: "", NeedRemove: false},
		"HELLO":                {Value: "\"hello\"", NeedRemove: false},
		"UNSET":                {NeedRemove: true},
		"WITH_TRAILING_SPACES": {Value: "hello", NeedRemove: false},
		"WITH_EQUALS":          {Value: "EQUALS=value", NeedRemove: false},
	}

	env, err := ReadDir("testdata/env")
	require.NoError(t, err)

	for name, expected := range testCases {
		val, ok := env[name]
		if !ok {
			t.Errorf("Expected variable %s not found in environment", name)
			continue
		}
		if val.Value != expected.Value || val.NeedRemove != expected.NeedRemove {
			t.Errorf("Variable %s value mismatch: got %+v, expected %+v", name, val, expected)
		}
	}

	// Ensure no extra variables
	for name := range env {
		if _, ok := testCases[name]; !ok {
			t.Errorf("Unexpected variable %s in environment", name)
		}
	}
}

func TestReadDirEdgeCases(t *testing.T) {
	t.Run("first line read", func(t *testing.T) {
		tmpDir := createTempDir(t, "MULTI", "first line\nsecond line\nthird line", 0o644)
		env, err := ReadDir(tmpDir)
		require.NoError(t, err)
		testVariable(t, env, "MULTI", "first line")
	})

	t.Run("only newline", func(t *testing.T) {
		tmpDir := createTempDir(t, "NEWLINE", "\n", 0o644)
		env, err := ReadDir(tmpDir)
		require.NoError(t, err)
		testVariable(t, env, "NEWLINE", "")
	})

	t.Run("CRLF", func(t *testing.T) {
		tmpDir := createTempDir(t, "CRLF", "value\r\n", 0o644)
		env, err := ReadDir(tmpDir)
		require.NoError(t, err)
		testVariable(t, env, "CRLF", "value\r")
	})

	t.Run("leading spaces", func(t *testing.T) {
		tmpDir := createTempDir(t, "LEADING", "   value", 0o644)
		env, err := ReadDir(tmpDir)
		require.NoError(t, err)
		testVariable(t, env, "LEADING", "   value")
	})

	t.Run("trailing tabs and spaces", func(t *testing.T) {
		tmpDir := createTempDir(t, "TRAILING", "value \t \n\x00", 0o644)
		env, err := ReadDir(tmpDir)
		require.NoError(t, err)
		testVariable(t, env, "TRAILING", "value")
	})

	t.Run("null bytes", func(t *testing.T) {
		tmpDir := createTempDir(t, "NULLMID", "first\x00second\x00third", 0o644)
		env, err := ReadDir(tmpDir)
		require.NoError(t, err)
		testVariable(t, env, "NULLMID", "first\nsecond\nthird")
	})

	t.Run("cannot read dir", func(t *testing.T) {
		// Test with non-existent directory
		_, err := ReadDir("testdata/fakedir")
		require.Error(t, err, "Expected error for non-existent directory")
	})

	t.Run("ignore subdir", func(t *testing.T) {
		tmpDir := t.TempDir()
		subDir := filepath.Join(tmpDir, "subdir")
		if err := os.Mkdir(subDir, 0o755); err != nil {
			t.Fatalf("Failed to create subdirectory: %v", err)
		}
		path := filepath.Join(subDir, "SOMEVAR")
		if err := os.WriteFile(path, []byte("value"), 0o644); err != nil {
			t.Fatalf("Failed to create file in subdirectory: %v", err)
		}
		env, err := ReadDir(tmpDir)
		require.NoError(t, err)
		require.Empty(t, env, "Expected empty environment, got %v", env)
	})

	t.Run("file with equals", func(t *testing.T) {
		tmpDir := createTempDir(t, "INVALID=NAME", "value", 0o644)
		_, err := ReadDir(tmpDir)
		require.Error(t, err, "Expected error for file name containing '='")
	})

	t.Run("unreadable file", func(t *testing.T) {
		tmpDir := createTempDir(t, "INVALID=NAME", "value", 0o000)
		_, err := ReadDir(tmpDir)
		require.Error(t, err, "Expected error for unreadable file")
	})
}

// creates a temporary directory and a file in it, the name of the file is a variable name
// and the value is a content of that file.
func createTempDir(t *testing.T, name, value string, perm os.FileMode) string {
	t.Helper()

	tempDir := t.TempDir()
	path := filepath.Join(tempDir, name)
	err := os.WriteFile(path, []byte(value), perm)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	return tempDir
}

func testVariable(t *testing.T, env Environment, name, value string) {
	t.Helper()

	val, ok := env[name]
	if !ok {
		t.Fatalf("Variable %s not found", name)
	}
	require.Equal(t, value, val.Value)
	require.False(t, val.NeedRemove, "Variable should not be marked for removal")
	if len(env) > 1 {
		keys := make([]string, 0, len(env)-1)
		for key := range env {
			if key != name {
				keys = append(keys, key)
			}
		}
		t.Errorf("Extra variables found: %v", keys)
	}
}
