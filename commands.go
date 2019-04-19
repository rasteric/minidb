package minidb

import (
	"sync"
	"time"
)

// CommandID represents the type of a command in the Exec() function.
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
	CmdGetInt
	CmdGetStr
	CmdGetBlob
	CmdGetDate
	CmdSetInt
	CmdSetStr
	CmdSetBlob
	CmdSetDate
	CmdDeleteInt
	CmdDeleteStr
	CmdDeleteBlob
	CmdDeleteDate
	CmdHasInt
	CmdHasStr
	CmdHasBlob
	CmdHasDate
	CmdListInt
	CmdListStr
	CmdListBlob
	CmdListDate
	CmdSetDateStr
	CmdFieldIsEmpty
	CmdBackup
	CmdRemoveItem
)

// CommandDB is the database that has been opened.
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
	IntArg2   int64     `json:"int2"`
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
	Bytes    []byte   `json:"binary"`
	Ints     []int64  `json:"ints"`
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
	ErrInvalidDate
	ErrBackupFailed
	ErrRemoveItemFailed
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

// CloseAllDBs closes all open DB connections and cleans up resources.
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
	case CmdBackup:
		err = theDB.Backup(cmd.StrArgs[0])
		if err != nil {
			r.HasError = true
			r.Int = ErrBackupFailed
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
	case CmdRemoveItem:
		err = theDB.RemoveItem(cmd.StrArgs[0], cmd.ItemArg)
		if err != nil {
			r.HasError = true
			r.Int = ErrRemoveItemFailed
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
	case CmdFieldIsEmpty:
		r.Bool = theDB.FieldIsEmpty(cmd.StrArgs[0], cmd.ItemArg, cmd.StrArgs[1])
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
	case CmdGetInt:
		r.Int = theDB.GetInt(cmd.IntArg)
	case CmdGetStr:
		r.Str = theDB.GetStr(cmd.IntArg)
	case CmdGetBlob:
		r.Bytes = theDB.GetBlob(cmd.IntArg)
	case CmdGetDate:
		r.Str = theDB.GetDateStr(cmd.IntArg)
	case CmdSetInt:
		theDB.SetInt(cmd.IntArg, cmd.IntArg2)
	case CmdSetStr:
		theDB.SetStr(cmd.IntArg, cmd.StrArgs[0])
	case CmdSetBlob:
		theDB.SetBlob(cmd.IntArg, []byte(cmd.StrArgs[0]))
	case CmdSetDate:
		t, err := ParseTime(cmd.StrArgs[0])
		if err != nil {
			r.HasError = true
			r.Int = ErrInvalidDate
			r.Str = err.Error()
		} else {
			theDB.SetDate(cmd.IntArg, t)
		}
	case CmdSetDateStr:
		theDB.SetDateStr(cmd.IntArg, cmd.StrArgs[0])
	case CmdHasInt:
		r.Bool = theDB.HasInt(cmd.IntArg)
	case CmdHasStr:
		r.Bool = theDB.HasStr(cmd.IntArg)
	case CmdHasBlob:
		r.Bool = theDB.HasBlob(cmd.IntArg)
	case CmdHasDate:
		r.Bool = theDB.HasDate(cmd.IntArg)
	case CmdDeleteInt:
		theDB.DeleteInt(cmd.IntArg)
	case CmdDeleteStr:
		theDB.DeleteStr(cmd.IntArg)
	case CmdDeleteBlob:
		theDB.DeleteBlob(cmd.IntArg)
	case CmdDeleteDate:
		theDB.DeleteDate(cmd.IntArg)
	case CmdListInt:
		r.Ints = theDB.ListInt()
	case CmdListStr:
		r.Ints = theDB.ListStr()
	case CmdListBlob:
		r.Ints = theDB.ListBlob()
	case CmdListDate:
		r.Ints = theDB.ListDate()
	default:
		r.HasError = true
		r.Str = Fail("exec failed: unhandled command").Error()
	}
	return &r
}

