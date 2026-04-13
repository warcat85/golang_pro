package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCopy(t *testing.T) {
	tests := []struct {
		name         string
		offset       int64
		limit        int64
		expectedFile string
		expectError  bool
		errorType    error
	}{
		{
			name:         "whole file",
			offset:       0,
			limit:        0,
			expectedFile: "out_offset0_limit0.txt",
		},
		{
			name:         "limit 10",
			offset:       0,
			limit:        10,
			expectedFile: "out_offset0_limit10.txt",
		},
		{
			name:         "limit 1000",
			offset:       0,
			limit:        1000,
			expectedFile: "out_offset0_limit1000.txt",
		},
		{
			name:         "offset 100, limit 1000",
			offset:       100,
			limit:        1000,
			expectedFile: "out_offset100_limit1000.txt",
		},
		{
			name:         "offset 6000, limit 1000",
			offset:       6000,
			limit:        1000,
			expectedFile: "out_offset6000_limit1000.txt",
		},
		{
			name:         "offset is file size",
			offset:       6617,
			limit:        0,
			expectedFile: "out_empty.txt",
		},
		{
			name:         "limit exceeds file size",
			offset:       0,
			limit:        10000,
			expectedFile: "out_offset0_limit10000.txt",
		},
		{
			name:        "offset exceeds file size",
			offset:      10000,
			limit:       0,
			expectError: true,
			errorType:   ErrOffsetExceedsFileSize,
		},
		{
			name:        "negative offset",
			offset:      -1,
			limit:       0,
			expectError: true,
			errorType:   ErrOffset,
		},
		{
			name:        "negative limit",
			offset:      0,
			limit:       -1,
			expectError: true,
			errorType:   ErrLimit,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srcPath := filepath.Join("testdata", "input.txt")
			tempFile := createTempFile(t)
			defer os.Remove(tempFile)

			err := Copy(srcPath, tempFile, tc.offset, tc.limit)
			if tc.expectError {
				checkError(t, err, tc.errorType)
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			expectedPath := filepath.Join("testdata", tc.expectedFile)
			compareFiles(t, expectedPath, tempFile)
		})
	}
}

func TestCopyEdgeCases(t *testing.T) {
	t.Run("empty source path", func(t *testing.T) {
		tempFile := createTempFile(t)
		defer os.Remove(tempFile)

		err := Copy("", tempFile, 0, 0)
		checkError(t, err, ErrPathParameters)
	})

	t.Run("empty destination path", func(t *testing.T) {
		srcPath := filepath.Join("testdata", "input.txt")
		err := Copy(srcPath, "", 0, 0)
		checkError(t, err, ErrPathParameters)
	})

	t.Run("empty file", func(t *testing.T) {
		srcPath := filepath.Join("testdata", "input_empty.txt")
		tempFile := createTempFile(t)
		err := Copy(srcPath, tempFile, 0, 0)
		if err != nil {
			t.Errorf("unexpected error copying empty file: %v", err)
		}

		expectedPath := filepath.Join("testdata", "out_empty.txt")
		compareFiles(t, expectedPath, tempFile)
	})

	t.Run("directory", func(t *testing.T) {
		srcPath := filepath.Join("testdata", "input_dir")
		tempFile := createTempFile(t)
		defer os.Remove(tempFile)

		err := Copy(srcPath, tempFile, 0, 0)
		checkError(t, err, ErrUnsupportedFile)
	})

	t.Run("symlink", func(t *testing.T) {
		srcPath := filepath.Join("testdata", "input_link.txt")
		tempFile := createTempFile(t)
		defer os.Remove(tempFile)

		err := Copy(srcPath, tempFile, 0, 0)
		if err != nil {
			t.Errorf("unexpected error copying symlink: %v", err)
		}

		expectedPath := filepath.Join("testdata", "out_offset0_limit0.txt")
		compareFiles(t, expectedPath, tempFile)
	})

	t.Run("/dev/urandom", func(t *testing.T) {
		tempFile := createTempFile(t)
		defer os.Remove(tempFile)

		err := Copy("/dev/urandom", tempFile, 0, 0)
		checkError(t, err, ErrUnsupportedFile)
	})

	t.Run("/dev/null", func(t *testing.T) {
		tempFile := createTempFile(t)
		defer os.Remove(tempFile)

		err := Copy("/dev/null", tempFile, 0, 0)
		checkError(t, err, ErrUnsupportedFile)
	})
}

func createTempFile(t *testing.T) string {
	t.Helper()

	tempFile, err := os.CreateTemp("", "test_*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tempFile.Close()
	return tempFile.Name()
}

func checkError(t *testing.T, err error, expected error) {
	t.Helper()
	if err == nil {
		t.Error("expected error but got none")
		return
	}
	if expected != nil && !errors.Is(err, expected) {
		t.Errorf("expected error %v, got %v", expected, err)
	}
}

func compareFiles(t *testing.T, expectedPath, actualPath string) {
	t.Helper()
	expected, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("failed to read expected file: %v", err)
	}
	actual, err := os.ReadFile(actualPath)
	if err != nil {
		t.Fatalf("failed to read actual file: %v", err)
	}

	if len(expected) != len(actual) {
		t.Errorf("file sizes differ: expected %d, got %d", len(expected), len(actual))
	}

	if !bytesEqual(expected, actual) {
		t.Error("file contents differ")
		// Optionally print first differing byte
		for i := 0; i < len(expected) && i < len(actual); i++ {
			if expected[i] != actual[i] {
				t.Errorf("first difference at byte %d: expected %x, got %x", i, expected[i], actual[i])
				break
			}
		}
	}
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
