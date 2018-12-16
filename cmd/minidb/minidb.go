// The minidb command line tool

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	minidb "github.com/rasteric/minidb"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	mangos "nanomsg.org/go/mangos/v2"
	"nanomsg.org/go/mangos/v2/protocol/req"

	// register transports
	_ "nanomsg.org/go/mangos/v2/transport/all"
)

const (
	ErrNone = iota + 1
	ErrCannotOpenDB
	ErrInvalidFields
	ErrFailedClose
	ErrCannotAddTable
	ErrCannotCreateItem
	ErrNotFound
	ErrSetFailed
	ErrCountFailed
	ErrListFailed
	ErrFailedListFields
	ErrSetTypeError
	ErrSearchFail
	ErrSyntaxError
	ErrNoServerExecutable
	ErrCannotStartServerExecutable
	ErrNoSocket
	ErrNoConnection
	ErrIO
)

func sendCommand(sock mangos.Socket, cmd *minidb.Command) (*minidb.Result, error) {
	msg, err := json.Marshal(&cmd)
	if err != nil {
		return nil, err
	}
	err = sock.Send(msg)
	if err != nil {
		return nil, err
	}
	if msg, err = sock.Recv(); err != nil {
		return nil, err
	}
	reply := minidb.Result{}
	err = json.Unmarshal(msg, &reply)
	if err != nil {
		return nil, err
	}
	if reply.HasError {
		return nil, errors.New(reply.Str)
	}
	return &reply, nil
}

func printItems(items []minidb.Item) {
	if len(items) == 0 {
		return
	}
	s := ""
	for i, item := range items {
		if i > 0 {
			s += " "
		}
		s += fmt.Sprintf("%d", item)
	}
	fmt.Printf("%s\n", s)
}

func die(errCode int, msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR "+msg, args...)
	os.Exit(errCode)
}