// OpenCommand returns a pointer to a command structure for mdb.Open().
func OpenCommand(driver string, file string) *Command {
	return &Command{
		ID:      CmdOpen,
		StrArgs: []string{driver, file},
	}
}

// AddTableCommand returns a pointer to a command structure for mdb.AddTable().
func AddTableCommand(db CommandDB, table string, fields []Field) *Command {
	return &Command{
		ID:        CmdAddTable,
		DB:        db,
		StrArgs:   []string{table},
		FieldArgs: fields,
	}
}

// CloseCommand returns a pointer to a command structure for mdb.Close().
func CloseCommand(db CommandDB) *Command {
	return &Command{
		ID: CmdClose,
		DB: db,
	}
}

// CountCommand returns a pointer to a command structure for mdb.Count()
func CountCommand(db CommandDB, table string) *Command {
	return &Command{
		ID:      CmdCount,
		DB:      db,
		StrArgs: []string{table},
	}
}

// FieldExistsCommand returns a pointer to a command structure for mdb.FieldExists().
func FieldExistsCommand(db CommandDB, table string, field string) *Command {
	return &Command{
		ID:      CmdFieldExists,
		DB:      db,
		StrArgs: []string{table, field},
	}
}

// FieldIsNullCommand returns a pointer to a command structure for mdb.FieldIsNull().
func FieldIsNullCommand(db CommandDB, item Item, field string) *Command {
	return &Command{
		ID:      CmdFieldIsNull,
		DB:      db,
		StrArgs: []string{field},
		ItemArg: item,
	}
}

// FieldIsEmptyCommand returns a pointer to a command structure for mdb.FieldIsEmpty().
func FieldIsEmptyCommand(db CommandDB, item Item, field string) *Command {
	return &Command{
		ID:      CmdFieldIsEmpty,
		DB:      db,
		StrArgs: []string{field},
		ItemArg: item,
	}
}

// FindCommand returns a pointer to a command structure for mdb.Find().
func FindCommand(db CommandDB, query *Query, limit int64) *Command {
	return &Command{
		ID:       CmdFind,
		DB:       db,
		QueryArg: *query,
		IntArg:   limit,
	}
}

// GetCommand returns a pointer to a command structure for mdb.Get().
func GetCommand(db CommandDB, table string, item Item, field string) *Command {
	return &Command{
		ID:      CmdGet,
		DB:      db,
		StrArgs: []string{table, field},
		ItemArg: item,
	}
}

// GetFieldsCommand returns a pointer to a command structure for mdb.GetFields().
func GetFieldsCommand(db CommandDB, table string) *Command {
	return &Command{
		ID:      CmdGetFields,
		DB:      db,
		StrArgs: []string{table},
	}
}

// GetTablesCommand returns a pointer to a command structure for mdb.GetTables().
func GetTablesCommand(db CommandDB) *Command {
	return &Command{
		ID: CmdGetTables,
		DB: db,
	}
}

// IsEmptyListFieldCommand returns a pointer to a command structure for mdb.IsEmptyListField().
func IsEmptyListFieldCommand(db CommandDB, table string, item Item, field string) *Command {
	return &Command{
		ID:      CmdIsEmptyListField,
		DB:      db,
		StrArgs: []string{table, field},
		ItemArg: item,
	}
}

// IsListFieldCommand returns a pointer to a command structure for mdb.IsListField().
func IsListFieldCommand(db CommandDB, table string, field string) *Command {
	return &Command{
		ID:      CmdIsListField,
		DB:      db,
		StrArgs: []string{table, field},
	}
}

// ItemExistsCommand returns a pointer to a command structure for mdb.ItemExists().
func ItemExistsCommand(db CommandDB, table string, item Item) *Command {
	return &Command{
		ID:      CmdItemExists,
		DB:      db,
		StrArgs: []string{table},
		ItemArg: item,
	}
}

