// Package linesqueak provides readline-like line editing functionality over io.Reader & io.Writer.
package linesqueak

import (
	"bufio"
	"container/ring"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
)

// Editor interacts with VT100 like terminals via io.Reader & io.Writer and displays an input line.
type Editor struct {
	// In reads key strokes from the terminal.
	// It's required to be raw mode i.e. send data as frequently as possible
	// instead of cooked mode i.e. buffer it until it reaches a meaningful chunk (= line).
	In *bufio.Reader

	// Out displays current editor states to the terminal.
	Out *bufio.Writer

	// Prompt is prepended to each editor state on the terminal.
	// Prompt is not part of user inputs but part of UI. so it doesn't appear in result input lines.
	Prompt string

	// Buffer keeps the current user input.
	Buffer []rune

	// Pos points the current cursor position in Buffer.
	Pos int

	// Cols is the terminal width.
	// If it's not provided, Editor assumes it's 80.
	Cols int

	// Rows is the terminal height.
	// If it's not provided, Editor assumes it's 24.
	Rows int

	// History holds previous input lines so that user can reuse or tweak it later.
	// It is your task to add lines to History, save History, or load it from disks.
	History *ring.Ring

	// Complete will be called when user wants you to complete their inputs.
	// It takes the current user input and returns some completion suggestions.
	// Complete is OPTIONAL. If no Complete is provided, completion will be disabled.
	Complete func(s string) []string

	// Hint will be called while user is typing and displayed on the right of the user input.
	// Hint is OPTIONAL. If no Hint is provided, no hint will be shown.
	Hint func(s string) *Hint

	// Width calculates character width on the terminal.
	// A lot of CJK characters and emojis are twice as wide as ASCII characters.
	// Width is OPTIONAL. By default,
	// it calculates the character width as 1 for all characters except tab which width is 4.
	Width func(rune) int

	// OldPos points the previous cursor position in Buffer.
	OldPos int

	// MaxRows is the height of editor status on the terminal.
	MaxRows int
}

