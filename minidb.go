// Package minidb is a minimalist database. It stores items in tables, where each item has a fixed number of fields.
// The package has two APIs. The direct API is centered around MDB structures that represent database connections.
// Functions of MDB call directly the underlying database layer. The second API is slower and may be used
// for cases when commands and results have to be serialized. It uses Command structures that are created
// by functions like OpenCommand, AddTableCommand, etc. These are passed to Exec() which returns a Result
// structure that is populated with result values.
package minidb

import (
	"database/sql"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3" // The driver for sqlite3 is pulled in.
)

// MDB is the main database object.
type MDB struct {
	base       *sql.DB
	tx         *Tx
	driver     string
	location   string
	globalLock *sync.Mutex
}

// Tx represents a transaction similar to sql.Tx.
type Tx struct {
	tx        *sql.Tx
	prev      *Tx
	mdb       *MDB
	savePoint uint
	released  bool
}

var savePointCounter uint

// Item is a database item. Fields and tables are identified by strings.
type Item int64

// FieldType is a field in the database, which might be a list type or a base type.
type FieldType int

const (
	// DBError represents an error in a field definition.
	DBError FieldType = iota + 1
	// DBInt is the type of an int64 field.
	DBInt
	// DBString is the type of a string field.
	DBString
	// DBBlob is the type of a []byte field.
	DBBlob
	// DBIntList is the type of a list of int64 field.
	DBIntList
	// DBStringList is the type of a list of strings field.
	DBStringList
	// DBBlobList is the type of a list of []byte field, i.e., corresponding to [][]byte.
	DBBlobList
	// DBDate is the type of an RFC 3339 date field.
	DBDate
	// DBDateList is the type of a list of RFC 3339 dates field.
	DBDateList
)

// ToBaseType converts a list type into the list's base type. A non-list type remains unchanged.
func ToBaseType(t FieldType) FieldType {
	switch t {
	case DBIntList:
		return DBInt
	case DBStringList:
		return DBString
	case DBBlobList:
		return DBBlob
	case DBDateList:
		return DBDate
	default:
		return t
	}
}

// Field represents a database field.
type Field struct {
	Name string    `json:"name"`
	Sort FieldType `json:"sort"`
}

// Fail returns a new error message formatted with fmt.Sprintf.
func Fail(msg string, args ...interface{}) error {
	return fmt.Errorf(msg, args...)
}

// Value holds the values that can be put into the database or retrieved from it.
type Value struct {
	Str  string    `json:"str"`
	Num  int64     `json:"num"`
	Sort FieldType `json:"sort"`
}

// Int returns the value as an int64 and panics if conversion is not possible
func (v *Value) Int() int64 {
	switch v.Sort {
	case DBInt:
		return v.Num
	default:
		panic(fmt.Sprintf("cannot convert %s value to integer",
			GetUserTypeString(v.Sort)))
	}
}

// String returns the string value. It automatically converts int and blob,
// where binary Blob data is Base64 encoded and the int is converted
// to decimal format.
func (v *Value) String() string {
	switch v.Sort {
	case DBInt:
		return fmt.Sprintf("%d", v.Num)
	case DBString, DBDate:
		return v.Str
	case DBBlob:
		return base64.StdEncoding.EncodeToString([]byte(v.Str))
	default:
		panic(fmt.Sprintf("cannot convert %s value to string",
			GetUserTypeString(v.Sort)))
	}
}

// Bytes returns the value as a bytes slice. It automatically converts int64 and string.
// An int64 is written in Little Endian format.
func (v *Value) Bytes() []byte {
	switch v.Sort {
	case DBInt:
		bs := make([]byte, 8)
		binary.LittleEndian.PutUint64(bs, uint64(v.Num))
		return bs
	case DBString, DBDate:
		bs := []byte(v.Str)
		return bs
	case DBBlob:
		bs := []byte(v.Str)
		return bs
	default:
		panic(fmt.Sprintf("cannot convert %s value to bytes",
			GetUserTypeString(v.Sort)))
	}
}

// Datetime returns the time value and panics of no valid date is stored.
func (v *Value) Datetime() time.Time {
	switch v.Sort {
	case DBString, DBBlob, DBDate:
		t, err := ParseTime(v.Str)
		if err != nil {
			panic(fmt.Sprintf("invalid datetime representation '%s'", v.Str))
		}
		return t
	default:
		panic(fmt.Sprintf("cannot convert %s value to date", GetUserTypeString(v.Sort)))
	}
}

// NewInt creates a value that stores an int64.
func NewInt(n int64) Value {
	return Value{Num: n, Sort: DBInt}
}

// NewString creates a value that stores a string.
func NewString(s string) Value {
	return Value{Str: s, Sort: DBString}
}

// NewBytes creates a value that holds a []byte slice. This is similar to String() but notice
// that strings and byte slices are handled differently in the database. For example,
// byte slices may contain NULL characters and may be converted to and from Base64.
func NewBytes(b []byte) Value {
	return Value{Str: string(b), Sort: DBBlob}
}

// NewDate create a value that holds a datetime.
func NewDate(t time.Time) Value {
	return Value{Str: t.UTC().Format(time.RFC3339), Sort: DBDate}
}

// NewDateStr creates a value that holds a datetime given by a RFC3339 representation.
// The correctness of the date string is not validated, so use this function with care.
func NewDateStr(d string) Value {
	return Value{Str: d, Sort: DBDate}
}

var validTable *regexp.Regexp
var validFieldName *regexp.Regexp
var validItemName *regexp.Regexp

func init() {
	validTable = regexp.MustCompile(`^\p{L}+[_0-9\p{L}]*$`)
	validFieldName = regexp.MustCompile(`^\p{L}+[_0-9\p{L}]*$`)
	validItemName = regexp.MustCompile(`^\d+$`)
}

func isListFieldType(field FieldType) bool {
	switch field {
	case DBStringList, DBIntList, DBBlobList, DBDateList:
		return true
	default:
		return false
	}
}

