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
	In         *bufio.Reader
	Out        *bufio.Writer
	Prompt     string
	Buffer     []rune
	Pos        int
	History    []string
	HistoryPos int
	Hint       func (*Editor) *Hint
	Width      func (rune) int
}

func (e *Editor) Line() (string, error) {
	if err := e.editReset(); err != nil {
		return string(e.Buffer), err
	}
line:
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
			break line
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
		case ctrlP:
			if err := e.editHistoryPrev(); err != nil {
				return string(e.Buffer), err
			}
		case ctrlN:
			if err := e.editHistoryNext(); err != nil {
				return string(e.Buffer), err
			}
		case ctrlU:
			if err := e.editReset(); err != nil {
				return string(e.Buffer), err
			}
		case ctrlK:
			if err := e.editKillForward(); err != nil {
				return string(e.Buffer), err
			}
		case ctrlA:
			if err := e.editMoveHome(); err != nil {
				return string(e.Buffer), err
			}
		case ctrlE:
			if err := e.editMoveEnd(); err != nil {
				return string(e.Buffer), err
			}
		case ctrlL:
			if err := e.clearScreen(); err != nil {
				return string(e.Buffer), err
			}

			if err := e.refreshLine(); err != nil {
				return string(e.Buffer), err
			}
		case ctrlW:
			if err := e.editDeletePrevWord(); err != nil {
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

func (e *Editor) editReset() error {
	if len(e.History) == 0 {
		e.History = []string{""}
	}
	e.Buffer = []rune{}
	e.Pos = 0

	return e.refreshLine()
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

func (e *Editor) editHistoryPrev() error {
	if len(e.History) <= 0 {
		e.beep()
		return nil
	}

	e.History[e.HistoryPos] = string(e.Buffer)

	if 0 < e.HistoryPos {
		e.HistoryPos--
	} else {
		e.HistoryPos = 0
	}

	var b []rune
	for _, r := range e.History[e.HistoryPos] {
		b = append(b, r)
	}

	e.Buffer = b
	e.Pos = len(e.Buffer)
	e.refreshLine()

	return nil
}

func (e *Editor) editHistoryNext() error {
	if len(e.History) <= 0 {
		e.beep()
		return nil
	}

	e.History[e.HistoryPos] = string(e.Buffer)

	if e.HistoryPos < len(e.History) - 1 {
		e.HistoryPos++
	} else {
		e.HistoryPos = len(e.History) - 1
	}

	var b []rune
	for _, r := range e.History[e.HistoryPos] {
		b = append(b, r)
	}

	e.Buffer = b
	e.Pos = len(e.Buffer)
	e.refreshLine()

	return nil
}

func (e *Editor) editKillForward() error {
	e.Buffer = e.Buffer[:e.Pos]
	return e.refreshLine()
}

func (e *Editor) editMoveHome() error {
	if e.Pos == 0 {
		e.beep()
		return nil
	}

	e.Pos = 0
	return e.refreshLine()
}

func (e *Editor) editMoveEnd() error {
	if e.Pos == len(e.Buffer) {
		e.beep()
		return nil
	}

	e.Pos = len(e.Buffer)
	return e.refreshLine()
}

func (e *Editor) editDeletePrevWord() error {
	var w bool
	var p int
	for i := e.Pos - 1; i >= 0; i-- {
		if e.Buffer[i] != space {
			w = true // found a word to delete
			continue
		}

		if !w {
			continue
		}

		p = i + 1
		break
	}

	e.Buffer = e.Buffer[:p]
	e.Pos = p
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
	space = 32
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

func (e *Editor) clearScreen() error {
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
	ew.WriteString("\r") // cursor to left edge
	ew.WriteString(e.Prompt)
	ew.WriteString(string(e.Buffer))
	if e.Hint != nil {
		h := e.Hint(e)
		ew.WriteString(h.String())
	}
	ew.WriteString("\x1B[0K") // erase to right
	ew.WriteString(fmt.Sprintf("\r\x1B[%dC", e.width())) // move cursor to original position
	return ew.err
}

func (e *Editor) width() int {
	f := func (r rune) int {
		return 1
	}

	if e.Width != nil {
		f = e.Width
	}

	var w int
	for _, r := range e.Prompt {
		w += f(r)
	}
	for _, r := range e.Buffer[:e.Pos] {
		w += f(r)
	}

	return w
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
