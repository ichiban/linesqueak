package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os/exec"

	"golang.org/x/crypto/ssh"

	"github.com/ichiban/linesqueak"
)

func main() {
	config := &ssh.ServerConfig{
		NoClientAuth: true,
	}

	key, err := serverPrivateKey()
	if err != nil {
		panic("Failed to load private key")
	}

	config.AddHostKey(key)

	listener, err := net.Listen("tcp", "0.0.0.0:2022")
	if err != nil {
		log.Fatalf("Failed to listen on 2022 (%s)", err)
	}

	for {
		tcp, err := listener.Accept()
		if err != nil {
			log.Printf("failed to accept tcp connection: %s", err)
			continue
		}

		conn, chans, reqs, err := ssh.NewServerConn(tcp, config)
		if err != nil {
			log.Printf("failed to handshake: %s", err)
			continue
		}

		log.Printf("connection from %s (%s)", conn.RemoteAddr(), conn.ClientVersion())
		go ssh.DiscardRequests(reqs)
		go handleChannels(chans)
	}
}

func handleChannels(chans <-chan ssh.NewChannel) {
	for c := range chans {
		go handleChannel(c)
	}
}

func handleChannel(c ssh.NewChannel) {
	if t := c.ChannelType(); t != "session" {
		msg := fmt.Sprintf("unknown channel type: %s", t)
		c.Reject(ssh.UnknownChannelType, msg)
		return
	}

	conn, reqs, err := c.Accept()
	if err != nil {
		log.Printf("failed to accept channel: %s", err)
		return
	}
	defer conn.Close()

	e := &linesqueak.Editor{
		In:     bufio.NewReader(conn),
		Out:    bufio.NewWriter(conn),
		Prompt: "> ",
		Complete: func(_ string) []string {
			return []string{
				"Completion #1",
				"Completion #2",
				"Completion #3",
			}
		},
		Hint: func(s string) *linesqueak.Hint {
			if s == "foo " {
				return &linesqueak.Hint{
					Message: "bar baz",
				}
			}

			if s == "foo bar " {
				return &linesqueak.Hint{
					Message: "baz",
					Bold:    true,
				}
			}

			return nil
		},
	}

	go func() {
		for req := range reqs {
			switch req.Type {
			case "pty-req":
				termLen := req.Payload[3]
				w, h := parseDims(req.Payload[termLen+4:])
				e.Rows = h
				e.Cols = w
				req.Reply(true, nil)
			case "shell":
				term := string(req.Payload)
				for _, t := range linesqueak.SupportedTerms {
					if t == term {
						req.Reply(true, nil)
					}
				}
			case "exec":
				log.Printf("exec: %s", req.Payload)
			case "window-change":
				w, h := parseDims(req.Payload)
				e.Rows = h
				e.Cols = w
			default:
				log.Printf("unknown req type: %s", req.Type)
			}
		}
	}()

	for {
		line, err := e.Line()
		if err != nil {
			break
		}

		log.Printf("line: %s\n", line)
		fmt.Fprintf(e.Out, "\ryou have typed: %s\n", line)

		e.History.Add(line)

		if line == "adjust" {
			log.Printf("adjusting: (%d, %d)\n", e.Cols, e.Rows)
			e.Adjust()
			log.Printf("adjusted: (%d, %d)\n", e.Cols, e.Rows)
		}
	}
}

func parseDims(b []byte) (int, int) {
	w := int(binary.BigEndian.Uint32(b))
	h := int(binary.BigEndian.Uint32(b[4:]))
	return w, h
}

func serverPrivateKey() (ssh.Signer, error) {
	b, err := serverPrivateKeyBytes()
	if err != nil {
		return nil, err
	}
	return ssh.ParsePrivateKey(b)
}

func serverPrivateKeyBytes() ([]byte, error) {
	if key, err := ioutil.ReadFile("example.rsa"); err == nil {
		return key, err
	}

	if err := exec.Command("ssh-keygen", "-f", "example.rsa", "-t", "rsa", "-N", "").Run(); err != nil {
		log.Fatalf("Failed to generate example.rsa: %s", err)
	}

	return ioutil.ReadFile("example.rsa")
}