func listFieldToTableName(ownerTable string, field string) string {
	return "_" + ownerTable + "_" + field
}

func getTypeString(field FieldType) string {
	switch field {
	case DBString, DBStringList:
		return "TEXT"
	case DBBlob, DBBlobList:
		return "BLOB"
	case DBDate, DBDateList:
		return "DATE"
	default:
		return "INTEGER"
	}
}

// GetUserTypeString returns a user-readable string for the type of a field.
func GetUserTypeString(field FieldType) string {
	switch field {
	case DBString:
		return "string"
	case DBStringList:
		return "string-list"
	case DBInt:
		return "int"
	case DBIntList:
		return "int-list"
	case DBBlob:
		return "blob"
	case DBBlobList:
		return "blob-list"
	case DBDate:
		return "date"
	case DBDateList:
		return "date-list"
	default:
		return "unknown"
	}
}

func parseFieldType(ident string) (FieldType, error) {
	s := strings.ToLower(ident)
	switch s {
	case "int", "integer":
		return DBInt, nil
	case "str", "string", "text", "txt":
		return DBString, nil
	case "blob":
		return DBBlob, nil
	case "date":
		return DBDate, nil
	case "string-list", "str-list", "text-list", "txt-list":
		return DBStringList, nil
	case "blob-list":
		return DBBlobList, nil
	case "int-list", "integer-list":
		return DBIntList, nil
	case "date-list":
		return DBDateList, nil
	}
	return DBError,
		Fail("Invalid field type '%s', should be one of int,string,blob,int-list,string-list,blob-list", ident)
}

// ParseFieldDesc parses the given string slice into a []Field slice based on
// the format "type name", or returns an error. This can be used for command line parsing.
func ParseFieldDesc(desc []string) ([]Field, error) {
	result := make([]Field, 0)
	if len(desc)%2 != 0 {
		return nil, Fail("invalid field descriptions, they must be of the form <type> <fieldname>!")
	}
	if len(desc) == 0 {
		return nil, Fail("no fields specified!")
	}
	for i := 0; i < len(desc)-1; i += 2 {
		ftype, err := parseFieldType(desc[i])
		if err != nil {
			return nil, err
		}
		if !validFieldName.MatchString(desc[i+1]) {
			return nil, Fail("invalid field name '%s'", desc[i+1])
		}
		if strings.ToLower(desc[i+1]) == "id" {
			return nil, Fail("fields may not be called 'id'!")
		}
		result = append(result, Field{desc[i+1], ftype})
	}
	return result, nil
}

var errNilDB = Fail("db object is nil")

