package minidb

import (
	"sync"
	"time"
)

// CommandID represents the type of a command in the Exec() function.
type CommandID int

// The actual Exec() CommandID values. Names mirror the respective functions.
const (
	// CmdOpen is the type of an Open command struct.
	CmdOpen CommandID = iota + 1
	// CmdBegin opens a transaction
	CmdBegin
	// CmdRollback rolls back a transaction
	CmdRollback
	// CmdCommit commits a transaction
	CmdCommit
	// CmdAddTable is the type of an AddTable command struct.
	CmdAddTable
	// CmdClose is the type of a Close command struct.
	CmdClose
	// CmdCount is the type of a Count command struct.
	CmdCount
	// CmdFind is the type of a Find command struct.
	CmdFind
	// CmdGet is the type of a Get command struct.
	CmdGet
	// CmdGetTables is the type of a GetTables command struct.
	CmdGetTables
	// CmdIsListField is the type of an IsListField command struct.
	CmdIsListField
	// CmdItemExists is the type of an ItemExists command struct.
	CmdItemExists
	// CmdListItems is the type of a ListItems command struct.
	CmdListItems
	// CmdNewItem is the type of a NewItem command struct.
	CmdNewItem
	// CmdParseFieldValues is the type of a ParseFieldValues command struct.
	CmdParseFieldValues
	// CmdSet is the type of a Set command struct.
	CmdSet
	// CmdTableExists is the type of a TableExists command struct.
	CmdTableExists
	// CmdToSQL is the type of a ToSQL command struct.
	CmdToSQL
	// CmdFieldIsNull is the type of a FieldIsNull command struct.
	CmdFieldIsNull
	// CmdFieldExists is the type of a FieldExists command struct.
	CmdFieldExists
	// CmdGetFields is the type of a GetFields command struct.
	CmdGetFields
	// CmdIsEmptyListField is the type of an IsEmptyListField command struct.
	CmdIsEmptyListField
	// CmdMustGetFieldType is the type of a MustGetFieldType command struct.
	CmdMustGetFieldType
	// CmdGetInt is the type of a GetInt command struct.
	CmdGetInt
	// CmdGetStr is the type of a GetStr command struct.
	CmdGetStr
	// CmdGetBlob is the type of a GetBlob command struct.
	CmdGetBlob
	// CmdGetDate is the type of a GetDate command struct.
	CmdGetDate
	// CmdSetInt is the type of a SetInt command struct.
	CmdSetInt
	// CmdSetStr is the type of a SetStr commmand struct.
	CmdSetStr
	// CmdSetBlob is the type of a SetBlob commmand struct.
	CmdSetBlob
	// CmdSetDate is the type of a SetDate command struct.
	CmdSetDate
	// CmdDeleteInt is the type of a DeleteInt command struct.
	CmdDeleteInt
	// CmdDeleteStr is the type of a DeleteStr command struct.
	CmdDeleteStr
	// CmdDeleteBlob is the type of a DeleteBlob command struct.
	CmdDeleteBlob
	// CmdDeleteDate is the type of a DeleteDate command struct.
	CmdDeleteDate
	// CmdHadInt is the type of a HasInt command struct.
	CmdHasInt
	// CmdHasStr is the type of a HasStr command struct.
	CmdHasStr
	// CmdHasBlob is the type of a HasBlob command struct.
	CmdHasBlob
	// CmdHasDate is the type of a HasDate command struct.
	CmdHasDate
	// CmdListInt is the type of a ListInt commmand struct.
	CmdListInt
	// CmdListStr is the type of a ListStr command struct.
	CmdListStr
	// CmdListBlob is the type of a ListBlob command struct.
	CmdListBlob
	// CmdListDate is the type of a ListDate command struct.
	CmdListDate
	// CmdSetDateStr is the type of a SetDateStr command struct.
	CmdSetDateStr
	// CmdFieldIsEmpty is the type of a FieldIsEmpty commmand struct.
	CmdFieldIsEmpty
	// CmdBackup is the type of a Backup commmand struct.
	CmdBackup
	// CmdRemoveItem is the type of a RemoveItem commmand struct.
	CmdRemoveItem
	// CmdIndex is the type of an Index command struct.
	CmdIndex
)

// CommandDB is the database that has been opened.
type CommandDB string

// TxID is the ID of a transaction.
type TxID int64

