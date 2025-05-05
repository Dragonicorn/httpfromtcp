package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestParse testing...")
	}
	var (
		headers Headers
		n       int
		done    bool
		err     error
	)
	// Test: Valid single header
	// headers := NewHeaders()
	headers = make(Headers)
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Valid single header with extra whitespace
	clear(headers)
	data = []byte("       HOST: localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 37, n)
	assert.False(t, done)

	// Test: Valid 2 headers with existing headers
	data = []byte("User-Agent:                 Mozilla/5.0 (X11; Linux x86_64; rv:12.0) Gecko/20100101 Firefox/12.0\r\nMax-Forwards: 10    \r\n")
	n, done, err = headers.Parse(data)
	assert.Equal(t, "Mozilla/5.0 (X11; Linux x86_64; rv:12.0) Gecko/20100101 Firefox/12.0", headers["user-agent"])
	require.NoError(t, err)
	assert.Equal(t, 98, n)
	assert.False(t, done)

	// Test: Valid done with existing headers
	data = []byte("\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.True(t, done)

	// Test: Invalid done with no headers
	clear(headers)
	data = []byte("\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 2, n)
	assert.True(t, done)

	// Test: Invalid spacing header
	clear(headers)
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid character in header
	clear(headers)
	data = []byte("       H@st: localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Missing field name in header
	clear(headers)
	data = []byte("        : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid field-value header
	clear(headers)
	data = []byte("       Host :        \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Valid single header
	clear(headers)
	data = []byte("Set-Person: lane-loves-go;\r\n\r\n")
	n, done, err = headers.Parse(data)
	assert.Equal(t, "lane-loves-go;", headers["set-person"])
	assert.Equal(t, 28, n)
	assert.False(t, done)

	// Test: Valid single header
	data = []byte("Set-Person: prime-loves-zig;\r\n\r\n")
	n, done, err = headers.Parse(data)
	assert.Equal(t, "lane-loves-go;, prime-loves-zig;", headers["set-person"])
	assert.Equal(t, 30, n)
	assert.False(t, done)
}