func (db *MDB) init() error {
	if db.base == nil {
		return errNilDB
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.tx.Exec(`CREATE TABLE IF NOT EXISTS _TABLES (Id INTEGER PRIMARY KEY,
Name TEXT NOT NULL)`)
	if err != nil {
		return err
	}
	_, err = tx.tx.Exec(`CREATE INDEX IF NOT EXISTS _TABIDX ON _TABLES (Name)`)
	if err != nil {
		return err
	}
	_, err = tx.tx.Exec(`CREATE TABLE IF NOT EXISTS _COLS (Id INTEGER PRIMARY KEY,
	 Name STRING NOT NULL,
	 FieldType INTEGER NOT NULL,
	Owner INTEGER NOT NULL,
	FOREIGN KEY(Owner) REFERENCES _TABLES(Id))`)
	if err != nil {
		return err
	}
	_, err = tx.tx.Exec(`CREATE TABLE IF NOT EXISTS _KVINT (Id INTEGER PRIMARY KEY NOT NULL, Value INTEGER NOT NULL)`)
	if err != nil {
		return err
	}
	_, err = tx.tx.Exec(`CREATE TABLE IF NOT EXISTS _KVSTR (Id INTEGER PRIMARY KEY NOT NULL, Value TEXT NOT NULL)`)
	if err != nil {
		return err
	}
	_, err = tx.tx.Exec(`CREATE TABLE IF NOT EXISTS _KVBLOB (Id INTEGER PRIMARY KEY NOT NULL, Value BLOB NOT NULL)`)
	if err != nil {
		return err
	}
	_, err = tx.tx.Exec(`CREATE TABLE IF NOT EXISTS _KVDATE (Id INTEGER PRIMARY KEY NOT NULL, Value TEXT NOT NULL)`)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// Open creates or opens a minidb.
func Open(driver string, file string) (*MDB, error) {
	db := new(MDB)
	base, err := sql.Open(driver, file)
	if err != nil {
		return nil, err
	}
	if base == nil {
		return nil, errNilDB
	}
	db.globalLock = &sync.Mutex{}
	db.base = base
	db.driver = driver
	db.location = file
	if err := db.init(); err != nil {
		return nil, Fail("cannot initialize database: %s", err)
	}
	return db, nil
}

// Backup a database and return the original opened, not the backed up database.
// The new MDB pointer needs to be used after this method unless an error
// has occurred since the old one might have become invalid.
// This function may result in a corrupt copy if the database is open by another
// process, so you need to make sure that it isn't.
func (db *MDB) Backup(destination string) error {
	if db.base == nil {
		return Fail("the database must be open to back it up, this one is closed")
	}
	// use manual copy for now (should use sqlite3 backup API for sqlite3)
	src := db.location
	driver := db.driver
	db.Close()
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("backup: %s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	dest, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer dest.Close()
	_, err = io.Copy(dest, source)
	if err != nil {
		return err
	}
	newdb, err := Open(driver, src)
	if err != nil {
		return err
	}
	db = newdb
	return err
}

// Close closes the database, making sure that all remaining transactions are finished.
func (db *MDB) Close() error {
	if db.base != nil {
		_, _ = db.base.Exec(`PRAGMA optimize;`)
		err := db.base.Close()
		if err != nil {
			return Fail("ERROR Failed to close database - %s.\n", err)
		}
		db.driver = ""
		db.location = ""
		db.globalLock = nil
	}
	return nil
}

// Base returns the base sqlx.DB that minidb uses for its underlying storage.
func (db *MDB) Base() *sql.DB {
	return db.base
}

// Begin starts a transaction.
func (db *MDB) Begin() (*Tx, error) {
	if db.globalLock == nil {
		return nil, errors.New("attempt to open a transaction on a closed DB")
	}
	db.globalLock.Lock()
	defer db.globalLock.Unlock()
	if db.tx == nil {
		//fmt.Println("*** new real transaction")
		sqltx, err := db.base.Begin()
		if err != nil {
			return nil, err
		}
		tx := &Tx{
			tx:  sqltx,
			mdb: db,
		}
		db.tx = tx
		return tx, nil
	}
	savePointCounter++
	tx := &Tx{
		tx:        db.tx.tx,
		mdb:       db,
		prev:      db.tx,
		savePoint: savePointCounter,
	}
	db.tx = tx
	_, err := db.tx.tx.Exec(fmt.Sprintf("SAVEPOINT SP%d;", tx.savePoint))
	//fmt.Printf("*** New savepoint SP%d\n", savePoint)
	if err != nil {
		return nil, fmt.Errorf("minidb begin transaction failed, %s", err)
	}
	return tx, nil
}

// Commit the changes to the database.
func (tx *Tx) Commit() error {
	if tx.mdb.globalLock == nil {
		return errors.New("attempt to commit a transaction of a closed DB")
	}
	tx.mdb.globalLock.Lock()
	defer tx.mdb.globalLock.Unlock()
	tx.mdb.tx = tx.prev
	if tx.prev == nil {
		//fmt.Println("*** real commit")
		return tx.tx.Commit()
	}
	if tx.released {
		return errors.New("nested transaction has already been rolled back or commmitted")
	}
	_, err := tx.tx.Exec(fmt.Sprintf("RELEASE SP%d;", tx.savePoint))
	if err != nil {
		return fmt.Errorf("minidb commit transaction failed, %s", err)
	}
	//fmt.Printf("*** release savepoint SP%d\n", savePoint)
	tx.released = true
	return nil
}

// Rollback the changes in the database.
func (tx *Tx) Rollback() error {
	if tx.mdb.globalLock == nil {
		return errors.New("attempt to rollback a transaction of a closed DB")
	}
	tx.mdb.globalLock.Lock()
	defer tx.mdb.globalLock.Unlock()
	if tx.released {
		//fmt.Println("*** rollback after savepoint release (do nothing)")
		return nil
	}
	tx.mdb.tx = tx.prev
	tx.released = true
	if tx.prev == nil {
		//fmt.Println("*** real rollback")
		return tx.tx.Rollback()
	}
	//fmt.Printf("*** rollback to savepoint SP%d\n", savePoint)
	_, err := tx.tx.Exec(fmt.Sprintf("ROLLBACK TO SP%d", tx.savePoint))
	if err != nil {
		return fmt.Errorf("minidb rollback transaction failed, %s", err)
	}
	return nil
}

// TableExists returns true if the table exists, false otherwise.
func (db *MDB) TableExists(table string) bool {
	var result int
	err := db.base.QueryRow(`SELECT EXISTS (SELECT 1 FROM _TABLES WHERE Name=? LIMIT 1)`, table).Scan(&result)
	switch {
	case err == sql.ErrNoRows:
		return false
	case err != nil:
		return false
	default:
		return result > 0
	}
}

// FieldIsNull returns true if the field is null for the item in the table, false otherwise.
func (db *MDB) FieldIsNull(table string, item Item, field string) bool {
	var result int
	err := db.base.QueryRow(fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM "%s" WHERE ? IS NULL and Id=?);`, table), field, item).Scan(&result)
	if err != nil {
		return false
	}
	return result == 0
}

// FieldIsEmpty returns true if the field is null or empty, false otherwise.
func (db *MDB) FieldIsEmpty(table string, item Item, field string) bool {
	if db.FieldIsNull(table, item, field) {
		return true
	}
	var result int
	err := db.base.QueryRow(fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM "%s" WHERE ?='' and Id=?)`, table), field, item).Scan(&result)
	if err != nil {
		return false
	}
	return result == 0
}

// ItemExists returns true if the item exists in the table, false otherwise.
func (db *MDB) ItemExists(table string, item Item) bool {
	var result int
	err := db.base.QueryRow(fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM "%s" WHERE Id=? LIMIT 1)`, table), item).Scan(&result)
	if err != nil {
		return false
	}
	return result > 0
}

// IsListField is true if the field in the table is a list, false otherwise.
// List fields are internally stored as special tables.
func (db *MDB) IsListField(table string, field string) bool {
	return db.TableExists(listFieldToTableName(table, field))
}

// IsEmptyListField returns true if the field is a list field and has no element matching item.
// If the item does not exist, the function returns true as well.
func (db *MDB) IsEmptyListField(table string, item Item, field string) bool {
	var result int
	if !db.IsListField(table, field) {
		return false
	}
	err := db.base.QueryRow(fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM "%s" WHERE Owner=? AND %s IS NOT NULL LIMIT 1)`, table, field), item).Scan(&result)
	if err != nil {
		return true
	}
	return result == 0
}

func (db *MDB) getTableId(table string) (int64, error) {
	var result int64
	err := db.base.QueryRow(`SELECT Id FROM _TABLES WHERE Name=?`, table).Scan(&result)
	if err != nil {
		return 0, err
	}
	return result, nil
}

// FieldExists returns true if the table has the field, false otherwise.
func (db *MDB) FieldExists(table string, field string) bool {
	var result int
	id, err := db.getTableId(table)
	if err != nil {
		return false
	}
	err = db.base.QueryRow(`SELECT EXISTS (SELECT 1 FROM _COLS WHERE Owner=? AND Name=? LIMIT 1)`, id, field).Scan(&result)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return result > 0
}

