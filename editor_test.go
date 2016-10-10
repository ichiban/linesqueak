package linesqueak_test

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/ichiban/linesqueak"
)

func TestEditor_LineEnter(t *testing.T) {
	in := bytes.NewBuffer([]byte("foo bar\x0d"))
	out := &checkedWriter{
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
		In:     bufio.NewReader(in),
		Out:    bufio.NewWriter(out),
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

func TestEditor_LineCtrlC(t *testing.T) {
	in := bytes.NewBuffer([]byte("foo b\x03"))
	out := &checkedWriter{
		expectations: []string{
			"\r> \x1b[0K\r\x1b[2C",
			"\r> f\x1b[0K\r\x1b[3C",
			"\r> fo\x1b[0K\r\x1b[4C",
			"\r> foo\x1b[0K\r\x1b[5C",
			"\r> foo \x1b[0K\r\x1b[6C",
			"\r> foo b\x1b[0K\r\x1b[7C",
		},
	}

	e := &linesqueak.Editor{
		In:     bufio.NewReader(in),
		Out:    bufio.NewWriter(out),
		Prompt: "> ",
	}

	l, err := e.Line()
	if err == nil {
		t.Error("err expected")
	}
	if l != "foo b" {
		t.Errorf(`expected "foo b" got %#v`, l)
	}
}

func TestEditor_LineBackspace(t *testing.T) {
	in := bytes.NewBuffer([]byte("fooo\x7f bar\x0d"))
	out := &checkedWriter{
		expectations: []string{
			"\r> \x1b[0K\r\x1b[2C",
			"\r> f\x1b[0K\r\x1b[3C",
			"\r> fo\x1b[0K\r\x1b[4C",
			"\r> foo\x1b[0K\r\x1b[5C",
			"\r> fooo\x1b[0K\r\x1b[6C",
			"\r> foo\x1b[0K\r\x1b[5C",
			"\r> foo \x1b[0K\r\x1b[6C",
			"\r> foo b\x1b[0K\r\x1b[7C",
			"\r> foo ba\x1b[0K\r\x1b[8C",
			"\r> foo bar\x1b[0K\r\x1b[9C",
		},
	}

	e := &linesqueak.Editor{
		In:     bufio.NewReader(in),
		Out:    bufio.NewWriter(out),
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

func TestEditor_LineCtrlH(t *testing.T) {
	in := bytes.NewBuffer([]byte("fooo\x08 bar\x0d"))
	out := &checkedWriter{
		expectations: []string{
			"\r> \x1b[0K\r\x1b[2C",
			"\r> f\x1b[0K\r\x1b[3C",
			"\r> fo\x1b[0K\r\x1b[4C",
			"\r> foo\x1b[0K\r\x1b[5C",
			"\r> fooo\x1b[0K\r\x1b[6C",
			"\r> foo\x1b[0K\r\x1b[5C",
			"\r> foo \x1b[0K\r\x1b[6C",
			"\r> foo b\x1b[0K\r\x1b[7C",
			"\r> foo ba\x1b[0K\r\x1b[8C",
			"\r> foo bar\x1b[0K\r\x1b[9C",
		},
	}

	e := &linesqueak.Editor{
		In:     bufio.NewReader(in),
		Out:    bufio.NewWriter(out),
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

func TestEditor_LineCtrlD(t *testing.T) {
	{ // with empty buffer
		in := bytes.NewBuffer([]byte("\x04"))
		out := &checkedWriter{
			expectations: []string{
				"\r> \x1b[0K\r\x1b[2C",
			},
		}

		e := &linesqueak.Editor{
			In:     bufio.NewReader(in),
			Out:    bufio.NewWriter(out),
			Prompt: "> ",
		}

		l, err := e.Line()
		if err != io.EOF {
			t.Error(err)
		}
		if l != "" {
			t.Errorf(`expected "" got %#v`, l)
		}
	}

	{ // with non-empty buffer
		in := bytes.NewBuffer([]byte("fooo\x02\x04 bar\x0d"))
		out := &checkedWriter{
			expectations: []string{
				"\r> \x1b[0K\r\x1b[2C",
				"\r> f\x1b[0K\r\x1b[3C",
				"\r> fo\x1b[0K\r\x1b[4C",
				"\r> foo\x1b[0K\r\x1b[5C",
				"\r> fooo\x1b[0K\r\x1b[6C",
				"\r> fooo\x1b[0K\r\x1b[5C",
				"\r> foo\x1b[0K\r\x1b[5C",
				"\r> foo \x1b[0K\r\x1b[6C",
				"\r> foo b\x1b[0K\r\x1b[7C",
				"\r> foo ba\x1b[0K\r\x1b[8C",
				"\r> foo bar\x1b[0K\r\x1b[9C",
			},
		}

		e := &linesqueak.Editor{
			In:     bufio.NewReader(in),
			Out:    bufio.NewWriter(out),
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
}

func TestEditor_LineCtrlT(t *testing.T) {
	in := bytes.NewBuffer([]byte("fo obra\x14\x02\x02\x02\x02\x14\x0d"))
	out := &checkedWriter{
		expectations: []string{
			"\r> \x1b[0K\r\x1b[2C",
			"\r> f\x1b[0K\r\x1b[3C",
			"\r> fo\x1b[0K\r\x1b[4C",
			"\r> fo \x1b[0K\r\x1b[5C",
			"\r> fo o\x1b[0K\r\x1b[6C",
			"\r> fo ob\x1b[0K\r\x1b[7C",
			"\r> fo obr\x1b[0K\r\x1b[8C",
			"\r> fo obra\x1b[0K\r\x1b[9C",
			"\r> fo obar\x1b[0K\r\x1b[9C",
			"\r> fo obar\x1b[0K\r\x1b[8C",
			"\r> fo obar\x1b[0K\r\x1b[7C",
			"\r> fo obar\x1b[0K\r\x1b[6C",
			"\r> fo obar\x1b[0K\r\x1b[5C",
			"\r> foo bar\x1b[0K\r\x1b[6C",
		},
	}

	e := &linesqueak.Editor{
		In:     bufio.NewReader(in),
		Out:    bufio.NewWriter(out),
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

func TestEditor_LineCtrlBCtrlF(t *testing.T) {
	in := bytes.NewBuffer([]byte("foo bar\x02\x02\x02\x02\x02\x02\x02\x02\x06\x06\x06\x06\x06\x06\x06\x0d"))
	out := &checkedWriter{
		expectations: []string{
			"\r> \x1b[0K\r\x1b[2C",
			"\r> f\x1b[0K\r\x1b[3C",
			"\r> fo\x1b[0K\r\x1b[4C",
			"\r> foo\x1b[0K\r\x1b[5C",
			"\r> foo \x1b[0K\r\x1b[6C",
			"\r> foo b\x1b[0K\r\x1b[7C",
			"\r> foo ba\x1b[0K\r\x1b[8C",
			"\r> foo bar\x1b[0K\r\x1b[9C",
			"\r> foo bar\x1b[0K\r\x1b[8C",
			"\r> foo bar\x1b[0K\r\x1b[7C",
			"\r> foo bar\x1b[0K\r\x1b[6C",
			"\r> foo bar\x1b[0K\r\x1b[5C",
			"\r> foo bar\x1b[0K\r\x1b[4C",
			"\r> foo bar\x1b[0K\r\x1b[3C",
			"\r> foo bar\x1b[0K\r\x1b[2C",
			"\a",
			"\r> foo bar\x1b[0K\r\x1b[3C",
			"\r> foo bar\x1b[0K\r\x1b[4C",
			"\r> foo bar\x1b[0K\r\x1b[5C",
			"\r> foo bar\x1b[0K\r\x1b[6C",
			"\r> foo bar\x1b[0K\r\x1b[7C",
			"\r> foo bar\x1b[0K\r\x1b[8C",
			"\r> foo bar\x1b[0K\r\x1b[9C",
			"\a",
		},
	}

	e := &linesqueak.Editor{
		In:     bufio.NewReader(in),
		Out:    bufio.NewWriter(out),
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

func TestEditor_LineCtrlPCtrlN(t *testing.T) {
	in := bytes.NewBuffer([]byte("\x10foo\x0d\x10\x10\x0e\x0ebar\x0d"))
	out := &checkedWriter{
		expectations: []string{
			"\r> \x1b[0K\r\x1b[2C",
			"\r> \x1b[0K\r\x1b[2C",
			"\r> f\x1b[0K\r\x1b[3C",
			"\r> fo\x1b[0K\r\x1b[4C",
			"\r> foo\x1b[0K\r\x1b[5C",
			"\r> \x1b[0K\r\x1b[2C",
			"\r> foo\x1b[0K\r\x1b[5C",
			"\r> \x1b[0K\r\x1b[2C",
			"\r> foo\x1b[0K\r\x1b[5C",
			"\r> \x1b[0K\r\x1b[2C",
			"\r> b\x1b[0K\r\x1b[3C",
			"\r> ba\x1b[0K\r\x1b[4C",
			"\r> bar\x1b[0K\r\x1b[5C",
		},
	}

	e := &linesqueak.Editor{
		In:     bufio.NewReader(in),
		Out:    bufio.NewWriter(out),
		Prompt: "> ",
	}

	l, err := e.Line()
	if err != nil {
		t.Error(err)
	}
	if l != "foo" {
		t.Errorf(`expected "foo" got %#v`, l)
	}

	e.HistoryAdd("foo")

	l, err = e.Line()
	if err != nil {
		t.Error(err)
	}
	if l != "bar" {
		t.Errorf(`expected "bar" got %#v`, l)
	}
}

func TestEditor_LineCtrlU(t *testing.T) {
	in := bytes.NewBuffer([]byte("foo bar\x15\x0d"))
	out := &checkedWriter{
		expectations: []string{
			"\r> \x1b[0K\r\x1b[2C",
			"\r> f\x1b[0K\r\x1b[3C",
			"\r> fo\x1b[0K\r\x1b[4C",
			"\r> foo\x1b[0K\r\x1b[5C",
			"\r> foo \x1b[0K\r\x1b[6C",
			"\r> foo b\x1b[0K\r\x1b[7C",
			"\r> foo ba\x1b[0K\r\x1b[8C",
			"\r> foo bar\x1b[0K\r\x1b[9C",
			"\r> \x1b[0K\r\x1b[2C",
		},
	}

	e := &linesqueak.Editor{
		In:     bufio.NewReader(in),
		Out:    bufio.NewWriter(out),
		Prompt: "> ",
	}

	l, err := e.Line()
	if err != nil {
		t.Error(err)
	}
	if l != "" {
		t.Errorf(`expected "" got %#v`, l)
	}
}

func TestEditor_LineCtrlK(t *testing.T) {
	in := bytes.NewBuffer([]byte("foo bar\x02\x02\x0b\x0d"))
	out := &checkedWriter{
		expectations: []string{
			"\r> \x1b[0K\r\x1b[2C",
			"\r> f\x1b[0K\r\x1b[3C",
			"\r> fo\x1b[0K\r\x1b[4C",
			"\r> foo\x1b[0K\r\x1b[5C",
			"\r> foo \x1b[0K\r\x1b[6C",
			"\r> foo b\x1b[0K\r\x1b[7C",
			"\r> foo ba\x1b[0K\r\x1b[8C",
			"\r> foo bar\x1b[0K\r\x1b[9C",
			"\r> foo bar\x1b[0K\r\x1b[8C",
			"\r> foo bar\x1b[0K\r\x1b[7C",
			"\r> foo b\x1b[0K\r\x1b[7C",
		},
	}

	e := &linesqueak.Editor{
		In:     bufio.NewReader(in),
		Out:    bufio.NewWriter(out),
		Prompt: "> ",
	}

	l, err := e.Line()
	if err != nil {
		t.Error(err)
	}
	if l != "foo b" {
		t.Errorf(`expected "foo b" got %#v`, l)
	}
}

func TestEditor_LineCtrlACtrlE(t *testing.T) {
	in := bytes.NewBuffer([]byte("foo bar\x01\x05\x0d"))
	out := &checkedWriter{
		expectations: []string{
			"\r> \x1b[0K\r\x1b[2C",
			"\r> f\x1b[0K\r\x1b[3C",
			"\r> fo\x1b[0K\r\x1b[4C",
			"\r> foo\x1b[0K\r\x1b[5C",
			"\r> foo \x1b[0K\r\x1b[6C",
			"\r> foo b\x1b[0K\r\x1b[7C",
			"\r> foo ba\x1b[0K\r\x1b[8C",
			"\r> foo bar\x1b[0K\r\x1b[9C",
			"\r> foo bar\x1b[0K\r\x1b[2C",
			"\r> foo bar\x1b[0K\r\x1b[9C",
		},
	}

	e := &linesqueak.Editor{
		In:     bufio.NewReader(in),
		Out:    bufio.NewWriter(out),
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

func TestEditor_LineCtrlL(t *testing.T) {
	in := bytes.NewBuffer([]byte("foo bar\x0c\x0d"))
	out := &checkedWriter{
		expectations: []string{
			"\r> \x1b[0K\r\x1b[2C",
			"\r> f\x1b[0K\r\x1b[3C",
			"\r> fo\x1b[0K\r\x1b[4C",
			"\r> foo\x1b[0K\r\x1b[5C",
			"\r> foo \x1b[0K\r\x1b[6C",
			"\r> foo b\x1b[0K\r\x1b[7C",
			"\r> foo ba\x1b[0K\r\x1b[8C",
			"\r> foo bar\x1b[0K\r\x1b[9C",
			"\x1b[H\x1b[2J\r> foo bar\x1b[0K\r\x1b[9C",
		},
	}

	e := &linesqueak.Editor{
		In:     bufio.NewReader(in),
		Out:    bufio.NewWriter(out),
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

func TestEditor_LineCtrlW(t *testing.T) {
	in := bytes.NewBuffer([]byte("foo  bar \x17\x17\x0d"))
	out := &checkedWriter{
		expectations: []string{
			"\r> \x1b[0K\r\x1b[2C",
			"\r> f\x1b[0K\r\x1b[3C",
			"\r> fo\x1b[0K\r\x1b[4C",
			"\r> foo\x1b[0K\r\x1b[5C",
			"\r> foo \x1b[0K\r\x1b[6C",
			"\r> foo  \x1b[0K\r\x1b[7C",
			"\r> foo  b\x1b[0K\r\x1b[8C",
			"\r> foo  ba\x1b[0K\r\x1b[9C",
			"\r> foo  bar\x1b[0K\r\x1b[10C",
			"\r> foo  bar \x1b[0K\r\x1b[11C",
			"\r> foo  \x1b[0K\r\x1b[7C",
			"\r> \x1b[0K\r\x1b[2C",
		},
	}

	e := &linesqueak.Editor{
		In:     bufio.NewReader(in),
		Out:    bufio.NewWriter(out),
		Prompt: "> ",
	}

	l, err := e.Line()
	if err != nil {
		t.Error(err)
	}
	if l != "" {
		t.Errorf(`expected "" got %#v`, l)
	}
}

func TestEditor_LineEscSquareBracket3Tilda(t *testing.T) {
	in := bytes.NewBuffer([]byte("abc\x02\x02\x1b[3~\x0d"))
	out := &checkedWriter{
		expectations: []string{
			"\r> \x1b[0K\r\x1b[2C",
			"\r> a\x1b[0K\r\x1b[3C",
			"\r> ab\x1b[0K\r\x1b[4C",
			"\r> abc\x1b[0K\r\x1b[5C",
			"\r> abc\x1b[0K\r\x1b[4C",
			"\r> abc\x1b[0K\r\x1b[3C",
			"\r> ac\x1b[0K\r\x1b[3C",
		},
	}

	e := &linesqueak.Editor{
		In:     bufio.NewReader(in),
		Out:    bufio.NewWriter(out),
		Prompt: "> ",
	}

	l, err := e.Line()
	if err != nil {
		t.Error(err)
	}
	if l != "ac" {
		t.Errorf(`expected "ac" got %#v`, l)
	}
}

func TestEditor_LineEscSquareBracketAEscSquareBracketB(t *testing.T) {
	in := bytes.NewBuffer([]byte("\x1b[Afoo\x0d\x1b[A\x1b[A\x1b[B\x1b[Bbar\x0d"))
	out := &checkedWriter{
		expectations: []string{
			"\r> \x1b[0K\r\x1b[2C",
			"\r> \x1b[0K\r\x1b[2C",
			"\r> f\x1b[0K\r\x1b[3C",
			"\r> fo\x1b[0K\r\x1b[4C",
			"\r> foo\x1b[0K\r\x1b[5C",
			"\r> \x1b[0K\r\x1b[2C",
			"\r> foo\x1b[0K\r\x1b[5C",
			"\r> \x1b[0K\r\x1b[2C",
			"\r> foo\x1b[0K\r\x1b[5C",
			"\r> \x1b[0K\r\x1b[2C",
			"\r> b\x1b[0K\r\x1b[3C",
			"\r> ba\x1b[0K\r\x1b[4C",
			"\r> bar\x1b[0K\r\x1b[5C",
		},
	}

	e := &linesqueak.Editor{
		In:     bufio.NewReader(in),
		Out:    bufio.NewWriter(out),
		Prompt: "> ",
	}

	l, err := e.Line()
	if err != nil {
		t.Error(err)
	}
	if l != "foo" {
		t.Errorf(`expected "foo" got %#v`, l)
	}

	e.HistoryAdd("foo")

	l, err = e.Line()
	if err != nil {
		t.Error(err)
	}
	if l != "bar" {
		t.Errorf(`expected "bar" got %#v`, l)
	}
}

func TestEditor_LineEscSquareBracketCEscSquareBracketD(t *testing.T) {
	in := bytes.NewBuffer([]byte("foo bar\x1b[D\x1b[D\x1b[D\x1b[D\x1b[D\x1b[D\x1b[D\x1b[D\x1b[C\x1b[C\x1b[C\x1b[C\x1b[C\x1b[C\x1b[C\x0d"))
	out := &checkedWriter{
		expectations: []string{
			"\r> \x1b[0K\r\x1b[2C",
			"\r> f\x1b[0K\r\x1b[3C",
			"\r> fo\x1b[0K\r\x1b[4C",
			"\r> foo\x1b[0K\r\x1b[5C",
			"\r> foo \x1b[0K\r\x1b[6C",
			"\r> foo b\x1b[0K\r\x1b[7C",
			"\r> foo ba\x1b[0K\r\x1b[8C",
			"\r> foo bar\x1b[0K\r\x1b[9C",
			"\r> foo bar\x1b[0K\r\x1b[8C",
			"\r> foo bar\x1b[0K\r\x1b[7C",
			"\r> foo bar\x1b[0K\r\x1b[6C",
			"\r> foo bar\x1b[0K\r\x1b[5C",
			"\r> foo bar\x1b[0K\r\x1b[4C",
			"\r> foo bar\x1b[0K\r\x1b[3C",
			"\r> foo bar\x1b[0K\r\x1b[2C",
			"\a",
			"\r> foo bar\x1b[0K\r\x1b[3C",
			"\r> foo bar\x1b[0K\r\x1b[4C",
			"\r> foo bar\x1b[0K\r\x1b[5C",
			"\r> foo bar\x1b[0K\r\x1b[6C",
			"\r> foo bar\x1b[0K\r\x1b[7C",
			"\r> foo bar\x1b[0K\r\x1b[8C",
			"\r> foo bar\x1b[0K\r\x1b[9C",
			"\a",
		},
	}

	e := &linesqueak.Editor{
		In:     bufio.NewReader(in),
		Out:    bufio.NewWriter(out),
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

func TestEditor_LineEscSquareBracketHEscSquareBracketF(t *testing.T) {
	in := bytes.NewBuffer([]byte("foo bar\x1b[H\x1b[F\x0d"))
	out := &checkedWriter{
		expectations: []string{
			"\r> \x1b[0K\r\x1b[2C",
			"\r> f\x1b[0K\r\x1b[3C",
			"\r> fo\x1b[0K\r\x1b[4C",
			"\r> foo\x1b[0K\r\x1b[5C",
			"\r> foo \x1b[0K\r\x1b[6C",
			"\r> foo b\x1b[0K\r\x1b[7C",
			"\r> foo ba\x1b[0K\r\x1b[8C",
			"\r> foo bar\x1b[0K\r\x1b[9C",
			"\r> foo bar\x1b[0K\r\x1b[2C",
			"\r> foo bar\x1b[0K\r\x1b[9C",
		},
	}

	e := &linesqueak.Editor{
		In:     bufio.NewReader(in),
		Out:    bufio.NewWriter(out),
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

func TestEditor_LineEscOHEscOF(t *testing.T) {
	in := bytes.NewBuffer([]byte("foo bar\x1bOH\x1bOF\x0d"))
	out := &checkedWriter{
		expectations: []string{
			"\r> \x1b[0K\r\x1b[2C",
			"\r> f\x1b[0K\r\x1b[3C",
			"\r> fo\x1b[0K\r\x1b[4C",
			"\r> foo\x1b[0K\r\x1b[5C",
			"\r> foo \x1b[0K\r\x1b[6C",
			"\r> foo b\x1b[0K\r\x1b[7C",
			"\r> foo ba\x1b[0K\r\x1b[8C",
			"\r> foo bar\x1b[0K\r\x1b[9C",
			"\r> foo bar\x1b[0K\r\x1b[2C",
			"\r> foo bar\x1b[0K\r\x1b[9C",
		},
	}

	e := &linesqueak.Editor{
		In:     bufio.NewReader(in),
		Out:    bufio.NewWriter(out),
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

func TestEditor_LineTabNoCompleteFunc(t *testing.T) {
	in := bytes.NewBuffer([]byte("foo\t\x0d"))
	out := &checkedWriter{
		expectations: []string{
			"\r> \x1b[0K\r\x1b[2C",
			"\r> f\x1b[0K\r\x1b[3C",
			"\r> fo\x1b[0K\r\x1b[4C",
			"\r> foo\x1b[0K\r\x1b[5C",
			"\r> foo\t\x1b[0K\r\x1b[9C",
		},
	}

	e := &linesqueak.Editor{
		In:     bufio.NewReader(in),
		Out:    bufio.NewWriter(out),
		Prompt: "> ",
	}

	l, err := e.Line()
	if err != nil {
		t.Error(err)
	}
	if l != "foo\t" {
		t.Errorf(`expected "foo\t" got %#v`, l)
	}
}

func TestEditor_LineTabNoCompletionAvailable(t *testing.T) {
	in := bytes.NewBuffer([]byte("foo\t\x0d"))
	out := &checkedWriter{
		expectations: []string{
			"\r> \x1b[0K\r\x1b[2C",
			"\r> f\x1b[0K\r\x1b[3C",
			"\r> fo\x1b[0K\r\x1b[4C",
			"\r> foo\x1b[0K\r\x1b[5C",
			"\a",
			"\r> foo\x1b[0K\r\x1b[5C",
		},
	}

	e := &linesqueak.Editor{
		In:     bufio.NewReader(in),
		Out:    bufio.NewWriter(out),
		Prompt: "> ",
		Complete: func(s string) []string {
			if s != "foo" {
				t.Errorf(`expected "foo" got %#v`, s)
			}
			return []string{}
		},
	}

	l, err := e.Line()
	if err != nil {
		t.Error(err)
	}
	if l != "foo" {
		t.Errorf(`expected "foo" got %#v`, l)
	}
}

func TestEditor_LineTabSomeCompletions(t *testing.T) {
	in := bytes.NewBuffer([]byte("foo\t\t\t\t\x0d"))
	out := &checkedWriter{
		expectations: []string{
			"\r> \x1b[0K\r\x1b[2C",
			"\r> f\x1b[0K\r\x1b[3C",
			"\r> fo\x1b[0K\r\x1b[4C",
			"\r> foo\x1b[0K\r\x1b[5C",
			"\r> foo bar\x1b[0K\r\x1b[9C",
			"\r> foo bar baz\x1b[0K\r\x1b[13C",
			"\r> foo\x1b[0K\r\x1b[5C",
			"\r> foo bar\x1b[0K\r\x1b[9C",
		},
	}

	e := &linesqueak.Editor{
		In:     bufio.NewReader(in),
		Out:    bufio.NewWriter(out),
		Prompt: "> ",
		Complete: func(s string) []string {
			if s != "foo" {
				t.Errorf(`expected "foo" got %#v`, s)
			}
			return []string{
				"foo bar",
				"foo bar baz",
			}
		},
	}

	l, err := e.Line()
	if err != nil {
		t.Error(err)
	}
	if l != "foo bar" {
		t.Errorf(`expected "foo bar" got %#v`, l)
	}
}

func TestEditor_LineHint(t *testing.T) {
	in := bytes.NewBuffer([]byte("foo bar\x0d"))
	out := &checkedWriter{
		expectations: []string{
			"\r> \x1b[0K\r\x1b[2C",
			"\r> f\x1b[0K\r\x1b[3C",
			"\r> fo\x1b[0K\r\x1b[4C",
			"\r> foo\x1b[0K\r\x1b[5C",
			"\r> foo \x1b[0;37;49mbar\x1b[0m\x1b[0K\r\x1b[6C",
			"\r> foo b\x1b[0K\r\x1b[7C",
			"\r> foo ba\x1b[0K\r\x1b[8C",
			"\r> foo bar\x1b[0K\r\x1b[9C",
		},
	}

	e := &linesqueak.Editor{
		In:     bufio.NewReader(in),
		Out:    bufio.NewWriter(out),
		Prompt: "> ",
		Hint: func(s string) *linesqueak.Hint {
			if s == "foo " {
				return &linesqueak.Hint{
					Message: "bar",
				}
			}

			return nil
		},
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
	expectations []string
	pos          int
}

var _ io.Writer = (*checkedWriter)(nil)

func (c *checkedWriter) Write(p []byte) (int, error) {
	e := c.expectations[c.pos]
	a := string(p)

	if e != a {
		return 0, fmt.Errorf(`expected %#v got %#v`, e, a)
	}

	c.pos++
	return len(p), nil
}
