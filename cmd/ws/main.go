package main

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

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
		os.Exit(2)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u, err := url.Parse(args[0])
	if err != nil {
		errorf("could not parse URL: %v", err)
	}

	headers := http.Header{}
	for _, h := range opts.Headers {
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

	fmt.Printf("Connecting to %s\n", u.String())
	c, _, err := websocket.DefaultDialer.Dial(u.String(), headers)
	if err != nil {
		errorf("could not dial %s: %v", u.String(), err)
		os.Exit(2)
	}
	defer c.Close()

	// reading
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				errorf("read: %v", err)
				return
			}
			printReceived(message)
		}
	}()

	// writing
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			msg := scanner.Bytes()
			err := c.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				errorf("write close: %v", err)
				return
			}
		}
	}()

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

var (
	yellow = color.New(color.FgYellow).SprintfFunc()
	blue   = color.New(color.FgBlue).SprintfFunc()
)

func printReceived(message []byte) {
	t := time.Now()
	timeStr := fmt.Sprintf("%02d:%02d:%02d", t.Hour(), t.Minute(), t.Second())
	fmt.Printf("%s %s\n", yellow(timeStr), blue(string(message)))

}

func errorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "ws: "+format+"\n", args...)
}

func showVersion() {
	fmt.Printf("ws version %s\n", revision)
}