// MustGetFieldType returns the type of the field. This method panics if the table or field don't exist.
func (db *MDB) MustGetFieldType(table string, field string) FieldType {
	id, _ := db.getTableId(table)
	var result int64
	db.base.QueryRow(`SELECT FieldType FROM _COLS WHERE Owner=? AND Name=? LIMIT 1;`, id, field).Scan(&result)
	return FieldType(result)
}

// ParseFieldValues parses potential value(s) for a field from strings, returns an error if their type is
// incompatible with the field type, the table or field don't exist, or if the input is empty.
// Data for a Blob field must be Base64 encoded, data for an Integer field must be a valid
// digit sequence for a 64 bit integer in base 10 format.
func (db *MDB) ParseFieldValues(table string, field string, data []string) ([]Value, error) {
	if !validTable.MatchString(table) {
		return nil, Fail("invalid table name '%s'", table)
	}
	if !db.TableExists(table) {
		return nil, Fail("table '%s' does not exist", table)
	}
	if !db.FieldExists(table, field) {
		return nil, Fail("field '%s' does not exist in table '%s'", field, table)
	}
	if len(data) == 0 {
		return nil, Fail("no input values given")
	}
	if !db.IsListField(table, field) && len(data) > 1 {
		return nil, Fail("too many input values: expected 1, given %d", len(data))
	}
	t := ToBaseType(db.MustGetFieldType(table, field))
	result := make([]Value, 0, len(data))
	for i := range data {
		switch t {
		case DBInt:
			j, err := strconv.ParseInt(data[i], 10, 64)
			if err != nil {
				return nil, Fail("type error: expected int, given '%s'", data[i])
			}
			result = append(result, NewInt(j))
		case DBBlob:
			b, err := base64.StdEncoding.DecodeString(data[i])
			if err != nil {
				return nil, Fail("type error: expected binary data in Base64 format but the given data seems invalid")
			}
			result = append(result, NewBytes(b))
		case DBString:
			result = append(result, NewString(data[i]))
		case DBDate:
			var t time.Time
			var err error
			if t, err = ParseTime(data[i]); err != nil {
				return nil, Fail("type error: expected datetime in RFC3339 format - %s", err)
			}
			result = append(result, NewDate(t))
		default:
			return nil,
				Fail("internal error: %s %s expects type %d, which is unknown to this version of minidb",
					table, field, t)
		}
	}
	return result, nil
}

// ParseTime parses a time string in RFC3339 format and returns the time or an error if
// the format is wrong.
func ParseTime(s string) (time.Time, error) {
	t, err := time.ParseInLocation(time.RFC3339, s, time.Now().Local().Location())
	if err == nil {
		return t, nil
	}
	return t, Fail("invalid date format '%s'", s)
}

// AddTable is used to create a new table. Table and field names are validated. They need to be alphanumeric
// sequences plus underscore "_" as the only allowed special character. None of the names may start with
// an underscore.
func (db *MDB) AddTable(table string, fields []Field) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if !validTable.MatchString(table) {
		return Fail("invalid table name '%s'", table)
	}
	// normal fields are just columns
	toExec := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s" (Id INTEGER PRIMARY KEY`, table)
	for _, field := range fields {
		if !isListFieldType(field.Sort) {
			toExec += fmt.Sprintf(",\n\"%s\" %s", field.Name, getTypeString(field.Sort))
		}
	}
	toExec += ");"
	_, err = tx.tx.Exec(toExec)
	if err != nil {
		return Fail("cannot create maintenance table: %s", err)
	}
	// list fields are composite tables with name _Basetable_Fieldname
	for _, field := range fields {
		if isListFieldType(field.Sort) {
			toExec = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s" (Id INTEGER PRIMARY KEY,
Owner INTEGER NOT NULL,
%s %s, 
FOREIGN KEY(Owner) REFERENCES %s(Id))`, listFieldToTableName(table, field.Name),
				field.Name,
				getTypeString(field.Sort),
				table)
			_, err = tx.tx.Exec(toExec)
			if err != nil {
				return Fail("cannot create list field %s in table %s: %s", field.Name, table, err)
			}
		}
	}
	// update the internal housekeeping tables
	toExec = "INSERT OR IGNORE INTO _TABLES (Name) VALUES (?)"
	result, err := tx.tx.Exec(toExec, table)
	if err != nil {
		return Fail("Failed to update maintenance table: %s", err)
	}
	tableID, err := result.LastInsertId()
	if err != nil {
		return Fail("failed to update maintenance table: %s", err)
	}
	for _, field := range fields {
		_, err := tx.tx.Exec(`INSERT INTO _COLS (Name,FieldType,Owner) VALUES (?,?,?)`, field.Name, field.Sort, tableID)
		if err != nil {
			return Fail("cannot insert maintenance field %s for table %s: %s",
				field.Name, table, err)
		}
		if isListFieldType(field.Sort) {
			_, err = tx.tx.Exec(`INSERT INTO _TABLES (Name) VALUES (?)`, listFieldToTableName(table, field.Name))
		}
		if err != nil {
			return Fail("cannot insert maintenance list table %s for table %s: %s",
				listFieldToTableName(table, field.Name), table, err)
		}
	}
	return tx.Commit()
}

// Index creates an index for field in table unless the index exists already.
// An index increases the search speed of certain string queries on the field, such as "Person name=joh%".
func (tx *Tx) Index(table, field string) error {
	if !validTable.MatchString(table) {
		return Fail("invalid table name '%s'", table)
	}
	if !tx.mdb.FieldExists(table, field) {
		return Fail("field '%s' does not exist in table '%s'", field, table)
	}
	var realtable string
	if tx.mdb.IsListField(table, field) {
		realtable = listFieldToTableName(table, field)
	} else {
		realtable = table
	}
	indexName := field + "_" + realtable + "_IDX"
	_, err := tx.tx.Exec(fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %s ON %s(%s);`, indexName, realtable, field))
	if err != nil {
		return Fail("failed to create index for field '%s' in table '%s': %s", field, table, err)
	}
	return nil
}

