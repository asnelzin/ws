package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"github.com/gorilla/websocket"
	"github.com/jessevdk/go-flags"
)

var revision string

var usage = "[OPTIONS] URL"
var opts struct {
	Headers []string `short:"H" long:"header" description:"Custom HTTP header. You can specify as many as needed by repeating the flag."`
	Version bool     `short:"v" long:"version" description:"Print the version information and exit"`
}
var parser = flags.NewParser(&opts, flags.Default)

var (
	green  = color.New(color.FgGreen).SprintfFunc()
	yellow = color.New(color.FgYellow).SprintfFunc()
	blue   = color.New(color.FgBlue).SprintfFunc()
)

func main() {
	parser.Usage = usage

	args, err := parser.Parse()
	if err != nil {
		os.Exit(1)
	}

	if opts.Version {
		showVersion()
		os.Exit(0)
	}

	if len(args) != 1 {
		errorf("the required argument URL was not provided")
		os.Exit(1)
	}

	u, err := url.Parse(args[0])
	if err != nil {
		errorf("could not parse URL: %v", err)
		os.Exit(1)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill)

	rl, err := readline.NewEx(&readline.Config{
		Prompt: green(">>> "),
	})
	if err != nil {
		errorf("could not create prompt: %v", err)
		os.Exit(2)
	}
	defer rl.Close()

	headers := parseHeaderOpts(opts.Headers)

	c, _, err := websocket.DefaultDialer.Dial(u.String(), headers)
	if err != nil {
		errorf("could not dial %s: %v", u.String(), err)
		os.Exit(2)
	}
	defer c.Close()

	showVersion()
	fmt.Printf("Connected to %s\n\n", green(u.String()))

	done := make(chan struct{})
	go readPump(c, rl, done)
	go writePump(c, rl, done, interrupt)

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				errorf("write close: %v", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func parseHeaderOpts(hs []string) http.Header {
	headers := http.Header{}
	for _, h := range hs {
		kv := strings.Split(h, ":")
		if len(kv) != 2 {
			continue
		}
		key, val := kv[0], kv[1]
		if len(key) == 0 || len(val) == 0 {
			continue
		}
		key, val = strings.TrimSpace(key), strings.TrimSpace(val)
		headers.Add(key, val)
	}
	return headers
}

func writePump(conn *websocket.Conn, rl *readline.Instance, done chan struct{}, interrupt chan os.Signal) {
	for {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			continue
		} else if err != nil {
			close(interrupt)
			return
		}
		if len(line) == 0 {
			continue
		}
		err = conn.WriteMessage(websocket.TextMessage, []byte(line))
		if err != nil {
			errorf("write close: %v", err)
			close(done)
			return
		}
	}
}

func readPump(conn *websocket.Conn, rl *readline.Instance, done chan struct{}) {
	defer close(done)
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			errorf("read: %v", err)
			return
		}
		printReceived(rl.Stdout(), message)
	}
}

func printReceived(w io.Writer, message []byte) {
	t := time.Now()
	timeStr := fmt.Sprintf("%02d:%02d:%02d", t.Hour(), t.Minute(), t.Second())
	fmt.Fprintf(w, "%s %s\n", yellow(timeStr), blue(string(message)))
}

func showVersion() {
	fmt.Printf("ws version %s\n", revision)
}

func errorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "ws: "+format+"\n", args...)
}
