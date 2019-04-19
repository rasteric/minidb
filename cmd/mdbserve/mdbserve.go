// The minidb command line tool

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	minidb "github.com/rasteric/minidb"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"nanomsg.org/go/mangos/v2"
	"nanomsg.org/go/mangos/v2/protocol/rep"

	// register transports
	_ "nanomsg.org/go/mangos/v2/transport/all"
)

// Constants used for indicating different types of network or I/O errors.
const (
	ErrNone int = iota + 1
	ErrSyntaxError
	ErrServerFail
	ErrNoSocket
	ErrListen
	ErrRecv
	ErrUnmarshal
	ErrMarshal
	ErrSendIO
)

type errmsg struct {
	number int
	msg    string
}

// ServerLoop starts the main server loop, listening for incoming client connections.
func serverLoop(url string, ctx context.Context, ch chan errmsg, timeout time.Duration) {
	var sock mangos.Socket
	var err error
	var msg []byte
	if sock, err = rep.NewSocket(); err != nil {
		ch <- errmsg{ErrNoSocket, fmt.Sprintf("can't get new socket, %s", err)}
		return
	}
	defer sock.Close()
	if err = sock.Listen(url); err != nil {
		ch <- errmsg{ErrListen, fmt.Sprintf("can't listen, %s", err.Error())}
		return
	}
	//	sock.SetOption(mangos.OptionRecvDeadline, timeout)
	//	sock.SetOption(mangos.OptionSendDeadline, timeout)
	// server loop
	for {
		msg, err = sock.Recv()
		if err != nil {
			ch <- errmsg{ErrRecv, fmt.Sprintf("i/o error, %s", err.Error())}
		}
		cmd := minidb.Command{}
		err := json.Unmarshal(msg, &cmd)
		if err != nil {
			ch <- errmsg{ErrUnmarshal, fmt.Sprintf("unmarshal command failed, %s", err.Error())}
		}
		reply := minidb.Exec(&cmd)
		msg, err = json.Marshal(reply)
		if err != nil {
			ch <- errmsg{ErrMarshal, fmt.Sprintf("marshal reply failed, %s", err.Error())}
		}
		err = sock.Send(msg)
		if err != nil {
			ch <- errmsg{ErrSendIO, fmt.Sprintf("can't send reply, %s", err.Error())}
		}
		select {
		case <-ctx.Done():
			sock.Close()
			minidb.CloseAllDBs()
		default:
			// do nothing
		}
	}
}

func main() {
	// parse the command line

	app := kingpin.New("mdbserve", "Minidb command line server tool.")
	//	debug := app.Flag("debug", "Enable debug mode.").Bool()
	timeout := app.Command("timeout", "Specify how long the server process is kept alive.")
	timeoutValue := timeout.Arg("value", "The timeout value in seconds, or 'none' to keep running until a ServerQuit command is received.").Required().String()
	url := app.Flag("url", "A custom url to listen to. If this is not provided, tcp//localhost:7873 is used.").String()

	command := kingpin.MustParse(app.Parse(os.Args[1:]))

	var tmax int
	var err error
	switch command {
	case timeout.FullCommand():
		if strings.ToLower(*timeoutValue) == "none" {
			tmax = 0
		} else {
			tmax, err = strconv.Atoi(*timeoutValue)
			if err != nil || tmax < 0 {
				fmt.Fprintf(os.Stderr, "syntax error: timeout value must be a positive number!\n")
				os.Exit(ErrSyntaxError)
			}
		}
	}
	theURL := ""
	if url == nil || *url == "" {
		theURL = "tcp://0.0.0.0:7873"
	} else {
		theURL = *url
	}

	// Start the server loop

	// context for timeout handling of server loop in main and serverloop
	var ctx, cancel = context.WithCancel(context.Background())
	var ch = make(chan errmsg, 1)
	var msg errmsg

	go serverLoop(theURL, ctx, ch, time.Duration(tmax)*time.Second)
	defer cancel()

	done := false
	for done == false {
		select {
		case msg := <-ch:
			fmt.Fprintf(os.Stderr, "%s", msg.msg)
			done = true
		case <-time.After(time.Duration(tmax) * time.Second):
			if tmax > 0 {
				done = true
			}
		}
	}
	cancel()
	os.Exit(msg.number)
}
