// The minidb command line tool

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	minidb "github.com/rasteric/minidb"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"nanomsg.org/go/mangos/v2"
	"nanomsg.org/go/mangos/v2/protocol/rep"

	// register transports
	_ "nanomsg.org/go/mangos/v2/transport/all"
)

const (
	ErrNone int = iota + 1
	ErrSyntaxError
	ErrServerFail
)

func ServerLoop(timeout int, url string, forever bool) error {
	var sock mangos.Socket
	var err error
	var msg []byte
	if sock, err = rep.NewSocket(); err != nil {
		return minidb.Fail("can't get new socket, %s", err)
	}
	if err = sock.Listen(url); err != nil {
		return minidb.Fail("can't listen, %s", err.Error())
	}
	for {
		msg, err = sock.Recv()
		if err != nil {
			return minidb.Fail("i/o error, %s", err.Error())
		}
		cmd := minidb.Command{}
		err := json.Unmarshal(msg, &cmd)
		if err != nil {
			return minidb.Fail("unmarshal command failed, %s", err.Error())
		}
		reply := minidb.Exec(&cmd)
		msg, err = json.Marshal(reply)
		if err != nil {
			return minidb.Fail("marshal reply failed, %s", err.Error())
		}
		err = sock.Send(msg)
		if err != nil {
			return minidb.Fail("can't send reply, %s", err.Error())
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
	var noTimeout bool
	var err error
	switch command {
	case timeout.FullCommand():
		if strings.ToLower(*timeoutValue) == "none" {
			noTimeout = true
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
	err = ServerLoop(tmax, theURL, noTimeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(ErrServerFail)
	}
}
