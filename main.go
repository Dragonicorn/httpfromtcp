package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func buildLine(ch chan string, line string, lines []string) string {
	line += lines[0]
	if len(lines) > 1 {
		// fmt.Printf("sending line '%s' to channel...\n", line)
		ch <- line
		line = buildLine(ch, "", lines[1:])
	}
	return line
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	ch := make(chan string)
	go func() {
		var (
			err  error
			line string
			n    int
		)
		b := make([]byte, 8)
		for {
			n, err = f.Read(b)
			if n == 0 && err == io.EOF {
				break
			}
			if err != nil {
				fmt.Printf("error reading data from file - %v\n", err)
				close(ch)
				f.Close()
				os.Exit(1)
			}
			line = buildLine(ch, line, strings.Split(string(b[:n]), "\n"))
		}
		if len(line) > 0 {
			// fmt.Printf("sending last line '%s' to channel...\n", line)
			ch <- line
			close(ch)
			f.Close()
		}
	}()
	return ch
}

func main() {
	const fn string = "messages.txt"

	f, err := os.Open(fn)
	if err != nil {
		fmt.Printf("error opening file '%s' - %v\n", f.Name(), err)
		os.Exit(1)
	}
	for line := range getLinesChannel(f) {
		fmt.Printf("read: %s\n", line)
	}
}
