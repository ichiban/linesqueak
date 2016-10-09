package linesqueak

import (
	"bufio"
	"container/ring"
	"errors"
	"fmt"
	"io"
)

type Editor struct {
	In       *bufio.Reader
	Out      *bufio.Writer
	Prompt   string
	Buffer   []rune
	Pos      int
	History  *ring.Ring
	Complete func(s string) []string
	Hint     func(s string) *Hint
	Width    func(rune) int
}

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

func (e *Editor) editReset() error {
	if e.History.Len() == 0 {
		e.History = ring.New(1)
		e.History.Value = ""
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
	if err := e.refreshSingleLine(); err != nil {
		return err
	}
	if err := e.Out.Flush(); err != nil {
		return err
	}
	return nil
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

func (e *Editor) refreshSingleLine() error {
	ew := &errWriter{w: e.Out}
	ew.WriteString("\r") // cursor to left edge
	ew.WriteString(e.Prompt)
	ew.WriteString(string(e.Buffer))
	ew.WriteString(e.hint())
	ew.WriteString("\x1b[0K")                            // erase to right
	ew.WriteString(fmt.Sprintf("\r\x1b[%dC", e.width())) // move cursor to original position
	return ew.err
}

func (e *Editor) width() int {
	f := func(r rune) int {
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
	Color   Color
	Bold    bool
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
		h.Color = Gray
	}

	var b int
	if h.Bold {
		b = 1
	}

	if h.Color != 0 || !h.Bold {
		return fmt.Sprintf("\x1b[%d;%d;49m%s\x1b[0m", b, h.Color, h.Message)
	}

	return h.Message
}

type Color byte

const (
	Gray         = 37
	Default      = 49
	Black        = 40
	Red          = 41
	Green        = 42
	Yellow       = 43
	Blue         = 44
	Magenta      = 45
	Cyan         = 46
	LightGray    = 47
	DarkGray     = 100
	LightRed     = 101
	LightGreen   = 102
	LightYellow  = 103
	LightBlue    = 104
	LightMagenta = 105
	LightCyan    = 106
	White        = 107
)

// https://blog.golang.org/errors-are-values
type errWriter struct {
	w   *bufio.Writer
	err error
}

func (ew *errWriter) WriteString(s string) {
	if ew.err != nil {
		return
	}
	_, ew.err = ew.w.WriteString(s)
}