// ListItemsCommand returns a pointer to a command structure for mdb.ListItems().
func ListItemsCommand(db CommandDB, table string, limit int64) *Command {
	return &Command{
		ID:      CmdListItems,
		DB:      db,
		StrArgs: []string{table},
		IntArg:  limit,
	}
}

// MustGetFieldTypeCommand returns a pointer to a command structure for mdb.MustGetFieldType().
func MustGetFieldTypeCommand(db CommandDB, table string, field string) *Command {
	return &Command{
		ID:      CmdMustGetFieldType,
		DB:      db,
		StrArgs: []string{table, field},
	}
}

// NewItemCommand returns a pointer to a command structure for mdb.NewItem().
func NewItemCommand(db CommandDB, table string) *Command {
	return &Command{
		ID:      CmdNewItem,
		DB:      db,
		StrArgs: []string{table},
	}
}

// ParseFieldValuesCommand returns a pointer to a command structure for mdb.ParseFieldValues().
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

// SetCommand returns a pointer to a command structure for mdb.Set().
func SetCommand(db CommandDB, table string, item Item, field string, data []Value) *Command {
	return &Command{
		ID:        CmdSet,
		DB:        db,
		StrArgs:   []string{table, field},
		ItemArg:   item,
		ValueArgs: data,
	}
}

// TableExistsCommand returns a pointer to a command structure for mdb.TableExists().
func TableExistsCommand(db CommandDB, table string) *Command {
	return &Command{
		ID:      CmdTableExists,
		DB:      db,
		StrArgs: []string{table},
	}
}

// ToSqlCommand returns a pointer to a command structure for mdb.ToSql().
func ToSqlCommand(db CommandDB, table string, query *Query, limit int64) *Command {
	return &Command{
		ID:       CmdToSQL,
		DB:       db,
		StrArgs:  []string{table},
		QueryArg: *query,
		IntArg:   limit,
	}
}

// GetIntCommand returns a pointer to a command structure for mdb.GetInt().
func GetIntCommand(db CommandDB, key int64) *Command {
	return &Command{
		ID:     CmdGetInt,
		DB:     db,
		IntArg: key,
	}
}

// GetStrCommand returns a pointer to a command structure for mdb.GetStr().
func GetStrCommand(db CommandDB, key int64) *Command {
	return &Command{
		ID:     CmdGetStr,
		DB:     db,
		IntArg: key,
	}
}

// GetBlobCommand returns a pointer to a command structure for mdb.GetBlob().
func GetBlobCommand(db CommandDB, key int64) *Command {
	return &Command{
		ID:     CmdGetBlob,
		DB:     db,
		IntArg: key,
	}
}

// GetDateCommand returns a pointer to a command structure for mdb.GetDate().
func GetDateCommand(db CommandDB, key int64) *Command {
	return &Command{
		ID:     CmdGetDate,
		DB:     db,
		IntArg: key,
	}
}

// SetIntCommand returns a pointer to a command structure for mdb.SetInt().
func SetIntCommand(db CommandDB, key int64, value int64) *Command {
	return &Command{
		ID:      CmdSetInt,
		DB:      db,
		IntArg:  key,
		IntArg2: value,
	}
}

// SetStrCommand returns a pointer to a command structure for mdb.SetStr().
func SetStrCommand(db CommandDB, key int64, value string) *Command {
	return &Command{
		ID:      CmdSetStr,
		DB:      db,
		IntArg:  key,
		StrArgs: []string{value},
	}
}

// SetBlobCommand returns a pointer to a command structure for mdb.SetBlob().
func SetBlobCommand(db CommandDB, key int64, value []byte) *Command {
	return &Command{
		ID:      CmdSetBlob,
		DB:      db,
		IntArg:  key,
		StrArgs: []string{string(value)},
	}
}

// SetDateCommand returns a pointer to a command structure for mdb.SetDate().
func SetDateCommand(db CommandDB, key int64, value time.Time) *Command {
	d := NewDate(value)
	return &Command{
		ID:      CmdSetDate,
		DB:      db,
		IntArg:  key,
		StrArgs: []string{d.String()},
	}
}

