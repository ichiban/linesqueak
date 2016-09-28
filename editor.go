package linesqueak

import (
	"bufio"
	"errors"
	"io"
	"regexp"
	"strconv"
	"fmt"
)

type Editor struct {
	In        *bufio.Reader
	Out       *bufio.Writer
	Buffer    []rune
	Prompt    string
	MultiLine bool
	Cols      int
	Rows      int
	Pos       int
	History   []string
	Hint      func (*Editor) *Hint
}

func (e *Editor) Line() (string, error) {
	e.reset()

	if _, err := e.Out.WriteString(e.Prompt); err != nil {
		return "", err
	}
	if err := e.Out.Flush(); err != nil {
		return "", err
	}

	for {
		r, _, err := e.In.ReadRune()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		switch r {
		case enter:
			err := e.refreshLine()
			return string(e.Buffer), err
		case ctrlC:
			return "", errors.New("try again")
		case backspace, ctrlH:
			if err := e.editBackspace(); err != nil {
				return string(e.Buffer), err
			}
		case ctrlD:
			if len(e.Buffer) == 0 {
				return string(e.Buffer), io.EOF
			}

			if err := e.editDelete(); err != nil {
				return string(e.Buffer), err
			}
		case ctrlT:
			if err := e.editSwap(); err != nil {
				return string(e.Buffer), err
			}
		case ctrlB:
			if err := e.editMoveLeft(); err != nil {
				return string(e.Buffer), err
			}
		case ctrlF:
			if err := e.editMoveRight(); err != nil {
				return string(e.Buffer), err
			}
		default:
			if err := e.editInsert(r); err != nil {
				return string(e.Buffer), err
			}
		}
	}

	return string(e.Buffer), nil
}

func (e *Editor) reset() {
	if len(e.History) == 0 {
		e.History = []string{""}
	}
	e.Buffer = []rune{}
	e.Pos = 0
}

func (e *Editor) editBackspace() error {
	if e.Pos == 0 {
		e.beep()
		return nil
	}

	e.Pos--

	// Delete https://github.com/golang/go/wiki/SliceTricks
	e.Buffer = e.Buffer[:e.Pos+copy(e.Buffer[e.Pos:], e.Buffer[e.Pos+1:])]

	return e.refreshLine()
}

func (e *Editor) editDelete() error {
	if e.Pos == len(e.Buffer) {
		e.beep()
		return nil
	}

	// Delete https://github.com/golang/go/wiki/SliceTricks
	e.Buffer = e.Buffer[:e.Pos+copy(e.Buffer[e.Pos:], e.Buffer[e.Pos+1:])]

	return e.refreshLine()
}

func (e *Editor) editSwap() error {
	if e.Pos == 0 || e.Pos == len(e.Buffer){
		e.beep()
		return nil
	}

	e.Buffer[e.Pos-1], e.Buffer[e.Pos] = e.Buffer[e.Pos], e.Buffer[e.Pos-1]

	if e.Pos < len(e.Buffer) {
		e.Pos++
	}

	return e.refreshLine()
}

func (e *Editor) editMoveLeft() error {
	if e.Pos == 0 {
		e.beep()
		return nil
	}

	e.Pos--

	return e.refreshLine()
}

func (e *Editor) editMoveRight() error {
	if e.Pos == len(e.Buffer) {
		e.beep()
		return nil
	}

	e.Pos++

	return e.refreshLine()
}

func (e *Editor) editInsert(r rune) error {
	// Insert https://github.com/golang/go/wiki/SliceTricks
	e.Buffer = append(e.Buffer, 0)
	copy(e.Buffer[e.Pos+1:], e.Buffer[e.Pos:])
	e.Buffer[e.Pos] = r

	e.Pos++
	return e.refreshLine()
}

const (
	ctrlA = 1
	ctrlB = 2
	ctrlC = 3
	ctrlD = 4
	ctrlE = 5
	ctrlF = 6
	ctrlH = 8
	tab = 9
	ctrlK = 11
	ctrlL = 12
	enter = 13
	ctrlN = 14
	ctrlP = 16
	ctrlT = 20
	ctrlU = 21
	ctrlW = 23
	esc = 27
	backspace = 127
)

// SupportedTerm returns false if the given terminal name is in the list of terminals we don't support.
// Otherwise returns true.
func SupportedTerm(term string) bool {
	for _, t := range []string{"dumb", "cons25", "emacs"} {
		if t == term {
			return false
		}
	}
	return true
}

var cursorPos = regexp.MustCompile("\x1B\\[(?P<cols>\\d+);(?P<rows>\\d+)R")

// CursorPos queries the horizontal cursor position and returns it.
// It uses the ESC [6n escape sequence.
func (e *Editor) CursorPos() (int, error) {
	n, err := e.Out.Write([]byte("\x1B[6n"))
	if err != nil {
		return 0, err
	}
	if n != 4 {
		return n, errors.New("failed to query cursor position")
	}

	buf := make([]byte, 32)
	_, err = e.In.Read(buf)
	if err != nil {
		return 0, err
	}

	match := cursorPos.FindStringSubmatch(string(buf))
	if match == nil {
		return 0, errors.New("invalid response")
	}
	for i, name := range cursorPos.SubexpNames() {
		if name == "cols" {
			cols := match[i]
			return strconv.Atoi(cols)
		}
	}
	return 0, errors.New("cols not found")
}

// ClearScreen clears entire screen.
func (e *Editor) ClearScreen() error {
	n, err := e.Out.Write([]byte("\x1B[H\x1B[2J"))
	if err != nil {
		return err
	}
	if n != 7 {
		return errors.New("failed to clear screen")
	}
	return nil
}

func (e *Editor) beep() error {
	if _, err := e.Out.WriteString("\a"); err != nil {
		return err
	}
	if err := e.Out.Flush(); err != nil {
		return err
	}
	return nil
}

func (e *Editor) refreshLine() error {
	if err := e.refreshSingleLine(); err != nil {
		return err
	}
	if err := e.Out.Flush(); err != nil {
		return err
	}
	return nil
}

func (e *Editor) refreshSingleLine() error {
	ew := &errWriter{w: e.Out}
	ew.WriteString("\r")
	ew.WriteString(e.Prompt)
	ew.WriteString(string(e.Buffer))
	if e.Hint != nil {
		h := e.Hint(e)
		ew.WriteString(h.String())
	}
	ew.WriteString("\x1B[0K")
	return ew.err
}

type Hint struct {
	Message string
	Color *BGColor
	Bold bool
}

func (h *Hint) String() string {
	if h.Color != nil || !h.Bold {
		return fmt.Sprintf("\x1B[%d;%d;49m", h.Bold, *h.Color)
	}
	return h.Message
}

type BGColor byte

const (
	BGDefault = 49
	BGBlack = 40
	BGRed = 41
	BGGreen = 42
	BGYellow = 43
	BGBlue = 44
	BGMagenta = 45
	BGCyan = 46
	BGLightGray = 47
	BGDarkGray = 100
	BGLightRed =  101
	BGLightGreen = 102
	BGLightYellow = 103
	BGLightBlue = 104
	BGLightMagenta = 105
	BGLightCyan = 106
	BGWhite = 107
)

// https://blog.golang.org/errors-are-values
type errWriter struct {
	w *bufio.Writer
	err error
}

func (ew *errWriter) WriteString(s string) {
	if ew.err != nil {
		return
	}
	_, ew.err = ew.w.WriteString(s)
}
