package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func buildLine(line string, lines []string) string {
	line += lines[0]
	if len(lines) > 1 {
		fmt.Printf("read: %s\n", line)
		line = buildLine("", lines[1:])
	}
	return line
}

func main() {
	var (
		line string
		fn   string = "messages.txt"
		n    int
	)

	f, err := os.Open(fn)
	if err != nil {
		fmt.Printf("error opening file '%s' - %v\n", fn, err)
		os.Exit(1)
	}
	defer f.Close()

	b := make([]byte, 8)
	for {
		n, err = f.Read(b)
		if n == 0 && err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("error reading data from '%s' - %v\n", fn, err)
			os.Exit(1)
		}
		line = buildLine(line, strings.Split(string(b[:n]), "\n"))
	}
	if len(line) > 0 {
		fmt.Printf("read: %s\n", line)
	}
}
