# Linesqueak

[![Build Status](https://travis-ci.org/ichiban/linesqueak.svg?branch=master)](https://travis-ci.org/ichiban/linesqueak)

Linesqueak is a simple pure-Go line editor.
It speaks to `io.Reader` and `io.Writer` instead of pty/tty,
which makes it easy to integrate with network based applications (see [examples/ssh](https://github.com/ichiban/linesqueak/blob/master/examples/ssh/main.go)).

It is inspired by [Linenoise](https://github.com/antirez/linenoise).

# Features

- [x] Standard Key Bindings
- [x] History
- [x] Completion
- [x] Hints

# Basic Usage

```go
e := &linesqueak.Editor{
	In:     bufio.NewReader(r), // `r` is `io.Reader` (e.g. `ssh.Channel`)
	Out:    bufio.NewWriter(w), // `w` is `io.Writer` (e.g. `ssh.Channel`)
	Prompt: "> ",               // you can be creative here :)
}

// you may want to process multiple lines
// if so, call `Line()` in a loop
for {
	line, err := e.Line() // `Line()` does all the editor things and returns input line
	if err != nil {
		panic(err) // TODO: proper error handling
	}
	
	fmt.Fprintf(e.Out, "\ryou have typed: %s\n", line)
}
```

# Similar Projects

- [Readline](https://github.com/chzyer/readline)
- [Liner](https://github.com/peterh/liner)

# License

See the LICENSE file for license rights and limitations (MIT).