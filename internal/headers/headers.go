package headers

import (
	"fmt"
	// "regexp"
	"strings"
	"unicode"
)

type Headers map[string]string

func (h Headers) Get(key string) string {
	if value, ok := h[strings.ToLower(key)]; ok {
		return value
	}
	return ""
}

// 0-9, A-Z, a-z, !, #, $, %, &, ', *, +, -, ., ^, _, `, |, ~
func ValidateString(s string) bool {
	rt := unicode.RangeTable{
		R16: []unicode.Range16{
			{'!', '!', 1},
			{'#', '\'', 1},
			{'*', '+', 1},
			{'-', '.', 1},
			{'0', '9', 1},
			{'A', 'Z', 1},
			{'^', '`', 1},
			{'a', 'z', 1},
			{'|', '|', 1},
			{'~', '~', 1},
		},
		R32:         []unicode.Range32{},
		LatinOffset: 26*2 + 10 + 15,
	}
	for _, r := range s {
		if !unicode.In(r, &rt) {
			return false
		}
	}
	return true
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	var (
		colon bool
		crlf  bool
		key   string
		line  string
		test  string
		value string
	)
	// fmt.Printf("Data to parse: \"%s\"\n", string(data))
	line, _, crlf = strings.Cut(string(data), "\r\n")
	// return zero bytes consumed if no end-of-line in message
	if !crlf {
		return 0, false, nil
	}
	n = len(line) + 2
	// fmt.Printf("line: \"%s\" (%d)\n", line, n)
	// return end of headers if line starts with CRLF
	if n == 2 {
		if len(h) == 0 {
			return n, true, fmt.Errorf("missing headers in request")
		}
		fmt.Printf("Data consumed: \"<CR><LF>\" (%d bytes)\n\tHeaders: %v\n", n, h)
		return n, true, nil
	}
	// return number of bytes consumed - including CRLF
	key, value, colon = strings.Cut(line, ":")
	// return error if no colon separator in header
	if !colon {
		return 0, false, fmt.Errorf("malformed header - %s", line)
	}
	// check for illegal whitespace between field-name and ':'
	test = strings.TrimSpace(key)
	// fmt.Printf("key: \"%s\" test: \"%s\"\n", key, test)
	if len(test) == 0 {
		return 0, false, fmt.Errorf("malformed header (missing field-name) - %s", line)
	}
	if key[len(key)-1] != test[len(test)-1] {
		return 0, false, fmt.Errorf("malformed header (illegal whitespace after field-name) - %s", line)
	}
	// fmt.Printf("key: \"%s\" value: \"%s\"\n", key, value)
	key = strings.ToLower(strings.TrimSpace(key))
	// check for illegal character in field-name
	if !ValidateString(key) {
		return 0, false, fmt.Errorf("malformed header (illegal characters in field-name) - %s", line)
	}
	value = strings.TrimSpace(value)
	// check for missing value
	if len(value) == 0 {
		return 0, false, fmt.Errorf("malformed header (missing field-value) - %s", line)
	}
	// fmt.Printf("key: \"%s\" value: \"%s\"\n", key, value)

	// append value if key already exists
	if test, ok := h[key]; ok {
		value = fmt.Sprintf("%s, %s", test, value)
	}
	// fill header map with valid data
	h[key] = value
	fmt.Printf("Data consumed: \"%s<CR><LF>\" (%d bytes)\n\tHeaders: %v\n", line, n, h)
	return n, false, nil
}