// Command structures contains all information needed to execute an arbitrary command.
// Use Exec() to execute a command and get the result.
// Every function <Name> in minidb has a corresponding function <Name>Command that
// returns the corresponding command. Consult the API for <Name> for help, as the input parameters
// are exactly the same, except that the first two arguments are a database path of type CommandDB and
// often also a transaction id of type TxID.
// You should never use Command structures directly, but use the provided wrapper functions for strong typing.
// Result structures have the HasError field set to true if an error has occurred.
// Commands and results can be serialized to json.
type Command struct {
	ID        CommandID `json:"id"`
	DB        CommandDB `json:"dbid"`
	Tx        TxID      `json:"txid"`
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
var openTxs map[TxID]*Tx
var connections map[CommandDB]int
var txCounter TxID

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
	ErrIndexFailed
	ErrUnknownTx
	ErrBeginFailed
	ErrCommitFailed
	ErrRollbackFailed
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

func getTx(cmd *Command) (*Tx, *Result) {
	mutex.RLock()
	defer mutex.RUnlock()
	theTx, ok := openTxs[cmd.Tx]
	if ok {
		return theTx, nil
	}
	r := Result{HasError: true, Int: ErrUnknownTx}
	r.Str = Fail("exec failed: transaction '%d' unknown", int64(cmd.Tx)).Error()
	return nil, &r
}

// CloseAllDBs closes all open DB connections and cleans up resources.
func CloseAllDBs() {
	mutex.Lock()
	defer mutex.Unlock()
	for tx := range openTxs {
		openTxs[tx].Commit()
		openTxs[tx] = nil
	}
	for db := range openDBs {
		openDBs[db].Close()
		connections[db] = 0
		openDBs[db] = nil
	}
}

func init() {
	openDBs = make(map[CommandDB]*MDB)
	connections = make(map[CommandDB]int)
	openTxs = make(map[TxID]*Tx)
	txCounter++
}

// Exec takes a Command structure and executes it, returning a Result or an error.
// This function is a large switch, as a wrapper around the more specific API functions.
// It incurs a runtime penalty and should only used when needed (e.g. when commands
// have to be marshalled and unmarshalled).
func Exec(cmd *Command) *Result {
	var r Result
	var theDB *MDB
	var theTx *Tx
	var err error
	var errResult *Result

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
	theTx, errResult = getTx(cmd)

	switch cmd.ID {
	case CmdBegin:
		mutex.Lock()
		defer mutex.Unlock()
		theTx, err = theDB.Begin()
		if err != nil {
			r.HasError = true
			r.Int = ErrBeginFailed
			r.Str = err.Error()
			return &r
		}
		txCounter++
		openTxs[txCounter] = theTx
		r.Int = int64(txCounter)

	case CmdCommit:
		if theTx == nil {
			return errResult
		}
		err = theTx.Commit()
		if err != nil {
			r.HasError = true
			r.Int = ErrCommitFailed
			r.Str = err.Error()
		}

	case CmdRollback:
		if theTx == nil {
			return errResult
		}
		err = theTx.Rollback()
		if err != nil {
			r.HasError = true
			r.Int = ErrRollbackFailed
			r.Str = err.Error()
		}

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
		defer mutex.Unlock()
		if connections[cmd.DB] == 1 {
			err = theDB.Close()
			delete(openDBs, cmd.DB)
			delete(connections, cmd.DB)
		} else {
			connections[cmd.DB] -= 1
		}
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
		if theTx == nil {
			return errResult
		}
		err = theTx.Set(cmd.StrArgs[0], cmd.ItemArg, cmd.StrArgs[1], cmd.ValueArgs)
		if err != nil {
			r.HasError = true
			r.Int = ErrSetFailed
			r.Str = err.Error()
		}

	case CmdRemoveItem:
		if theTx == nil {
			return errResult
		}
		err = theTx.RemoveItem(cmd.StrArgs[0], cmd.ItemArg)
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
		if theTx == nil {
			return errResult
		}
		theTx.SetInt(cmd.IntArg, cmd.IntArg2)

	case CmdSetStr:
		if theTx == nil {
			return errResult
		}
		theTx.SetStr(cmd.IntArg, cmd.StrArgs[0])

	case CmdSetBlob:
		if theTx == nil {
			return errResult
		}
		theTx.SetBlob(cmd.IntArg, []byte(cmd.StrArgs[0]))

	case CmdSetDate:
		if theTx == nil {
			return errResult
		}
		t, err := ParseTime(cmd.StrArgs[0])
		if err != nil {
			r.HasError = true
			r.Int = ErrInvalidDate
			r.Str = err.Error()
		} else {
			theTx.SetDate(cmd.IntArg, t)
		}

	case CmdSetDateStr:
		if theTx == nil {
			return errResult
		}
		theTx.SetDateStr(cmd.IntArg, cmd.StrArgs[0])

	case CmdHasInt:
		r.Bool = theDB.HasInt(cmd.IntArg)

	case CmdHasStr:
		r.Bool = theDB.HasStr(cmd.IntArg)

	case CmdHasBlob:
		r.Bool = theDB.HasBlob(cmd.IntArg)

	case CmdHasDate:
		r.Bool = theDB.HasDate(cmd.IntArg)

	case CmdDeleteInt:
		if theTx == nil {
			return errResult
		}
		theTx.DeleteInt(cmd.IntArg)

	case CmdDeleteStr:
		if theTx == nil {
			return errResult
		}
		theTx.DeleteStr(cmd.IntArg)

	case CmdDeleteBlob:
		if theTx == nil {
			return errResult
		}
		theTx.DeleteBlob(cmd.IntArg)

	case CmdDeleteDate:
		if theTx == nil {
			return errResult
		}
		theTx.DeleteDate(cmd.IntArg)

	case CmdListInt:
		r.Ints = theDB.ListInt()

	case CmdListStr:
		r.Ints = theDB.ListStr()

	case CmdListBlob:
		r.Ints = theDB.ListBlob()

	case CmdListDate:
		r.Ints = theDB.ListDate()

	case CmdIndex:
		if theTx == nil {
			return errResult
		}
		err := theTx.Index(cmd.StrArgs[0], cmd.StrArgs[1])
		if err != nil {
			r.HasError = true
			r.Int = ErrIndexFailed
			r.Str = err.Error()
		}

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

// BeginCommand returns a pointer to a command structure for mdb.Begin().
func BeginCommand(db CommandDB) *Command {
	return &Command{
		ID: CmdBegin,
		DB: db,
	}
}

// CommitCommand returns a pointer to a command structure for tx.Commit().
func CommitCommand(db CommandDB, tx TxID) *Command {
	return &Command{
		ID: CmdCommit,
		DB: db,
		Tx: tx,
	}
}

// RollbackCommand returns a pointer to a command structure for tx.Rollback().
func RollbackCommand(db CommandDB, tx TxID) *Command {
	return &Command{
		ID: CmdRollback,
		DB: db,
		Tx: tx,
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

// CountCommand returns a pointer to a command structure for tx.Count()
func CountCommand(db CommandDB, table string) *Command {
	return &Command{
		ID:      CmdCount,
		DB:      db,
		StrArgs: []string{table},
	}
}

// FieldExistsCommand returns a pointer to a command structure for tx.FieldExists().
func FieldExistsCommand(db CommandDB, table string, field string) *Command {
	return &Command{
		ID:      CmdFieldExists,
		DB:      db,
		StrArgs: []string{table, field},
	}
}

// FieldIsNullCommand returns a pointer to a command structure for tx.FieldIsNull().
func FieldIsNullCommand(db CommandDB, item Item, field string) *Command {
	return &Command{
		ID:      CmdFieldIsNull,
		DB:      db,
		StrArgs: []string{field},
		ItemArg: item,
	}
}

// FieldIsEmptyCommand returns a pointer to a command structure for tx.FieldIsEmpty().
func FieldIsEmptyCommand(db CommandDB, item Item, field string) *Command {
	return &Command{
		ID:      CmdFieldIsEmpty,
		DB:      db,
		StrArgs: []string{field},
		ItemArg: item,
	}
}

// FindCommand returns a pointer to a command structure for tx.Find().
func FindCommand(db CommandDB, query *Query, limit int64) *Command {
	return &Command{
		ID:       CmdFind,
		DB:       db,
		QueryArg: *query,
		IntArg:   limit,
	}
}

// GetCommand returns a pointer to a command structure for tx.Get().
func GetCommand(db CommandDB, table string, item Item, field string) *Command {
	return &Command{
		ID:      CmdGet,
		DB:      db,
		StrArgs: []string{table, field},
		ItemArg: item,
	}
}

// GetFieldsCommand returns a pointer to a command structure for tx.GetFields().
func GetFieldsCommand(db CommandDB, table string) *Command {
	return &Command{
		ID:      CmdGetFields,
		DB:      db,
		StrArgs: []string{table},
	}
}

// GetTablesCommand returns a pointer to a command structure for tx.GetTables().
func GetTablesCommand(db CommandDB) *Command {
	return &Command{
		ID: CmdGetTables,
		DB: db,
	}
}

// IsEmptyListFieldCommand returns a pointer to a command structure for tx.IsEmptyListField().
func IsEmptyListFieldCommand(db CommandDB, table string, item Item, field string) *Command {
	return &Command{
		ID:      CmdIsEmptyListField,
		DB:      db,
		StrArgs: []string{table, field},
		ItemArg: item,
	}
}

// IsListFieldCommand returns a pointer to a command structure for tx.IsListField().
func IsListFieldCommand(db CommandDB, table string, field string) *Command {
	return &Command{
		ID:      CmdIsListField,
		DB:      db,
		StrArgs: []string{table, field},
	}
}

// ItemExistsCommand returns a pointer to a command structure for tx.ItemExists().
func ItemExistsCommand(db CommandDB, table string, item Item) *Command {
	return &Command{
		ID:      CmdItemExists,
		DB:      db,
		StrArgs: []string{table},
		ItemArg: item,
	}
}

// ListItemsCommand returns a pointer to a command structure for tx.ListItems().
func ListItemsCommand(db CommandDB, table string, limit int64) *Command {
	return &Command{
		ID:      CmdListItems,
		DB:      db,
		StrArgs: []string{table},
		IntArg:  limit,
	}
}

// MustGetFieldTypeCommand returns a pointer to a command structure for tx.MustGetFieldType().
func MustGetFieldTypeCommand(db CommandDB, table string, field string) *Command {
	return &Command{
		ID:      CmdMustGetFieldType,
		DB:      db,
		StrArgs: []string{table, field},
	}
}

// NewItemCommand returns a pointer to a command structure for tx.NewItem().
func NewItemCommand(db CommandDB, tx TxID, table string) *Command {
	return &Command{
		ID:      CmdNewItem,
		DB:      db,
		Tx:      tx,
		StrArgs: []string{table},
	}
}

// ParseFieldValuesCommand returns a pointer to a command structure for tx.ParseFieldValues().
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

// SetCommand returns a pointer to a command structure for tx.Set().
func SetCommand(db CommandDB, tx TxID, table string, item Item, field string, data []Value) *Command {
	return &Command{
		ID:        CmdSet,
		DB:        db,
		Tx:        tx,
		StrArgs:   []string{table, field},
		ItemArg:   item,
		ValueArgs: data,
	}
}

// TableExistsCommand returns a pointer to a command structure for tx.TableExists().
func TableExistsCommand(db CommandDB, table string) *Command {
	return &Command{
		ID:      CmdTableExists,
		DB:      db,
		StrArgs: []string{table},
	}
}

// ToSqlCommand returns a pointer to a command structure for tx.ToSql().
func ToSqlCommand(db CommandDB, table string, query *Query, limit int64) *Command {
	return &Command{
		ID:       CmdToSQL,
		DB:       db,
		StrArgs:  []string{table},
		QueryArg: *query,
		IntArg:   limit,
	}
}

// GetIntCommand returns a pointer to a command structure for tx.GetInt().
func GetIntCommand(db CommandDB, key int64) *Command {
	return &Command{
		ID:     CmdGetInt,
		DB:     db,
		IntArg: key,
	}
}

// GetStrCommand returns a pointer to a command structure for tx.GetStr().
func GetStrCommand(db CommandDB, key int64) *Command {
	return &Command{
		ID:     CmdGetStr,
		DB:     db,
		IntArg: key,
	}
}

// GetBlobCommand returns a pointer to a command structure for tx.GetBlob().
func GetBlobCommand(db CommandDB, key int64) *Command {
	return &Command{
		ID:     CmdGetBlob,
		DB:     db,
		IntArg: key,
	}
}

// GetDateCommand returns a pointer to a command structure for tx.GetDate().
func GetDateCommand(db CommandDB, key int64) *Command {
	return &Command{
		ID:     CmdGetDate,
		DB:     db,
		IntArg: key,
	}
}

// SetIntCommand returns a pointer to a command structure for tx.SetInt().
func SetIntCommand(db CommandDB, tx TxID, key int64, value int64) *Command {
	return &Command{
		ID:      CmdSetInt,
		DB:      db,
		Tx:      tx,
		IntArg:  key,
		IntArg2: value,
	}
}

// SetStrCommand returns a pointer to a command structure for tx.SetStr().
func SetStrCommand(db CommandDB, tx TxID, key int64, value string) *Command {
	return &Command{
		ID:      CmdSetStr,
		DB:      db,
		Tx:      tx,
		IntArg:  key,
		StrArgs: []string{value},
	}
}

// SetBlobCommand returns a pointer to a command structure for tx.SetBlob().
func SetBlobCommand(db CommandDB, tx TxID, key int64, value []byte) *Command {
	return &Command{
		ID:      CmdSetBlob,
		DB:      db,
		Tx:      tx,
		IntArg:  key,
		StrArgs: []string{string(value)},
	}
}

// SetDateCommand returns a pointer to a command structure for tx.SetDate().
func SetDateCommand(db CommandDB, tx TxID, key int64, value time.Time) *Command {
	d := NewDate(value)
	return &Command{
		ID:      CmdSetDate,
		DB:      db,
		Tx:      tx,
		IntArg:  key,
		StrArgs: []string{d.String()},
	}
}

// SetDateStrCommand returns a pointer to a command structure for tx.SetDateStr().
func SetDateStrCommand(db CommandDB, tx TxID, key int64, value string) *Command {
	return &Command{
		ID:      CmdSetDateStr,
		DB:      db,
		Tx:      tx,
		IntArg:  key,
		StrArgs: []string{value},
	}
}

// HasIntCommand returns a pointer to a command structure for tx.HasInt().
func HasIntCommand(db CommandDB, key int64) *Command {
	return &Command{
		ID:     CmdHasInt,
		DB:     db,
		IntArg: key,
	}
}

// HasStrCommand returns a pointer to a command structure for tx.HasStr().
func HasStrCommand(db CommandDB, key int64) *Command {
	return &Command{
		ID:     CmdHasStr,
		DB:     db,
		IntArg: key,
	}
}

// HasBlobCommand returns a pointer to a command structure for tx.HasBlob().
func HasBlobCommand(db CommandDB, key int64) *Command {
	return &Command{
		ID:     CmdHasBlob,
		DB:     db,
		IntArg: key,
	}
}

// HasDateCommand returns a pointer to a command structure for tx.HasDate().
func HasDateCommand(db CommandDB, key int64) *Command {
	return &Command{
		ID:     CmdHasDate,
		DB:     db,
		IntArg: key,
	}
}

// DeleteIntCommand returns a pointer to a command structure for tx.DeleteInt().
func DeleteIntCommand(db CommandDB, tx TxID, key int64) *Command {
	return &Command{
		ID:     CmdDeleteInt,
		DB:     db,
		Tx:     tx,
		IntArg: key,
	}
}

// DeleteStrCommand returns a pointer to a command structure for tx.DeleteStr().
func DeleteStrCommand(db CommandDB, tx TxID, key int64) *Command {
	return &Command{
		ID:     CmdDeleteStr,
		DB:     db,
		Tx:     tx,
		IntArg: key,
	}
}

// DeleteBlobCommand returns a pointer to a command structure for tx.DeleteBlob().
func DeleteBlobCommand(db CommandDB, tx TxID, key int64) *Command {
	return &Command{
		ID:     CmdDeleteBlob,
		DB:     db,
		Tx:     tx,
		IntArg: key,
	}
}

// DeleteDateCommand returns a pointer to a command structure for tx.DeleteDate().
func DeleteDateCommand(db CommandDB, tx TxID, key int64) *Command {
	return &Command{
		ID:     CmdDeleteDate,
		DB:     db,
		Tx:     tx,
		IntArg: key,
	}
}

// ListIntCommand returns a pointer to a command structure for tx.ListInt().
func ListIntCommand(db CommandDB, tx TxID) *Command {
	return &Command{
		ID: CmdListInt,
		DB: db,
		Tx: tx,
	}
}

// ListStrCommand returns a pointer to a command structure for tx.ListStr().
func ListStrCommand(db CommandDB) *Command {
	return &Command{
		ID: CmdListStr,
		DB: db,
	}
}

// ListBlobCommand returns a pointer to a command structure for tx.ListBlob().
func ListBlobCommand(db CommandDB) *Command {
	return &Command{
		ID: CmdListBlob,
		DB: db,
	}
}

// ListDateCommand returns a pointer to a command structure for tx.ListDate().
func ListDateCommand(db CommandDB) *Command {
	return &Command{
		ID: CmdListDate,
		DB: db,
	}
}

// BackupCommand returns a pointer to a command structure for tx.Backup().
func BackupCommand(db CommandDB, destination string) *Command {
	return &Command{
		ID:      CmdBackup,
		DB:      db,
		StrArgs: []string{destination},
	}
}

// RemoveItemCommand returns a pointer to a command structure for tx.RemoveItem().
func RemoveItemCommand(db CommandDB, tx TxID, table string, item Item) *Command {
	return &Command{
		ID:      CmdRemoveItem,
		DB:      db,
		Tx:      tx,
		StrArgs: []string{table},
		ItemArg: item,
	}
}

// IndexCommand returns a pointer to a command structure for tx.Index().
func IndexCommand(db CommandDB, tx TxID, table string, field string) *Command {
	return &Command{
		ID:      CmdIndex,
		DB:      db,
		Tx:      tx,
		StrArgs: []string{table, field},
	}
}
