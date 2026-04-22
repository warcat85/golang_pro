package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

var (
	from, to      string
	limit, offset int64
)

func init() {
	flag.StringVar(&from, "from", "", "file to read from")
	flag.StringVar(&to, "to", "", "file to write to")
	flag.Int64Var(&limit, "limit", 0, "limit of bytes to copy")
	flag.Int64Var(&offset, "offset", 0, "offset in input file")
}

func main() {
	flag.Parse()

	err := Copy(from, to, offset, limit)
	if err != nil {
		switch {
		case errors.Is(err, ErrPathParameters):
			fmt.Println("Error: -from and -to flags are required")
			flag.Usage()
		case errors.Is(err, ErrOffset):
			fmt.Println("Error: offset cannot be negative")
		case errors.Is(err, ErrLimit):
			fmt.Println("Error: limit cannot be negative")
		default:
			fmt.Printf("Error: %v\n", err)
		}
		os.Exit(1)
	}

	fmt.Printf("File [%s] copied successfully!\n", from)
}
