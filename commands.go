package minidb

import "sync"

// Represents a command in Exec().
type CommandID int

// The actual Exec() CommandID values. Names mirror the respective functions.
const (
	CmdOpen CommandID = iota + 1
	CmdAddTable
	CmdClose
	CmdCount
	CmdFind
	CmdGet
	CmdGetTables
	CmdIsListField
	CmdItemExists
	CmdListItems
	CmdNewItem
	CmdParseFieldValues
	CmdSet
	CmdTableExists
	CmdToSQL
	CmdFieldIsNull
	CmdFieldExists
	CmdGetFields
	CmdIsEmptyListField
	CmdMustGetFieldType
)

// A database that has been opened.
type CommandDB string

// Command structures contains all information needed to execute an arbitrary command.
// Use Exec() to execute a command and get the result.
// Every function <Name> in minidb has a corresponding function <Name>Command that
// returns the corresponding command. Consult the API for <Name> for help, as the input parameters are exactly the same,
// except that the first argument is a database path of type CommandDB.
// You should never use Command structures directly, but use the provided wrapper functions for strong typing.
// Result structures have the HasError field set to true if an error has occurred.
// Commands and results can be serialized to json.
type Command struct {
	ID        CommandID `json:"id"`
	DB        CommandDB `json:"dbid"`
	StrArgs   []string  `json:"strings"`
	ItemArg   Item      `json:"item"`
	FieldArgs []Field   `json:"fields"`
	ValueArgs []Value   `json:"values"`
	QueryArg  Query     `json:"query"`
	IntArg    int64     `json:"int"`
}

// Result is a structure representing the result of a command execution via Exec().
// If an error has occurred, then HasError is true and the Int and S fields contain
// the numeric error code and the error message string. Otherwise the respective fields
// are filled in, as corresponding to the return value(s) of the respective function call.
type Result struct {
	Str      string   `json:"str"`
	Strings  []string `json:"strings"`
	Int      int64    `json:"int64"`
	Bool     bool     `json:"bool"`
	Items    []Item   `json:"items"`
	Values   []Value  `json:"values"`
	Fields   []Field  `json:"fields"`
	HasError bool     `json:"iserror"`
}

var openDBs map[CommandDB]*MDB
var connections map[CommandDB]int

var mutex sync.RWMutex

// Numeric error codes returned by Exec() in a Result structure's Int field.
const (
	NoErr int64 = iota + 1
	ErrCannotOpen
	ErrUnknownDB
	ErrUnknownCommand
	ErrAddTableFailed
	ErrClosingDB
	ErrCountFailed
	ErrFindFailed
	ErrGetFailed
	ErrGetTablesFailed
	ErrListItemsFailed
	ErrNewItemFailed
	ErrParseFieldValuesFailed
	ErrSetFailed
	ErrToSQLFailed
	ErrFieldExistsFailed
	ErrGetFieldsFailed
)

func getDB(cmd *Command) (*MDB, *Result) {
	mutex.RLock()
	defer mutex.RUnlock()
	theDB, ok := openDBs[cmd.DB]
	if ok {
		return theDB, nil
	}
	r := Result{HasError: true, Int: ErrUnknownDB}
	r.Str = Fail("exec failed: db '%s' unknown", cmd.DB).Error()
	return nil, &r
}

func CloseAllDBs() {
	mutex.Lock()
	defer mutex.Unlock()
	for db, _ := range openDBs {
		openDBs[db].Close()
		connections[db] = 0
		openDBs[db] = nil
	}
}

