package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/cheggaaa/pb/v3"
)

var (
	ErrPathParameters        = errors.New("invalid path specified")
	ErrLimit                 = errors.New("offset is negative")
	ErrOffset                = errors.New("limit is negative")
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	if fromPath == "" || toPath == "" {
		return fmt.Errorf("source or destination is not specified: %w", ErrPathParameters)
	}

	if offset < 0 {
		return fmt.Errorf("offset is negative: %w", ErrOffset)
	}

	if limit < 0 {
		return fmt.Errorf("limit is negative: %w", ErrLimit)
	}

	inFile, err := os.Open(fromPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inFile.Close()

	// get the file info to check size and mode
	info, err := inFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// check that the file is a regular file
	mode := info.Mode()
	if !mode.IsRegular() {
		return ErrUnsupportedFile
	}

	// check the file size
	size := info.Size()
	if offset > size {
		return ErrOffsetExceedsFileSize
	}

	if offset > 0 {
		_, err = inFile.Seek(offset, io.SeekStart)
		if err != nil {
			return fmt.Errorf("failed to seek: %w", err)
		}
	}

	outFile, err := os.Create(toPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	remainingSize := size - offset
	// how many bytes to copy
	var bytesToCopy int64
	if limit == 0 || limit > remainingSize {
		bytesToCopy = remainingSize
	} else {
		bytesToCopy = limit
	}

	bar := pb.StartNew(int(bytesToCopy))
	bar.SetTemplateString(`{{counters . }} {{bar . }} {{percent . }} {{speed . }}`)
	reader := bar.NewProxyReader(inFile)

	var bytesCopied int64
	if limit == 0 {
		// Copy all remaining bytes
		bytesCopied, err = io.Copy(outFile, reader)
	} else {
		bytesCopied, err = io.CopyN(outFile, reader, bytesToCopy)
	}
	bar.Finish()

	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("copy failed: %w", err)
	}

	if err == nil && bytesCopied != bytesToCopy {
		return fmt.Errorf("incomplete copy: copied %d bytes, expected %d", bytesCopied, bytesToCopy)
	}

	return nil
}