// NewItem creates a new item in the table and returns its numerical ID.
func (db *MDB) NewItem(table string) (Item, error) {
	if !validTable.MatchString(table) {
		return 0, Fail("invalid table name '%s'", table)
	}
	if !db.TableExists(table) {
		return 0, Fail("table '%s' does not exist", table)
	}

	toExec := fmt.Sprintf("INSERT INTO %s DEFAULT VALUES;", table)
	result, err := db.base.Exec(toExec)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return Item(id), nil
}

// UseItem creates a new item with the given ID or returns the item with the given ID
// if it already exists. This may be used when fixed IDs are needed, but should be avoided
// when these are not strictly necessary.
func (db *MDB) UseItem(table string, id uint64) (Item, error) {
	if !validTable.MatchString(table) {
		return 0, Fail("invalid table name '%s'", table)
	}
	if !db.TableExists(table) {
		return 0, Fail("table '%s' does not exist", table)
	}
	if db.ItemExists(table, Item(id)) {
		return Item(id), nil
	}
	toExec := fmt.Sprintf("INSERT INTO %s(Id) VALUES (?);", table)
	_, err := db.base.Exec(toExec, id)
	if err != nil {
		return 0, err
	}
	return Item(id), nil
}

// RemoveItem remove an item from the table.
func (tx *Tx) RemoveItem(table string, item Item) error {
	if !validTable.MatchString(table) {
		return Fail(`invalid table name "%s"`, table)
	}
	if !tx.mdb.TableExists(table) {
		return Fail("table '%s' does not exist", table)
	}
	if tx.mdb.ItemExists(table, item) {
		_, err := tx.tx.Exec(fmt.Sprintf(`DELETE FROM %s WHERE Id=?;`, table), item)
		if err != nil {
			return Fail(`error while deleting %s %d`, table, item)
		}
	}
	return nil
}

// Count returns the number of items in the table.
func (db *MDB) Count(table string) (int64, error) {
	if !validTable.MatchString(table) {
		return 0, Fail("invalid table name '%s'", table)
	}
	if !db.TableExists(table) {
		return 0, Fail("table '%s' does not exist", table)
	}
	var result int64
	err := db.base.QueryRow(fmt.Sprintf(`SELECT COUNT(*) FROM %s;`, table)).Scan(&result)
	if err != nil {
		return 0, err
	}
	return result, nil
}

// ListItems returns a list of items in the table.
func (db *MDB) ListItems(table string, limit int64) ([]Item, error) {
	empty := make([]Item, 0)
	if !validTable.MatchString(table) {
		return empty, Fail("invalid table name '%s'", table)
	}
	if !db.TableExists(table) {
		return empty, Fail("table '%s' does not exist", table)
	}
	rows, err := db.base.Query(fmt.Sprintf(`SELECT (Id) FROM %s;`, table))
	if err != nil {
		return empty, err
	}
	defer rows.Close()
	results := make([]Item, 0)
	var c int64
	for rows.Next() {
		var datum sql.NullInt64
		if err := rows.Scan(&datum); err == nil && datum.Valid {
			results = append(results, Item(datum.Int64))
		}
		c++
		if limit > 0 && c >= limit {
			break
		}
	}
	return results, nil
}

// Get returns the value(s) of a field of an item in a table.
func (db *MDB) Get(table string, item Item, field string) ([]Value, error) {
	if !validTable.MatchString(table) {
		return nil, Fail("invalid table name '%s'", table)
	}
	if !db.TableExists(table) {
		return nil, Fail("table '%s' does not exist", table)
	}
	if !db.ItemExists(table, item) {
		return nil, Fail("no %s %d", table, item)
	}
	if db.IsListField(table, field) {
		return db.getListField(table, item, field)
	}
	return db.getSingleField(table, item, field)
}

func (db *MDB) getSingleField(table string, item Item, field string) ([]Value, error) {
	if !db.FieldExists(table, field) {
		return nil,
			Fail(`no field %s in table %s`, field, table)
	}
	t := db.MustGetFieldType(table, field)
	row := db.base.QueryRow(fmt.Sprintf(`SELECT "%s" FROM "%s" WHERE Id=?;`, field, table), item)
	var intResult sql.NullInt64
	var strResult sql.NullString
	var err error
	switch t {
	case DBInt:
		err = row.Scan(&intResult)
	case DBString, DBBlob, DBDate:
		err = row.Scan(&strResult)
	default:
		return nil,
			Fail("unsupported field type for %s %d %s: %d (try a newer version?)", table, item, field, int(t))
	}
	if err == sql.ErrNoRows {
		return nil,
			Fail("no value for %s %d %s", table, item, field)
	}
	if err != nil {
		return nil,
			Fail("cannot find value for %s %d %s: %s", table, item, field, err)
	}
	vslice := make([]Value, 1)
	switch t {
	case DBInt:
		if !intResult.Valid {
			return nil,
				Fail("no int value for %s %d %s", table, item, field)
		}
		vslice[0] = NewInt(intResult.Int64)
	case DBString, DBBlob, DBDate:
		if !strResult.Valid {
			return nil,
				Fail("no string, blob, or date value for %s %d %s", table, item, field)
		}
		vslice[0] = NewString(strResult.String)
		vslice[0].Sort = t
	}
	return vslice, nil
}

