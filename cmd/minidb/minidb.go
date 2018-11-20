// The minidb command line tool

package minidb

import (
	"fmt"
	"os"
	"strings"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const (
	OK = iota + 1
	ERR_CannotOpenDB
	ERR_InvalidFields
	ERR_FailedClose
	ERR_CannotAddTable
	ERR_CannotCreateItem
	ERR_NotFound
	ERR_SetFailed
	ERR_CountFailed
	ERR_ListFailed
	ERR_FailedListFields
	ERR_SetTypeError
	ERR_SearchFail
	ERR_SyntaxError
)

func printItems(items []Item) {
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

	command := kingpin.MustParse(app.Parse(os.Args[1:]))

	// open the db
	if *dbfile == "" {
		*dbfile = "db.sqlite"
	}
	db, err := Open("sqlite3", *dbfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR Unable to open database '%s' - %s.\n", *dbfile, err)
		os.Exit(ERR_CannotOpenDB)
	}
	defer db.Close()

	switch command {
	case table.FullCommand():
		fields, err := ParseFieldDesc(*tableFields)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR Invalid table or field descriptions - %s.\n", err)
			os.Exit(ERR_InvalidFields)
		}
		err = db.AddTable(*tableName, fields)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR Unable to create table - %s.\n", err)
			os.Exit(ERR_CannotAddTable)
		}
	case new.FullCommand():
		id, err := db.NewItem(*newTable)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR Unable to create item - %s.\n", err)
			os.Exit(ERR_CannotCreateItem)
		}
		fmt.Printf("%d\n", id)
	case get.FullCommand():
		if len(*getFields) == 0 {
			fields, _ := db.GetFields(*getTable)
			fieldNames := make([]string, 0)
			for i, _ := range fields {
				fieldNames = append(fieldNames, (fields)[i].Name)
			}
			getFields = &fieldNames
		}
		errCount := 0
		for i, _ := range *getFields {
			data, err := db.Get(*getTable, Item(*getItem), (*getFields)[i])
			if err != nil {
				if *dbview == "titled" {
					fmt.Printf("%s %d %s: 0\n", *getTable, *getItem, (*getFields)[i])
				}
				fmt.Fprintf(os.Stderr, "ERROR Not found - %s.\n", err)
				errCount++
			} else {
				if *dbview == "titled" {
					fmt.Printf("%s %d %s: %d\n", *getTable, *getItem, (*getFields)[i], len(data))
				}
				for j, _ := range data {
					fmt.Printf("%s\n", data[j].String())
				}
			}
		}
		if errCount > 0 {
			os.Exit(ERR_NotFound)
		}
	case set.FullCommand():
		data, err := db.ParseFieldValues(*setTable, *setField, *setValues)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR Set failed - %s\n", err)
			os.Exit(ERR_SetTypeError)
		}
		err = db.Set(*setTable, Item(*setItem), *setField, data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR Set failed - %s\n", err)
			os.Exit(ERR_SetFailed)
		}
	case count.FullCommand():
		c, err := db.Count(*countTable)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR - %s\n", err)
			os.Exit(ERR_CountFailed)
		}
		fmt.Printf("%d\n", c)
	case list.FullCommand():
		results, err := db.ListItems(*listTable, *listLimit)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR - %s\n", err)
			os.Exit(ERR_ListFailed)
		}
		printItems(results)
	case listFields.FullCommand():
		fields, err := db.GetFields(*listFieldsTable)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR Cannot list fields for '%s' - %s.\n", *listFieldsTable, err)
			os.Exit(ERR_FailedListFields)
		}
		for i, _ := range fields {
			fmt.Printf("%s %s\n", getUserTypeString(fields[i].Sort), fields[i].Name)
		}
	case listTables.FullCommand():
		tables := db.GetTables()
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
		query, err := ParseQuery(s)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR Syntax error - %s.\n", err)
			os.Exit(ERR_SyntaxError)
		}
		found, err := db.Find(query, *findLimit)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR Search '%s' failed - %s.\n", *findQuery, err)
			os.Exit(ERR_SearchFail)
		}
		printItems(found)
	}
}