// Exec takes a Command structure and executes it, returning a Result or an error.
// This function is a large switch, as a wrapper around the more specific API functions.
// It incurs a runtime penalty and should only used when needed (e.g. when commands
// have to be marshalled and unmarshalled).
func Exec(cmd *Command) *Result {
	var r Result
	var theDB *MDB
	var err error
	var errResult *Result

	if openDBs == nil {
		openDBs = make(map[CommandDB]*MDB)
	}
	if connections == nil {
		connections = make(map[CommandDB]int)
	}

	if cmd.ID == CmdOpen {
		mutex.Lock()
		defer mutex.Unlock()
		if _, ok := openDBs[CommandDB(cmd.StrArgs[1])]; ok {
			connections[CommandDB(cmd.StrArgs[1])] += 1
		} else {
			theDB, err := Open(cmd.StrArgs[0], cmd.StrArgs[1])
			if err != nil {
				r.HasError = true
				r.Int = ErrCannotOpen
				r.Str = err.Error()
				return &r
			}
			openDBs[CommandDB(cmd.StrArgs[1])] = theDB
			connections[CommandDB(cmd.StrArgs[1])] = 1
		}
		return &r
	}

	if theDB, errResult = getDB(cmd); errResult != nil {
		return errResult
	}

	switch cmd.ID {
	case CmdAddTable:
		err = theDB.AddTable(cmd.StrArgs[0], cmd.FieldArgs)
		if err != nil {
			r.HasError = true
			r.Int = ErrAddTableFailed
			r.Str = err.Error()
		}
	case CmdClose:
		err = nil
		mutex.Lock()
		if connections[cmd.DB] == 1 {
			err = theDB.Close()
			delete(openDBs, cmd.DB)
			delete(connections, cmd.DB)
		} else {
			connections[cmd.DB] -= 1
		}
		mutex.Unlock()
		if err != nil {
			r.HasError = true
			r.Int = ErrClosingDB
			r.Str = err.Error()
		}
	case CmdCount:
		r.Int, err = theDB.Count(cmd.StrArgs[0])
		if err != nil {
			r.HasError = true
			r.Int = ErrCountFailed
			r.Str = err.Error()
		}
	case CmdFind:
		r.Items, err = theDB.Find(&(cmd.QueryArg), cmd.IntArg)
		if err != nil {
			r.HasError = true
			r.Int = ErrFindFailed
			r.Str = err.Error()
		}
	case CmdGet:
		r.Values, err = theDB.Get(cmd.StrArgs[0], cmd.ItemArg, cmd.StrArgs[1])
		if err != nil {
			r.HasError = true
			r.Int = ErrGetFailed
			r.Str = err.Error()
		}
	case CmdGetTables:
		r.Strings = theDB.GetTables()
	case CmdIsListField:
		r.Bool = theDB.IsListField(cmd.StrArgs[0], cmd.StrArgs[1])
	case CmdItemExists:
		r.Bool = theDB.ItemExists(cmd.StrArgs[0], cmd.ItemArg)
	case CmdListItems:
		r.Items, err = theDB.ListItems(cmd.StrArgs[0], cmd.IntArg)
		if err != nil {
			r.HasError = true
			r.Int = ErrListItemsFailed
			r.Str = err.Error()
		}
	case CmdNewItem:
		item, err := theDB.NewItem(cmd.StrArgs[0])
		if err != nil {
			r.HasError = true
			r.Int = ErrNewItemFailed
			r.Str = err.Error()
			return &r
		}
		r.Items = make([]Item, 1)
		r.Items[0] = item
	case CmdParseFieldValues:
		r.Values, err = theDB.ParseFieldValues(cmd.StrArgs[0], cmd.StrArgs[1], cmd.StrArgs[2:])
		if err != nil {
			r.HasError = true
			r.Int = ErrParseFieldValuesFailed
			r.Str = err.Error()
		}
	case CmdSet:
		err = theDB.Set(cmd.StrArgs[0], cmd.ItemArg, cmd.StrArgs[1], cmd.ValueArgs)
		if err != nil {
			r.HasError = true
			r.Int = ErrSetFailed
			r.Str = err.Error()
		}
	case CmdTableExists:
		r.Bool = theDB.TableExists(cmd.StrArgs[0])
	case CmdToSQL:
		s, err := theDB.ToSql(cmd.StrArgs[0], &cmd.QueryArg, cmd.IntArg)
		if err != nil {
			r.HasError = true
			r.Int = ErrToSQLFailed
			r.Str = err.Error()
			return &r
		}
		r.Strings = make([]string, 1)
		r.Strings[0] = s
	case CmdFieldIsNull:
		r.Bool = theDB.FieldIsNull(cmd.StrArgs[0], cmd.ItemArg, cmd.StrArgs[1])
	case CmdFieldExists:
		r.Bool = theDB.FieldExists(cmd.StrArgs[0], cmd.StrArgs[1])
	case CmdGetFields:
		r.Fields, err = theDB.GetFields(cmd.StrArgs[0])
		if err != nil {
			r.HasError = true
			r.Int = ErrGetFieldsFailed
			r.Str = err.Error()
		}
	case CmdIsEmptyListField:
		r.Bool = theDB.IsEmptyListField(cmd.StrArgs[0], cmd.ItemArg, cmd.StrArgs[1])
	case CmdMustGetFieldType:
		r.Int = int64(theDB.MustGetFieldType(cmd.StrArgs[0], cmd.StrArgs[1]))
	default:
		r.HasError = true
		r.Str = Fail("exec failed: unhandled command").Error()
	}
	return &r
}