func (db *MDB) getListField(table string, item Item, field string) ([]Value, error) {
	tableName := listFieldToTableName(table, field)
	if !db.TableExists(tableName) {
		return nil,
			Fail("list field %s does not exist in table %s", field, table)
	}
	if db.IsEmptyListField(tableName, item, field) {
		return nil,
			Fail("no values for %s %d %s", table, item, field)
	}
	t := db.MustGetFieldType(table, field)
	rows, err := db.base.Query(fmt.Sprintf(`SELECT "%s" FROM "%s" WHERE Owner=?`, field, tableName), item)
	if err != nil {
		return nil,
			Fail("cannot find values for %s %d %s: %s", table, item, field, err)
	}
	defer rows.Close()
	results := make([]Value, 0)
	var intResult sql.NullInt64
	var strResult sql.NullString
	for rows.Next() {
		switch t {
		case DBInt, DBIntList:
			if err := rows.Scan(&intResult); err != nil {
				return nil,
					Fail("cannot find int values for %s %d %s: %s", table, item, field, err)
			}
			if !intResult.Valid {
				return nil, Fail("no int value for %s %d %s", table, item, field)
			}
			results = append(results, NewInt(intResult.Int64))
		case DBString, DBStringList:
			if err := rows.Scan(&strResult); err != nil {
				return nil,
					Fail("cannot find string values for %s %d %s: %s", table, item, field, err)
			}
			if !strResult.Valid {
				return nil,
					Fail("no string value for %s %d %s", table, item, field)
			}
			results = append(results, NewString(strResult.String))
		case DBBlob, DBBlobList:
			if err := rows.Scan(&strResult); err != nil {
				return nil,
					Fail("cannot find string values for %s %d %s: %s", table, item, field, err)
			}
			if !strResult.Valid {
				return nil,
					Fail("no string value for %s %d %s", table, item, field)
			}
			b := []byte(strResult.String)
			results = append(results, NewBytes(b))
		case DBDate, DBDateList:
			if err := rows.Scan(&strResult); err != nil {
				return nil, Fail("cannot find date values for %s %d %s: %s", table, item, field, err)
			}
			if !strResult.Valid {
				return nil,
					Fail("no date value for %s %d %s", table, item, field)
			}
			t, err := ParseTime(strResult.String)
			if err != nil {
				return nil, Fail("invalid date representation in %s %d %s", table, item, field)
			}
			results = append(results, NewDate(t))
		default:
			return nil,
				Fail("cannot find values for %s %d %s: unknown field type %d (version too low?)",
					table, item, field, t)
		}
	}
	if err := rows.Err(); err != nil {
		return nil,
			Fail("cannot find values for %s %d %s: %s", table, item, field, err)
	}
	return results, nil
}

// Set the given values in the item in table and given field. An error is returned
// if the field types don't match the data.
func (tx *Tx) Set(table string, item Item, field string, data []Value) error {
	if !validTable.MatchString(table) {
		return Fail("invalid table name '%s'", table)
	}
	if !tx.mdb.TableExists(table) {
		return Fail("table '%s' does not exist", table)
	}
	if !tx.mdb.FieldExists(table, field) {
		return Fail("field '%s' does not exist in table '%s'", field, table)
	}
	if !tx.mdb.ItemExists(table, item) {
		return Fail("no %s %d", table, item)
	}
	t := ToBaseType(tx.mdb.MustGetFieldType(table, field))
	for i := range data {
		if data[i].Sort != t {
			return Fail("type error %s %d %s: expected %s, encountered %s",
				table, item, field, GetUserTypeString(t), GetUserTypeString(data[i].Sort))
		}
	}
	if tx.mdb.IsListField(table, field) {
		return tx.setListFields(table, item, field, data)
	}
	if len(data) > 1 {
		return Fail("attempt to set %d values in single field %s %d %s, should be just one value",
			len(data), table, item, field)
	}
	return tx.setSingleField(table, item, field, data[0])
}

func (tx *Tx) setSingleField(table string, item Item, field string, datum Value) error {
	var err error
	switch datum.Sort {
	case DBInt:
		_, err = tx.tx.Exec(fmt.Sprintf(`UPDATE "%s" SET "%s" = ? WHERE Id=?;`, table, field), datum.Int(), item)
	case DBBlob:
		_, err = tx.tx.Exec(fmt.Sprintf(`UPDATE "%s" SET "%s" = ? WHERE Id=?`, table, field),
			datum.Bytes(), item)
	default:
		_, err = tx.tx.Exec(fmt.Sprintf(`UPDATE "%s" SET "%s" = ? WHERE Id=?;`, table, field), datum.String(), item)
	}
	return err
}

