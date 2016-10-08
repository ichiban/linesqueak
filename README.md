# Linesqueak

[![Build Status](https://travis-ci.org/ichiban/linesqueak.svg?branch=master)](https://travis-ci.org/ichiban/linesqueak)

Linesqueak is a simple pure-Go line editor.
It speaks to `io.Reader` and `io.Writer` instead of pty/tty,
which makes it easy to integrate with network based applications (see [examples/ssh](https://github.com/ichiban/linesqueak/blob/master/examples/ssh/main.go)).

It is inspired by [Linenoise](https://github.com/antirez/linenoise).

# Features

- [x] Standard Key Bindings
- [x] History
- [ ] Completion
- [ ] Hints
 
# Similar Projects

- [Readline](https://github.com/chzyer/readline)
- [Liner](https://github.com/peterh/liner)

# License

See the LICENSE file for license rights and limitations (MIT).