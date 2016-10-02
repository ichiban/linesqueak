package linesqueak_test

import (
	"testing"
	"bytes"
	"github.com/ichiban/linesqueak"
	"bufio"
	"io"
)

func TestEditor_Line(t *testing.T) {
	in := bytes.NewBuffer([]byte("foo bar\x0D"))
	out := &checkedWriter{
		t: t,
		expectations: []string{
			"\r> \x1b[0K\r\x1b[2C",
			"\r> f\x1b[0K\r\x1b[3C",
			"\r> fo\x1b[0K\r\x1b[4C",
			"\r> foo\x1b[0K\r\x1b[5C",
			"\r> foo \x1b[0K\r\x1b[6C",
			"\r> foo b\x1b[0K\r\x1b[7C",
			"\r> foo ba\x1b[0K\r\x1b[8C",
			"\r> foo bar\x1b[0K\r\x1b[9C",
		},
	}

	e := &linesqueak.Editor{
		In: bufio.NewReader(in),
		Out: bufio.NewWriter(out),
		Prompt: "> ",
	}

	l, err := e.Line()
	if err != nil {
		t.Error(err)
	}
	if l != "foo bar" {
		t.Errorf(`expected "foo bar" got %#v`, l)
	}
}

type checkedWriter struct {
	t *testing.T
	expectations []string
	pos int
}

var _ io.Writer = (*checkedWriter)(nil)

func (c *checkedWriter) Write(p []byte) (int, error) {
	t := c.t
	e := c.expectations[c.pos]
	a := string(p)

	if e != a {
		t.Errorf(`expected %#v got %#v`, e, a)
	}

	c.pos++
	return len(p), nil
}