// Line reads user key strokes and returns a confirmed input line while displaying editor states on the terminal.
func (e *Editor) Line() (string, error) {
	if err := e.editReset(); err != nil {
		return string(e.Buffer), err
	}
line:
	for {
		r, _, err := e.In.ReadRune()
		if err != nil {
			return string(e.Buffer), err
		}

		switch r {
		case enter:
			break line
		case ctrlC:
			return string(e.Buffer), errors.New("try again")
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
		case esc:
			r, _, err := e.In.ReadRune()
			if err != nil {
				return string(e.Buffer), err
			}

			switch r {
			case '[':
				r, _, err := e.In.ReadRune()
				if err != nil {
					return string(e.Buffer), err
				}

				switch r {
				case '0', '1', '2', '4', '5', '6', '7', '8', '9':
					_, _, err := e.In.ReadRune()
					if err != nil {
						return string(e.Buffer), err
					}
				case '3':
					r, _, err := e.In.ReadRune()
					if err != nil {
						return string(e.Buffer), err
					}

					switch r {
					case '~':
						if err := e.editDelete(); err != nil {
							return string(e.Buffer), err
						}
					}
				case 'A':
					if err := e.editHistoryPrev(); err != nil {
						return string(e.Buffer), err
					}
				case 'B':
					if err := e.editHistoryNext(); err != nil {
						return string(e.Buffer), err
					}
				case 'C':
					if err := e.editMoveRight(); err != nil {
						return string(e.Buffer), err
					}
				case 'D':
					if err := e.editMoveLeft(); err != nil {
						return string(e.Buffer), err
					}
				case 'H':
					if err := e.editMoveHome(); err != nil {
						return string(e.Buffer), err
					}
				case 'F':
					if err := e.editMoveEnd(); err != nil {
						return string(e.Buffer), err
					}
				}
			case 'O':
				r, _, err := e.In.ReadRune()
				if err != nil {
					return string(e.Buffer), err
				}

				switch r {
				case 'H':
					if err := e.editMoveHome(); err != nil {
						return string(e.Buffer), err
					}
				case 'F':
					if err := e.editMoveEnd(); err != nil {
						return string(e.Buffer), err
					}
				}
			}
		case tab:
			if err := e.completeLine(); err != nil {
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

// HistoryAdd adds a line to History.
func (e *Editor) HistoryAdd(l string) {
	if e.History == nil {
		e.History = ring.New(1)
		e.History.Value = ""
	}

	// Don't add duplicate lines
	if e.History.Value.(string) == l {
		return
	}

	r := ring.New(1)
	r.Value = l
	e.History.Prev().Link(r)
}

var dimPattern = regexp.MustCompile("\x1b\\[(\\d+);(\\d+)R")

// AdjustDimensions queries the terminal about rows and cols and updates Editor's Rows and Cols.
func (e *Editor) AdjustDimensions() error {
	// https://groups.google.com/forum/#!topic/comp.os.vms/bDKSY6nG13k
	if _, err := e.Out.WriteString("\x1b7\x1b[999;999H\x1b[6n"); err != nil {
		return err
	}

	if err := e.Out.Flush(); err != nil {
		return err
	}

	res, err := e.In.ReadString('R')
	if err != nil {
		return err
	}

	ms := dimPattern.FindStringSubmatch(res)
	r, err := strconv.Atoi(ms[1])
	if err != nil {
		return err
	}
	c, err := strconv.Atoi(ms[2])
	if err != nil {
		return err
	}

	if _, err := e.Out.WriteString("\x1b8"); err != nil {
		return err
	}

	e.Cols = c
	e.Rows = r

	return nil
}

func (e *Editor) editReset() error {
	if e.History.Len() == 0 {
		e.History = ring.New(1)
		e.History.Value = ""
	}
	e.Buffer = []rune{}
	e.OldPos = 0
	e.Pos = 0
	e.MaxRows = 0

	if e.Rows == 0 {
		e.Rows = 24
	}

	if e.Cols == 0 {
		e.Cols = 80
	}

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
	p := e.Pos
	if p == len(e.Buffer) {
		p = len(e.Buffer) - 1
	}

	if p == 0 {
		e.beep()
		return nil
	}

	e.Buffer[p-1], e.Buffer[p] = e.Buffer[p], e.Buffer[p-1]

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
	if e.History.Len() == 0 {
		e.beep()
		return nil
	}

	e.History.Value = string(e.Buffer)
	e.History = e.History.Prev()
	e.Buffer = []rune(e.History.Value.(string))
	e.Pos = len(e.Buffer)
	e.refreshLine()

	return nil
}

func (e *Editor) editHistoryNext() error {
	if e.History.Len() == 0 {
		e.beep()
		return nil
	}

	e.History.Value = string(e.Buffer)
	e.History = e.History.Next()
	e.Buffer = []rune(e.History.Value.(string))
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

func (e *Editor) completeLine() error {
	if e.Complete == nil {
		return e.editInsert(tab)
	}

	opts := e.Complete(string(e.Buffer))

	if len(opts) == 0 {
		e.beep()
		return nil
	}

	cs := ring.New(len(opts) + 1)
	for _, o := range opts {
		cs.Value = o
		cs = cs.Next()
	}
	cs.Value = string(e.Buffer)
	cs = cs.Next()

complete:
	for {
		c := cs.Value.(string)

		if err := e.refreshLineString(c); err != nil {
			return err
		}

		b, err := e.In.Peek(1)
		if err != nil {
			return err
		}

		switch b[0] {
		case tab:
			if _, _, err := e.In.ReadRune(); err != nil {
				return err
			}
			cs = cs.Next()
		case esc:
			if _, _, err := e.In.ReadRune(); err != nil {
				return err
			}
			if err := e.refreshLine(); err != nil {
				return err
			}
			break complete
		default:
			e.Buffer = []rune(c)
			e.Pos = len(e.Buffer)
			break complete
		}
	}

	return nil
}

const (
	ctrlA     = 1
	ctrlB     = 2
	ctrlC     = 3
	ctrlD     = 4
	ctrlE     = 5
	ctrlF     = 6
	ctrlH     = 8
	tab       = 9
	ctrlK     = 11
	ctrlL     = 12
	enter     = 13
	ctrlN     = 14
	ctrlP     = 16
	ctrlT     = 20
	ctrlU     = 21
	ctrlW     = 23
	esc       = 27
	space     = 32
	backspace = 127
)

// SupportedTerms is a list of supported terminals.
var SupportedTerms = []string{"dumb", "cons25", "emacs"}

func (e *Editor) clearScreen() error {
	n, err := e.Out.WriteString("\x1b[H\x1b[2J")
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
	h := e.hint()

	f := defaultWidth
	if e.Width != nil {
		f = e.Width
	}

	var pw int
	for _, r := range e.Prompt {
		pw += f(r)
	}

	var bw, cw, ocw int
	for i, r := range e.Buffer {
		if i < e.Pos {
			cw += f(r)
		}
		if i < e.OldPos {
			ocw += f(r)
		}
		bw += f(r)
	}

	var hw int
	for _, r := range h {
		hw += f(r)
	}

	ep := pos{
		cols: (pw + bw + hw) % e.Cols,
		rows: (pw + bw + hw) / e.Cols,
	}

	cp := pos{
		cols: (pw + cw) % e.Cols,
		rows: (pw + cw) / e.Cols,
	}

	ocp := pos{
		cols: (pw + ocw) % e.Cols,
		rows: (pw + ocw) / e.Cols,
	}

	ew := &errWriter{w: e.Out}

	oldRows := e.MaxRows
	if ep.rows > e.MaxRows {
		e.MaxRows = ep.rows
	}

	// go to the bottom of editor region
	if oldRows - ocp.rows > 0 {
		ew.writeString(fmt.Sprintf("\x1b[%dB", oldRows - ocp.rows))
	}

	for i := 1; i < oldRows; i++ {
		ew.writeString("\x1b[2K") // kill line
		ew.writeString("\x1b[1A") // go up
	}

	ew.writeString("\r")
	ew.writeString(e.Prompt)
	ew.writeString(string(e.Buffer))
	ew.writeString(h)
	ew.writeString("\x1b[0K")

	// If we are at the right edge,
	// move cursor to the beginning of next line.
	if e.Pos == len(e.Buffer) && cp.cols == 0 {
		ew.writeString("\n\r")
		cp.rows++
		ep.rows++
		if ep.rows > e.MaxRows {
			e.MaxRows = ep.rows
		}
	}

	// Go up till we reach the expected position.
	if ep.rows - cp.rows > 0 {
		ew.writeString(fmt.Sprintf("\x1b[%dA", ep.rows - cp.rows))
	}

	ew.writeString("\r")
	if cp.cols > 0 {
		ew.writeString(fmt.Sprintf("\x1b[%dC", cp.cols))
	}

	ew.flush()

	e.OldPos = e.Pos

	return ew.err
}

func (e *Editor) refreshLineString(s string) error {
	b := e.Buffer
	p := e.Pos
	e.Buffer = []rune(s)
	e.Pos = len(e.Buffer)
	if err := e.refreshLine(); err != nil {
		return err
	}
	e.Buffer = b
	e.Pos = p
	return nil
}

func defaultWidth(r rune) int {
	if r == tab {
		return 4
	}

	return 1
}

// Hint displays helpful message with styles on the right of user input.
type Hint struct {
	// Message is the message to be displayed.
	Message string

	// Color is the text color of the hint.
	Color Color

	// Bold increases intensity if true.
	Bold bool
}

func (e *Editor) hint() string {
	if e.Hint == nil {
		return ""
	}

	h := e.Hint(string(e.Buffer))

	if h == nil {
		return ""
	}

	if h.Color == 0 {
		h.Color = White
	}

	var b int
	if h.Bold {
		b = 1
	}

	return fmt.Sprintf("\x1b[%d;%d;49m%s\x1b[0m", b, h.Color, h.Message)
}

// Color represents text color.
type Color byte

const (
	Black Color = 30 + iota
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
)

// https://blog.golang.org/errors-are-values
type errWriter struct {
	w   *bufio.Writer
	err error
}

func (ew *errWriter) writeString(s string) {
	if ew.err != nil {
		return
	}
	_, ew.err = ew.w.WriteString(s)
}

func (ew *errWriter) flush() {
	if ew.err != nil {
		return
	}
	ew.err = ew.w.Flush()
}

type pos struct {
	cols, rows int
}