func OpenCommand(driver string, file string) *Command {
	return &Command{
		ID:      CmdOpen,
		StrArgs: []string{driver, file},
	}
}

func AddTableCommand(db CommandDB, table string, fields []Field) *Command {
	return &Command{
		ID:        CmdAddTable,
		DB:        db,
		StrArgs:   []string{table},
		FieldArgs: fields,
	}
}

func CloseCommand(db CommandDB) *Command {
	return &Command{
		ID: CmdClose,
		DB: db,
	}
}

func CountCommand(db CommandDB, table string) *Command {
	return &Command{
		ID:      CmdCount,
		DB:      db,
		StrArgs: []string{table},
	}
}

func FieldExistsCommand(db CommandDB, table string, field string) *Command {
	return &Command{
		ID:      CmdFieldExists,
		DB:      db,
		StrArgs: []string{table, field},
	}
}

func FieldIsNullCommand(db CommandDB, item Item, field string) *Command {
	return &Command{
		ID:      CmdFieldIsNull,
		DB:      db,
		StrArgs: []string{field},
		ItemArg: item,
	}
}

func FindCommand(db CommandDB, query *Query, limit int64) *Command {
	return &Command{
		ID:       CmdFind,
		DB:       db,
		QueryArg: *query,
		IntArg:   limit,
	}
}
func GetCommand(db CommandDB, table string, item Item, field string) *Command {
	return &Command{
		ID:      CmdGet,
		DB:      db,
		StrArgs: []string{table, field},
		ItemArg: item,
	}
}

func GetFieldsCommand(db CommandDB, table string) *Command {
	return &Command{
		ID:      CmdGetFields,
		DB:      db,
		StrArgs: []string{table},
	}
}

func GetTablesCommand(db CommandDB) *Command {
	return &Command{
		ID: CmdGetTables,
		DB: db,
	}
}

func IsEmptyListFieldCommand(db CommandDB, table string, item Item, field string) *Command {
	return &Command{
		ID:      CmdIsEmptyListField,
		DB:      db,
		StrArgs: []string{table, field},
		ItemArg: item,
	}
}

func IsListFieldCommand(db CommandDB, table string, field string) *Command {
	return &Command{
		ID:      CmdIsListField,
		DB:      db,
		StrArgs: []string{table, field},
	}
}

func ItemExistsCommand(db CommandDB, table string, item Item) *Command {
	return &Command{
		ID:      CmdItemExists,
		DB:      db,
		StrArgs: []string{table},
		ItemArg: item,
	}
}

func ListItemsCommand(db CommandDB, table string, limit int64) *Command {
	return &Command{
		ID:      CmdListItems,
		DB:      db,
		StrArgs: []string{table},
		IntArg:  limit,
	}
}
func MustGetFieldTypeCommand(db CommandDB, table string, field string) *Command {
	return &Command{
		ID:      CmdMustGetFieldType,
		DB:      db,
		StrArgs: []string{table, field},
	}
}

func NewItemCommand(db CommandDB, table string) *Command {
	return &Command{
		ID:      CmdNewItem,
		DB:      db,
		StrArgs: []string{table},
	}
}

func ParseFieldValuesCommand(db CommandDB, table string, field string, data []string) *Command {
	cmd := Command{
		ID: CmdParseFieldValues,
		DB: db,
	}
	cmd.StrArgs = make([]string, len(data)+2)
	cmd.StrArgs[0] = table
	cmd.StrArgs[1] = field
	for i := 0; i < len(data); i++ {
		cmd.StrArgs[i+2] = data[i]
	}
	return &cmd
}

func SetCommand(db CommandDB, table string, item Item, field string, data []Value) *Command {
	return &Command{
		ID:        CmdSet,
		DB:        db,
		StrArgs:   []string{table, field},
		ItemArg:   item,
		ValueArgs: data,
	}
}

func TableExistsCommand(db CommandDB, table string) *Command {
	return &Command{
		ID:      CmdTableExists,
		DB:      db,
		StrArgs: []string{table},
	}
}

func ToSqlCommand(db CommandDB, table string, query *Query, limit int64) *Command {
	return &Command{
		ID:       CmdToSQL,
		DB:       db,
		StrArgs:  []string{table},
		QueryArg: *query,
		IntArg:   limit,
	}
}