func main() {
	// parse the command line

	app := kingpin.New("minidb", "A minimalist command line database.")
	//	debug := app.Flag("debug", "Enable debug mode.").Bool()
	dbfile := app.Flag("db", "Database file.").Default("db.sqlite").String()
	dbview := app.Flag("view", "Specify how to display results: plain - one per line, titled - with eader information").Enum("plain", "titled")

	table := app.Command("table", "Create a new table.")
	tableName := table.Arg("name", "Name of the table.").Required().String()
	tableFields := table.Arg("fields", "Fields of the table.").Required().Strings()

	new := app.Command("new", "Create new item in a table and return its ID.")
	newTable := new.Arg("table", "Table in which a new item is created.").Required().String()

	get := app.Command("get", "Get a field value for an item. (Returns a list in case of a list field.)")
	getTable := get.Arg("table", "The table to which the item belongs.").Required().String()
	getItem := get.Arg("item", "The item to query.").Required().Int64()
	getFields := get.Arg("fields", "The fields to query.").Strings()

	set := app.Command("set", "Set a field's value for an item. A list of values can be provided for a list field.")
	setTable := set.Arg("table", "The table to which the item belongs.").Required().String()
	setItem := set.Arg("item", "The item whose field to set.").Required().Int64()
	setField := set.Arg("field", "The field whose value to set.").Required().String()
	setValues := set.Arg("values", "The value to set (values in case of a list field).").Required().Strings()

	count := app.Command("count", "Count the number of items in a table.")
	countTable := count.Arg("table", "The table whose items are to be counted.").Required().String()

	list := app.Command("list", "List all items in a table.")
	listTable := list.Arg("table", "The table whose items to list.").Required().String()
	listLimit := list.Arg("limit", "The maximum list size (omit=no limit)").Int64()

	listFields := app.Command("list-fields", "List all fields of a table.")
	listFieldsTable := listFields.Arg("table", "The table whose fields to list.").Required().String()

	listTables := app.Command("list-tables", "List all tables in the database.")

	find := app.Command("find", "Find elements matching a query.")
	findQuery := find.Arg("query", `A query starts with the table and then contains a logical combination of Fieldname=Query clauses.`).Required().Strings()
	findEscape := app.Flag("escape", "The escape character for find queries.").String()
	findLimit := app.Flag("limit", "The maximum number of items to return (omit=no limit).").Int64()

	serverTimeout := app.Flag("keep-up", "Time to keep the database server running before it needs to be restarted. Use 'forever' to keep it running.").String()
	serverExecutable := app.Flag("server", "Path to the minidb-server executable.").String()
	serverURL := app.Flag("connection", "Mangos-compatible transport URL to connect to the server executable. If this is not provided, tcp://localhost:7873 is used.").String()
	serverConnectTrials := app.Flag("connection-trials", "Number of times minidb tries to connect to the database server process before it gives up.").Int32()

	command := kingpin.MustParse(app.Parse(os.Args[1:]))

	var keepUp int = 600
	var noTimeout bool
	if strings.ToLower(*serverTimeout) == "forever" {
		noTimeout = true
	}
	timeout, err := strconv.Atoi(*serverTimeout)
	if err == nil {
		if timeout <= 0 {
			die(ErrSyntaxError, "keep-up seconds need to be a positive integer or 'forever'.\n")
		}
		keepUp = timeout
	}
	var connectTrials int32 = 10
	if *serverConnectTrials > 0 {
		connectTrials = *serverConnectTrials
	}

	// fix dbfile
	if *dbfile == "" {
		*dbfile = "db.sqlite"
	}

	// run the server executable if needed
	_, err = FindProcessByName("mdbserve")
	if err != nil {
		if *serverExecutable == "" {
			*serverExecutable = "../mdbserve/mdbserve"
			if _, err := os.Stat(*serverExecutable); os.IsNotExist(err) {
				*serverExecutable, err = exec.LookPath("mdbserve")
				if err != nil {
					die(ErrNoServerExecutable, "cannot find server executable!\n")
				}
			}
		}
		var timeoutArg string
		if noTimeout {
			timeoutArg = "none"
		} else {
			timeoutArg = strconv.Itoa(keepUp)
		}
		cmd := exec.Command(*serverExecutable, "timeout", timeoutArg)
		if err := cmd.Start(); err != nil {
			die(ErrCannotStartServerExecutable, "cannot start server executable: %s.\n", err)
		}
	}

	// open the connection
	var sock mangos.Socket
	if sock, err = req.NewSocket(); err != nil {
		die(ErrNoSocket, "cannot get a socket to connect to server: %s.\n", err)
	}
	defer sock.Close()
	if *serverURL == "" {
		*serverURL = "tcp://localhost:7873"
	}
	// we try dialing several times before giving up
	var c int32 = 0
	success := false
	for c < connectTrials {
		if err = sock.Dial(*serverURL); err == nil {
			success = true
			break
		}
		c++
		time.Sleep(10 * time.Millisecond)
	}
	if !success {
		die(ErrNoConnection, "cannot connect to server executable: %s.\n", err)
	}
	// connection established, now send the open command
	var result *minidb.Result
	if result, err = sendCommand(sock, minidb.OpenCommand("sqlite3", *dbfile)); err != nil {
		die(ErrIO, "could not open database: %s\n", err)
	}
	theDB := minidb.CommandDB(*dbfile)

	switch command {
	case table.FullCommand():
		fields, err := minidb.ParseFieldDesc(*tableFields)
		if err != nil {
			die(ErrInvalidFields, "invalid table or field descriptions - %s.\n", err)
		}
		result, err = sendCommand(sock, minidb.AddTableCommand(theDB, *tableName, fields))
		if err != nil {
			die(ErrCannotAddTable, "unable to create table - %s.\n", err)
		}
	case new.FullCommand():
		result, err := sendCommand(sock, minidb.NewItemCommand(theDB, *newTable))
		if err != nil {
			die(ErrCannotCreateItem, "unable to create item - %s.\n", err)
		}
		fmt.Printf("%d\n", result.Items[0])
	case get.FullCommand():
		if len(*getFields) == 0 {
			result, err := sendCommand(sock, minidb.GetFieldsCommand(theDB, *getTable))
			if err != nil {
				die(ErrIO, "cannot get fields: %s.\n", err)
			}
			fields := result.Fields
			fieldNames := make([]string, 0)
			for i, _ := range fields {
				fieldNames = append(fieldNames, (fields)[i].Name)
			}
			getFields = &fieldNames
		}
		errCount := 0
		for i, _ := range *getFields {
			result, err := sendCommand(sock, minidb.GetCommand(theDB, *getTable, minidb.Item(*getItem), (*getFields)[i]))
			if err != nil {
				if *dbview == "titled" {
					fmt.Printf("%s %d %s: 0\n", *getTable, *getItem, (*getFields)[i])
				}
				fmt.Fprintf(os.Stderr, "not found - %s.\n", err)
				errCount++
			} else {
				if *dbview == "titled" {
					fmt.Printf("%s %d %s: %d\n", *getTable, *getItem, (*getFields)[i], len(result.Values))
				}
				for j, _ := range result.Values {
					fmt.Printf("%s\n", result.Values[j].String())
				}
			}
		}
		if errCount > 0 {
			os.Exit(ErrNotFound)
		}
	case set.FullCommand():
		result, err := sendCommand(sock, minidb.ParseFieldValuesCommand(theDB, *setTable, *setField, *setValues))
		if err != nil {
			die(ErrSetTypeError, "set failed - %s\n", err)
		}
		_, err = sendCommand(sock,
			minidb.SetCommand(theDB, *setTable, minidb.Item(*setItem), *setField, result.Values))
		if err != nil {
			die(ErrSetFailed, "set failed - %s\n", err)
		}
	case count.FullCommand():
		result, err := sendCommand(sock, minidb.CountCommand(theDB, *countTable))
		if err != nil {
			die(ErrCountFailed, "%s\n", err)
		}
		fmt.Printf("%d\n", result.Int)
	case list.FullCommand():
		result, err := sendCommand(sock, minidb.ListItemsCommand(theDB, *listTable, *listLimit))
		if err != nil {
			die(ErrListFailed, "%s\n", err)
		}
		printItems(result.Items)
	case listFields.FullCommand():
		result, err := sendCommand(sock, minidb.GetFieldsCommand(theDB, *listFieldsTable))
		if err != nil {
			die(ErrFailedListFields, "cannot list fields for '%s' - %s.\n", *listFieldsTable, err)
		}
		for i, _ := range result.Fields {
			fmt.Printf("%s %s\n", minidb.GetUserTypeString(result.Fields[i].Sort), result.Fields[i].Name)
		}
	case listTables.FullCommand():
		if result, err = sendCommand(sock, minidb.GetTablesCommand(theDB)); err != nil {
			die(ErrIO, "transport failed: %s\n", err)
		}
		tables := result.Strings
		for i, _ := range tables {
			if i > 0 {
				fmt.Print(" ")
			}
			fmt.Print(tables[i])
		}
		fmt.Printf("\n")
	case find.FullCommand():
		_ = findEscape
		s := strings.Join(*findQuery, " ")
		query, err := minidb.ParseQuery(s)
		if err != nil {
			die(ErrSyntaxError, "syntax error - %s.\n", err)
		}
		result, err := sendCommand(sock, minidb.FindCommand(theDB, query, *findLimit))
		if err != nil {
			die(ErrSearchFail, "search Search '%s' failed - %s.\n", *findQuery, err)
		}
		printItems(result.Items)
	}
}