// SetDateStrCommand returns a pointer to a command structure for mdb.SetDateStr().
func SetDateStrCommand(db CommandDB, key int64, value string) *Command {
	return &Command{
		ID:      CmdSetDateStr,
		DB:      db,
		IntArg:  key,
		StrArgs: []string{value},
	}
}

// HasIntCommand returns a pointer to a command structure for mdb.HasInt().
func HasIntCommand(db CommandDB, key int64) *Command {
	return &Command{
		ID:     CmdHasInt,
		DB:     db,
		IntArg: key,
	}
}

// HasStr returns a pointer to a command structure for mdb.HasStr().
func HasStrCommand(db CommandDB, key int64) *Command {
	return &Command{
		ID:     CmdHasStr,
		DB:     db,
		IntArg: key,
	}
}

// HasBlobCommand returns a pointer to a command structure for mdb.HasBlob().
func HasBlobCommand(db CommandDB, key int64) *Command {
	return &Command{
		ID:     CmdHasBlob,
		DB:     db,
		IntArg: key,
	}
}

// HasDateCommand returns a pointer to a command structure for mdb.HasDate().
func HasDateCommand(db CommandDB, key int64) *Command {
	return &Command{
		ID:     CmdHasDate,
		DB:     db,
		IntArg: key,
	}
}

// DeleteIntCommand returns a pointer to a command structure for mdb.DeleteInt().
func DeleteIntCommand(db CommandDB, key int64) *Command {
	return &Command{
		ID:     CmdDeleteInt,
		DB:     db,
		IntArg: key,
	}
}

// DeleteStrCommand returns a pointer to a command structure for mdb.DeleteStr().
func DeleteStrCommand(db CommandDB, key int64) *Command {
	return &Command{
		ID:     CmdDeleteStr,
		DB:     db,
		IntArg: key,
	}
}

// DeleteBlobCommand returns a pointer to a command structure for mdb.DeleteBlob().
func DeleteBlobCommand(db CommandDB, key int64) *Command {
	return &Command{
		ID:     CmdDeleteBlob,
		DB:     db,
		IntArg: key,
	}
}

// DeleteDateCommand returns a pointer to a command structure for mdb.DeleteDate().
func DeleteDateCommand(db CommandDB, key int64) *Command {
	return &Command{
		ID:     CmdDeleteDate,
		DB:     db,
		IntArg: key,
	}
}

// ListIntCommand returns a pointer to a command structure for mdb.ListInt().
func ListIntCommand(db CommandDB) *Command {
	return &Command{
		ID: CmdListInt,
		DB: db,
	}
}

// ListStrCommand returns a pointer to a command structure for mdb.ListStr().
func ListStrCommand(db CommandDB) *Command {
	return &Command{
		ID: CmdListStr,
		DB: db,
	}
}

// ListBlobCommand returns a pointer to a command structure for mdb.ListBlob().
func ListBlobCommand(db CommandDB) *Command {
	return &Command{
		ID: CmdListBlob,
		DB: db,
	}
}

// ListDateCommand returns a pointer to a command structure for mdb.ListDate().
func ListDateCommand(db CommandDB) *Command {
	return &Command{
		ID: CmdListDate,
		DB: db,
	}
}

// BackupCommand returns a pointer to a command structure for mdb.Backup().
func BackupCommand(db CommandDB, destination string) *Command {
	return &Command{
		ID:      CmdBackup,
		DB:      db,
		StrArgs: []string{destination},
	}
}

// RemoveItemCommand returns a pointer to a command structure for mdb.RemoveItem().
func RemoveItemCommand(db CommandDB, table string, item Item) *Command {
	return &Command{
		ID:      CmdRemoveItem,
		DB:      db,
		StrArgs: []string{table},
		ItemArg: item,
	}
}