func (tx *Tx) setListFields(table string, item Item, field string, data []Value) error {
	var err error
	tableName := listFieldToTableName(table, field)
	if !tx.mdb.TableExists(tableName) {
		return Fail("internal error, table %s does not exist (database has been tampered)",
			tableName)
	}
	_, err = tx.tx.Exec(fmt.Sprintf(`DELETE FROM %s WHERE Owner=?`, tableName), item)
	if err != nil {
		return err
	}

	for i := range data {
		switch data[i].Sort {
		case DBInt:
			_, err = tx.tx.Exec(fmt.Sprintf(`INSERT INTO %s(%s,Owner) VALUES(?,?)`, tableName, field),
				data[i].Int(), item)
		case DBBlob:
			_, err = tx.tx.Exec(fmt.Sprintf(`INSERT INTO %s(%s,Owner) VALUES(?,?)`, tableName, field),
				data[i].Bytes(), item)
		default:
			_, err = tx.tx.Exec(fmt.Sprintf(`INSERT INTO %s(%s,Owner) VALUES(?,?)`, tableName, field),
				data[i].String(), item)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// GetFields returns the fields that belong to a table, including list fields.
func (db *MDB) GetFields(table string) ([]Field, error) {
	if !db.TableExists(table) {
		return nil, Fail("table '%s' does not exist", table)
	}
	id, err := db.getTableId(table)
	if err != nil {
		return nil, err
	}
	rows, err := db.base.Query(`SELECT Name,FieldType FROM _COLS WHERE Owner=?;`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Field, 0)
	for rows.Next() {
		var s string
		var n int64
		if err := rows.Scan(&s, &n); err == nil {
			result = append(result, Field{s, FieldType(n)})
		} else {
			return nil, err
		}
	}
	return result, nil
}

// GetTables returns the tables in the database.
func (db *MDB) GetTables() []string {
	result := make([]string, 0)
	rows, err := db.base.Query(`SELECT Name FROM _TABLES;`)
	if err != nil {
		return result
	}
	defer rows.Close()
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err == nil {
			if len(s) > 0 && s[:1] != `_` {
				result = append(result, s)
			}
		}
	}
	return result
}

// -------
// Queries
// -------

// QuerySort is the type of a query. Some of these sorts are only used internally.
type QuerySort int

// The sorts of query entries.
const (
	// ParseError indicates that query parsing was unsuccessful.
	ParseError QuerySort = iota + 1
	// TableString is the type of a string for a table, e.g. "Person".
	TableString
	// QueryString is the type of a string for a query, i.e., what's right of "=".
	QueryString
	// LogicalAnd is the type of "and".
	LogicalAnd
	// LogicalOr is the type of "or".
	LogicalOr
	// LogicalNot is the type of "not".
	LogicalNot
	// FieldString is the type of a string for a field, e.g. "name".
	FieldString
	// SearchClause is the type of a main search clause (outermost level of a query).
	SearchClause
	// NoTerm is the type of "no" in a list-field query like "Person no name=John".
	NoTerm
	// EveryTerm is the type of "every" in a list-field query like "Person every name=%s%".
	EveryTerm
	// LeftParen is the type of "(" and variants, used internally.
	LeftParen
	// RightParen is the type of ")" and variants, used internally.
	RightParen
	// InfixOP is the type of "=".
	InfixOP
)

// QuerySortToStr convert the sort of a query to a string. This is merely used for debugging and testing.
func QuerySortToStr(s QuerySort) string {
	switch s {
	case ParseError:
		return "<parse-error>"
	case TableString:
		return "TableString"
	case QueryString:
		return "QueryString"
	case LogicalAnd:
		return "LogicalAnd"
	case LogicalOr:
		return "LogicalOr"
	case LogicalNot:
		return "LogicalNot"
	case FieldString:
		return "FieldString"
	case SearchClause:
		return "SearchClause"
	case NoTerm:
		return "NoTerm"
	case EveryTerm:
		return "EveryTerm"
	case LeftParen:
		return "LeftParen"
	case RightParen:
		return "RightParen"
	case InfixOP:
		return "InfixOP"
	default:
		return "<unknown>"
	}
}

// Query represents a simple or complex database query.
type Query struct {
	Sort     QuerySort `json:"sort"`
	Children []Query   `json:"children"`
	Data     string    `json:"data"`
}

// DebugDump returns a string representation of a query. This is used for debugging and testing, the result
// is neither pretty-printed nor intended for human consumption.
func (q *Query) DebugDump() string {
	if q == nil {
		return "<nil>"
	}
	s := ""
	if q.Children != nil {
		for i, c := range q.Children {
			if i > 0 {
				s = s + ","
			}
			s = s + c.DebugDump()
		}
	}
	return fmt.Sprintf(`%s("%s",[%s])`, QuerySortToStr(q.Sort), q.Data, s)
}

// FailedQuery returns a failed Query pointer with the given message as explanation
// why it failed.
func FailedQuery(msg string) *Query {
	return &Query{ParseError, nil, msg}
}

func fPrintEscape(s string) string {
	return strings.Replace(s, "%", "%%", -1)
}

func blobQueryEscape(s string) string {
	return strings.Replace(strings.Replace(s, `\`, `\\`, -1),
		`%%`, `\%%`, -1)
}

type fieldDesc struct {
	name     string
	paramN   int
	joined   []bool
	paramIdx int
}

// Return the LIKE part of a query, but not the select statement or joins needed
// for list fields. Use db.ToSql to get the whole query including joins and select statement.
//
// To understand this helper function, look at the way queries are created by ToSql.
//
// Example outputs: "(Name LIKE 'John%') AND (Address LIKE '%New York%')"
// "NOT EXISTS (SELECT 1 FROM _Person_Name AS <P1> WHERE <P1>.Name LIKE 'John' AND Person.Id=<P1>.Owner)"
// The field descriptions contain information about the number of parameters of the form
// "<P1>", "<P2>", ..., to replace in the final result. If joined is true for such a parameter
// an INNER JOIN table will be created. (EVERY and NO operators need an additional param but only one join.)
func (db *MDB) toSqlSearchTerm(q *Query, table string,
	fieldDescs *[]fieldDesc, paramStartIdx *int) (string, error) {
	switch (*q).Sort {

	case InfixOP:
		if (*q).Children == nil {
			return "", Fail("empty clause")
		}
		if len((*q).Children) < 2 {
			return "", Fail("missing argument in clause")
		}
		if len((*q).Children) > 2 {
			return "", Fail("too many arguments in clause")
		}
		if (*q).Children[0].Sort != FieldString {
			return "", Fail("first part of a clause must be the field")
		}
		if (*q).Children[1].Sort != QueryString {
			return "", Fail("second part of a clause must be the search term")
		}
		fieldName, err := db.toSqlSearchTerm(&(*q).Children[0], table, fieldDescs, paramStartIdx)
		if err != nil {
			return "", err
		}
		if !db.FieldExists(table, fieldName) {
			return "", Fail("field '%s' does not exist in table '%s'", fieldName, table)
		}
		searchTerm, err := db.toSqlSearchTerm(&(*q).Children[1], table, fieldDescs, paramStartIdx)
		if err != nil {
			return "", Fail("syntax error in query: %s", err)
		}
		*paramStartIdx++
		*fieldDescs = append(*fieldDescs, fieldDesc{fieldName, 1, []bool{true}, *paramStartIdx})
		sort := db.MustGetFieldType(table, fieldName)
		switch sort {
		case DBInt, DBIntList:
			return `CAST(<P` + strconv.Itoa(*paramStartIdx) + `>.` + fmt.Sprintf(`%s AS TEXT) LIKE '%s'`, fieldName, searchTerm), nil
		case DBBlob, DBBlobList:
			return `<P` + strconv.Itoa(*paramStartIdx) + `>.` + fmt.Sprintf(`%s LIKE '%s' ESCAPE '\'`, fieldName,
				blobQueryEscape(searchTerm)), nil
		default:
			return `<P` + strconv.Itoa(*paramStartIdx) + `>.` + fmt.Sprintf(`%s LIKE '%s'`, fieldName, searchTerm), nil
		}

	case LogicalAnd, LogicalOr:
		var connective string
		if (*q).Sort == LogicalAnd {
			connective = "AND"
		} else {
			connective = "OR"
		}
		if len((*q).Children) < 2 {
			return "", Fail("missing argument")
		}
		if len((*q).Children) > 2 {
			return "", Fail("too many arguments")
		}
		clause1, err := db.toSqlSearchTerm(&(*q).Children[0], table, fieldDescs, paramStartIdx)
		if err != nil {
			return "", err
		}
		clause2, err := db.toSqlSearchTerm(&(*q).Children[1], table, fieldDescs, paramStartIdx)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(`(%s) %s (%s)`, clause1, connective, clause2), nil
	case LogicalNot:
		if len((*q).Children) != 1 {
			return "", Fail("NOT takes only one argument, given %d", len((*q).Children))
		}
		clause, err := db.toSqlSearchTerm(&(*q).Children[0], table, fieldDescs, paramStartIdx)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(`NOT (%s)`, clause), nil

	case NoTerm, EveryTerm:
		if len((*q).Children) == 0 {
			return "", Fail("missing argument")
		}
		if len((*q).Children) > 1 {
			return "", Fail("NO and EVERY take only one argument, given %d", len((*q).Children))
		}
		if len((*q).Children[0].Children) != 2 {
			return "", Fail("ill-formed NO or EVERY clause, expected field name and search term")
		}
		name := (*q).Children[0].Children[0].Data
		if !db.IsListField(table, name) {
			return "", Fail("not a list field '%s', NO and EVERY can only be applied to list fields", name)
		}
		switch (*q).Sort {
		case NoTerm, EveryTerm:
			*paramStartIdx++
			searchTerm := (*q).Children[0].Children[1].Data
			*fieldDescs = append(*fieldDescs, fieldDesc{name, 2, []bool{true, false}, *paramStartIdx})
			paramStr := "<P" + strconv.Itoa(*paramStartIdx) + ">"
			maybeNegated := ""
			if (*q).Sort == EveryTerm {
				maybeNegated = " NOT"
			}
			return fmt.Sprintf("NOT EXISTS (SELECT 1 FROM %s AS "+paramStr+" WHERE "+paramStr+".%s"+maybeNegated+" LIKE '%s' AND %s.Id="+paramStr+".Owner)", listFieldToTableName(table, name), name, searchTerm, table), nil

		default:
			return "", Fail("unsupported search modifier %d (version too low?)", int((*q).Sort))
		}

	case FieldString:
		if !validFieldName.MatchString((*q).Data) {
			return "", Fail("invalid field name '%s'", (*q).Data)
		}
		return (*q).Data, nil

	case QueryString:
		return fPrintEscape((*q).Data), nil
	default:
		return "", Fail("unsupported query element %d (version too low?)", int((*q).Sort))
	}
}

// ToSql returns the sql query for the table, taking into account list fields,
// or returns an error if the query structure is ill-formed.
func (db *MDB) ToSql(table string, inquery *Query, limit int64) (string, error) {
	if !db.TableExists(table) {
		return "", Fail("table '%s' does not exist", table)
	}
	// check if the query is embedded into a search clause
	// if so, we check against the table name and remove the outer layer
	// since toSqlSearchTerm works on the basis of a known and validated table
	var query *Query
	if inquery.Sort == SearchClause {
		if inquery.Data != table {
			return "", Fail("query with SearchClause for table '%s' requested for table '%s'", inquery.Data, table)
		}
		if len(inquery.Children) != 1 {
			return "", Fail("query with malformed SearchClause, it should have one child node but contains %d",
				len(inquery.Children))
		}
		query = &inquery.Children[0]
	} else {
		query = inquery
	}
	fieldDescs := make([]fieldDesc, 0)
	c := 0
	condition, err := db.toSqlSearchTerm(query, table, &fieldDescs, &c)
	if err != nil {
		return "", err
	}
	for _, field := range fieldDescs {
		if !db.FieldExists(table, field.name) {
			return "", Fail("invalid query, %s %s field does not exist", table, field.name)
		}
	}
	joins := ""
	for _, field := range fieldDescs {
		j := 0
		for i := field.paramIdx; i < field.paramIdx+field.paramN; i++ {
			toReplace := "<P" + strconv.Itoa(i) + ">"
			if db.IsListField(table, field.name) {
				if field.joined[j] {
					joins += fmt.Sprintf(" INNER JOIN %s AS __T%d ON %s.Id = __T%d.Owner",
						listFieldToTableName(table, field.name), i, table, i)
				}
				condition = strings.Replace(condition, toReplace, "__T"+strconv.Itoa(i), -1)
			} else {
				condition = strings.Replace(condition, toReplace, table, -1)
			}
			j++
		}
	}
	var result string
	if limit > 0 {
		result = fmt.Sprintf("SELECT DISTINCT %s.Id FROM %s%s WHERE %s LIMIT %d;",
			table, table, joins, condition, limit)
	} else {
		result = fmt.Sprintf("SELECT DISTINCT %s.Id FROM %s%s WHERE %s;",
			table, table, joins, condition)
	}
	return result, nil
}

// Find items matching the query, return error if the query is ill-formed
// and the items otherwise.
func (db *MDB) Find(query *Query, limit int64) ([]Item, error) {
	result := make([]Item, 0)
	table := (*query).Data
	if len((*query).Children) == 0 {
		return result, Fail("incomplete query, only table given")
	}
	query = &query.Children[0]
	toExec, err := db.ToSql(table, query, limit)
	//fmt.Println(toExec) // the final query, for debugging
	if err != nil {
		return result, Fail("invalid query - %s", err)
	}
	if !db.TableExists(table) {
		return result, Fail("invalid query - table '%s' does not exist", table)
	}

	var rows *sql.Rows
	rows, err = db.base.Query(toExec)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var datum sql.NullInt64
		if err := rows.Scan(&datum); err == nil && datum.Valid {
			result = append(result, Item(datum.Int64))
		}
	}
	return result, nil
}
