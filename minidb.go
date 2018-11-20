// Minidb is a minimalist database. It stores items in tables, where each item has a fixed number of fields.
package minidb

import (
	"database/sql"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/rasteric/minidb/parser"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	_ "github.com/mattn/go-sqlite3"
)

// The main database object.
type MDB struct {
	base *sql.DB
}

// A database item. Fields and tables are identified by strings.
type Item int64

// FieldType is a field in the database, which might be a list type or a base type.
type FieldType int

const (
	DBError FieldType = iota + 1
	DBInt
	DBString
	DBBlob
	DBIntList
	DBStringList
	DBBlobList
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
	default:
		return t
	}
}

type Field struct {
	Name string
	Sort FieldType
}

func failure(msg string, args ...interface{}) error {
	return errors.New(fmt.Sprintf(msg, args...))
}

// Value holds the values that can be put into the database or retrieved from it.
type Value struct {
	str  string
	num  int64
	sort FieldType
}

// Int returns the value as an int64 and panics if conversion is not possible
func (v *Value) Int() int64 {
	switch v.sort {
	case DBInt:
		return v.num
	default:
		panic(fmt.Sprintf("cannot convert %s value to integer",
			GetUserTypeString(v.sort)))
	}
}

// String returns the string value. It automatically converts int and blob,
// where binary Blob data is Base64 encoded and the int is converted
// to decimal format.
func (v *Value) String() string {
	switch v.sort {
	case DBInt:
		return fmt.Sprintf("%d", v.num)
	case DBString:
		return v.str
	case DBBlob:
		return base64.StdEncoding.EncodeToString([]byte(v.str))
	default:
		panic(fmt.Sprintf("cannot convert %s value to string",
			GetUserTypeString(v.sort)))
	}
}

// Bytes returns the value as a bytes slice. It automatically converts int64 and string.
// An int64 is written in Little Endian format.
func (v *Value) Bytes() []byte {
	switch v.sort {
	case DBInt:
		bs := make([]byte, 8)
		binary.LittleEndian.PutUint64(bs, uint64(v.num))
		return bs
	case DBString:
		bs := []byte(v.str)
		return bs
	case DBBlob:
		bs := []byte(v.str)
		return bs
	default:
		panic(fmt.Sprintf("cannot convert %s value to bytes",
			GetUserTypeString(v.sort)))
	}
}

// Sort returns the FieldType of the value stored. The field type should never be a list type,
// because list types are slices of type []Value.
func (v *Value) Sort() FieldType {
	return v.sort
}

// NewInt creates a value that stores an int64.
func NewInt(n int64) Value {
	return Value{num: n, sort: DBInt}
}

// NewString creates a value that stores a string.
func NewString(s string) Value {
	return Value{str: s, sort: DBString}
}

// NewBytes creates a value that holds a []byte slice. This is similar to String() but notice
// that strings and byte slices are handled differently in the database. For example,
// byte slices may contain NULL characters and may be converted to and from Base64.
func NewBytes(b []byte) Value {
	return Value{str: string(b), sort: DBBlob}
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
	case DBStringList, DBIntList, DBBlobList:
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
	default:
		return "INTEGER"
	}
}

// Return a user-readable string for the type of a field.
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
	case "string-list", "str-list", "text-list", "txt-list":
		return DBStringList, nil
	case "blob-list":
		return DBBlobList, nil
	case "int-list", "integer-list":
		return DBIntList, nil
	}
	return DBError,
		failure("Invalid field type '%s', should be one of int,string,blob,int-list,string-list,blob-list", ident)
}

// ParseFieldDesc parses the given string slice into a []Field slice based on
// the format "type name", or returns an error. This can be used for command line parsing.
func ParseFieldDesc(desc []string) ([]Field, error) {
	result := make([]Field, 0)
	if len(desc)%2 != 0 {
		return nil, failure("invalid field descriptions, they must be of the form <type> <fieldname>!")
	}
	if len(desc) == 0 {
		return nil, failure("no fields specified!")
	}
	for i := 0; i < len(desc)-1; i += 2 {
		ftype, err := parseFieldType(desc[i])
		if err != nil {
			return nil, err
		}
		if !validFieldName.MatchString(desc[i+1]) {
			return nil, failure("invalid field name '%s'", desc[i+1])
		}
		if strings.ToLower(desc[i+1]) == "id" {
			return nil, failure("fields may not be called 'id'!")
		}
		result = append(result, Field{desc[i+1], ftype})
	}
	return result, nil
}

var errNilDB = failure("db object is nil")

func (db *MDB) init() error {
	if db.base == nil {
		return errNilDB
	}
	_, err := db.base.Exec(`CREATE TABLE IF NOT EXISTS _TABLES (Id INTEGER PRIMARY KEY,
Name TEXT NOT NULL)`)
	if err != nil {
		return err
	}
	_, err = db.base.Exec(`CREATE INDEX IF NOT EXISTS _TABIDX ON _TABLES (Name)`)
	if err != nil {
		return err
	}
	_, err = db.base.Exec(`CREATE TABLE IF NOT EXISTS _COLS (Id INTEGER PRIMARY KEY,
	 Name STRING NOT NULL,
	 FieldType INTEGER NOT NULL,
	Owner INTEGER NOT NULL,
	FOREIGN KEY(Owner) REFERENCES _TABLES(Id))`)
	return err
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
	db.base = base
	if err := db.init(); err != nil {
		return nil, failure("cannot initialize database: %s", err)
	}
	return db, nil
}

func (db *MDB) Close() error {
	if db.base != nil {
		err := db.base.Close()
		if err != nil {
			return failure("ERROR Failed to close database - %s.\n", err)
		}
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

// FieldIsNull returns true if the field is not null for the item in the table, false otherwise.
func (db *MDB) FieldIsNull(table string, item Item, field string) bool {
	var result int
	err := db.base.QueryRow(fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM "%s" WHERE ? IS NULL and Id=?)`, table), field, item).Scan(&result)
	if err != nil {
		return false
	}
	return result > 0
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

// IsEmptyListField returns true if the field has at least one result matching item.
func (db *MDB) IsEmptyListField(table string, item Item, field string) bool {
	var result int
	err := db.base.QueryRow(fmt.Sprintf(`SELECT COUNT (%s) FROM "%s" WHERE Owner=? AND %s IS NOT NULL LIMIT 1`, field, table, field), item).Scan(&result)
	if err != nil {
		return false
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

// IsListField is true if the field in the table is a list, false otherwise.
// List fields are internally stored as special tables.
func (db *MDB) IsListField(table string, field string) bool {
	return db.TableExists(listFieldToTableName(table, field))
}

// ParseFieldValues parses potential value(s) for a field from strings, returns an error if their type is
// incompatible with the field type or the table or field don't exist.
// Data for a Blob field must be Base64 encoded, data for an Integer field must be a valid
// digit sequence for a 64 bit integer in base 10 format.
func (db *MDB) ParseFieldValues(table string, field string, data []string) ([]Value, error) {
	if !validTable.MatchString(table) {
		return nil, failure("invalid table name '%s'", table)
	}
	if !db.TableExists(table) {
		return nil, failure("table '%s' does not exist", table)
	}
	if !db.FieldExists(table, field) {
		return nil, failure("field '%s' does not exist in table '%s'", field, table)
	}
	if len(data) == 0 {
		return nil, failure("no input values given")
	}
	if !db.IsListField(table, field) && len(data) > 1 {
		return nil, failure("too many input values: expected 1, given %d", len(data))
	}
	t := ToBaseType(db.MustGetFieldType(table, field))
	result := make([]Value, 0, len(data))
	for i, _ := range data {
		switch t {
		case DBInt:
			j, err := strconv.ParseInt(data[i], 10, 64)
			if err != nil {
				return nil, failure("type error: expected int, given '%s'", data[i])
			}
			result = append(result, NewInt(j))
		case DBBlob:
			b, err := base64.StdEncoding.DecodeString(data[i])
			if err != nil {
				return nil, failure("type error: expected binary data in Base64 format but the given data  seems invalid")
			}
			result = append(result, NewBytes(b))
		case DBString:
			result = append(result, NewString(data[i]))
		default:
			return nil,
				failure("internal error: %s %s expects type %d, which is unknown to this version of minidb",
					table, field, t)
		}
	}
	return result, nil
}

// AddTable is used to create a new table. Table and field names are validated. They need to be alphanumeric
// sequences plus underscore "_" as the only allowed special character. None of the names may start with
// an underscore.
func (db *MDB) AddTable(table string, fields []Field) error {
	if !validTable.MatchString(table) {
		return failure("invalid table name '%s'", table)
	}
	// normal fields are just columns
	toExec := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s" (Id INTEGER PRIMARY KEY`, table)
	for _, field := range fields {
		if !isListFieldType(field.Sort) {
			toExec += fmt.Sprintf(",\n\"%s\" %s", field.Name, getTypeString(field.Sort))
		}
	}
	toExec += ");"
	_, err := db.base.Exec(toExec)
	if err != nil {
		return failure("cannot create maintenance table: %s", err)
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
			_, err = db.base.Exec(toExec)
			if err != nil {
				return failure("cannot create list field %s in table %s: %s", field.Name, table, err)
			}
		}
	}
	// update the internal housekeeping tables
	toExec = "INSERT OR IGNORE INTO _TABLES (Name) VALUES (?)"
	result, err := db.base.Exec(toExec, table)
	if err != nil {
		return failure("failureed to update maintenance table: %s", err)
	}
	tableID, err := result.LastInsertId()
	if err != nil {
		return failure("failed to update maintenance table: %s", err)
	}
	for _, field := range fields {
		_, err := db.base.Exec(`INSERT INTO _COLS (Name,FieldType,Owner) VALUES (?,?,?)`, field.Name, field.Sort, tableID)
		if err != nil {
			return failure("cannot insert maintenance field %s for table %s: %s",
				field.Name, table, err)
		}
		if isListFieldType(field.Sort) {
			_, err = db.base.Exec(`INSERT INTO _TABLES (Name) VALUES (?)`, listFieldToTableName(table, field.Name))
		}
		if err != nil {
			return failure("cannot insert maintenance list table %s for table %s: %s",
				listFieldToTableName(table, field.Name), table, err)
		}
	}
	return nil
}

// NewItem creates a new item in the table and returns its numerical ID.
func (db *MDB) NewItem(table string) (Item, error) {
	if !validTable.MatchString(table) {
		return 0, failure("invalid table name '%s'", table)
	}
	if !db.TableExists(table) {
		return 0, failure("table '%s' does not exist", table)
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

// Count returns the number of items in the table.
func (db *MDB) Count(table string) (int64, error) {
	if !validTable.MatchString(table) {
		return 0, failure("invalid table name '%s'", table)
	}
	if !db.TableExists(table) {
		return 0, failure("table '%s' does not exist", table)
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
		return empty, failure("invalid table name '%s'", table)
	}
	if !db.TableExists(table) {
		return empty, failure("table '%s' does not exist", table)
	}
	rows, err := db.base.Query(fmt.Sprintf(`SELECT (Id) FROM %s;`, table))
	if err != nil {
		return empty, err
	}
	defer rows.Close()
	results := make([]Item, 0)
	var c int64 = 0
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
		return nil, failure("invalid table name '%s'", table)
	}
	if !db.TableExists(table) {
		return nil, failure("table '%s' does not exist", table)
	}
	if !db.ItemExists(table, item) {
		return nil, failure("no %s %d", table, item)
	}
	if db.IsListField(table, field) {
		return db.getListField(table, item, field)
	} else {
		return db.getSingleField(table, item, field)
	}
}

func (db *MDB) getSingleField(table string, item Item, field string) ([]Value, error) {
	if !db.FieldExists(table, field) {
		return nil,
			failure(`no field %s in table %s`, field, table)
	}
	if db.FieldIsNull(table, item, field) {
		return nil,
			failure(`no value for %s %d %s`, table, item, field)
	}
	t := db.MustGetFieldType(table, field)
	row := db.base.QueryRow(fmt.Sprintf(`SELECT "%s" FROM "%s" WHERE Id=?;`, field, table), item)
	var intResult sql.NullInt64
	var strResult sql.NullString
	var err error
	switch t {
	case DBInt:
		err = row.Scan(&intResult)
	case DBString, DBBlob:
		err = row.Scan(&strResult)
	default:
		return nil,
			failure("unsupported field type for %s %d %s: %d (try a newer version?)", table, item, field, int(t))
	}
	if err == sql.ErrNoRows {
		return nil,
			failure("no value for %s %d %s", table, item, field)
	}
	if err != nil {
		return nil,
			failure("cannot find value for %s %d %s: %s", table, item, field, err)
	}
	vslice := make([]Value, 1)
	switch t {
	case DBInt:
		if !intResult.Valid {
			return nil,
				failure("no int value for %s %d %s", table, item, field)
		}
		vslice[0] = NewInt(intResult.Int64)
	case DBString, DBBlob:
		if !strResult.Valid {
			return nil,
				failure("no string or blob value for %s %d %s", table, item, field)
		}
		vslice[0] = NewString(strResult.String)
	}
	return vslice, nil
}

func (db *MDB) getListField(table string, item Item, field string) ([]Value, error) {
	tableName := listFieldToTableName(table, field)
	if !db.TableExists(tableName) {
		return nil,
			failure("list field %s does not exist in table %s", field, table)
	}
	if db.IsEmptyListField(tableName, item, field) {
		return nil,
			failure("no values for %s %d %s", table, item, field)
	}
	t := db.MustGetFieldType(table, field)
	rows, err := db.base.Query(fmt.Sprintf(`SELECT %s FROM "%s" WHERE Owner=?`, field, tableName), item)
	if err != nil {
		return nil,
			failure("cannot find values for %s %d %s: %s", table, item, field, err)
	}
	results := make([]Value, 0)
	var intResult sql.NullInt64
	var strResult sql.NullString
	for rows.Next() {
		switch t {
		case DBInt, DBIntList:
			if err := rows.Scan(&intResult); err != nil {
				rows.Close()
				return nil,
					failure("cannot find int values for %s %d %s: %s", table, item, field, err)
			}
			if !intResult.Valid {
				rows.Close()
				return nil, failure("no int value for %s %d %s", table, item, field)
			}
			results = append(results, NewInt(intResult.Int64))
		case DBString, DBStringList:
			if err := rows.Scan(&strResult); err != nil {
				rows.Close()
				return nil,
					failure("cannot find string values for %s %d %s: %s", table, item, field, err)
			}
			if !strResult.Valid {
				rows.Close()
				return nil,
					failure("no string value for %s %d %s", table, item, field)
			}
			results = append(results, NewString(strResult.String))
		case DBBlob, DBBlobList:
			if err := rows.Scan(&strResult); err != nil {
				rows.Close()
				return nil,
					failure("cannot find string values for %s %d %s: %s", table, item, field, err)
			}
			if !strResult.Valid {
				rows.Close()
				return nil,
					failure("no string value for %s %d %s", table, item, field)
			}
			b := []byte(strResult.String)
			results = append(results, NewBytes(b))
		default:
			rows.Close()
			return nil,
				failure("cannot find values for %s %d %s: unknown field type %d (version too low?)",
					table, item, field, t)
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil,
			failure("cannot find values for %s %d %s: %s", table, item, field, err)
	}
	return results, nil
}

// Set the given values in the item in table and given field. An error is returned
// if the field types don't match the data.
func (db *MDB) Set(table string, item Item, field string, data []Value) error {
	if !validTable.MatchString(table) {
		return failure("invalid table name '%s'", table)
	}
	if !db.TableExists(table) {
		return failure("table '%s' does not exist", table)
	}
	if !db.ItemExists(table, item) {
		return failure("no %s %d", table, item)
	}
	if len(data) == 0 {
		return failure("no value given to set in %s %d %s", table, item, field)
	}
	t := ToBaseType(db.MustGetFieldType(table, field))
	for i, _ := range data {
		if data[i].Sort() != t {
			return failure("type error %s %d %s: expected %d, encountered %d",
				table, item, field, t, data[i].Sort())
		}
	}
	if db.IsListField(table, field) {
		return db.setListFields(table, item, field, data)
	}
	if len(data) > 1 {
		return failure("attempt to set %d values in single field %s %d %s, should be just one value",
			len(data), table, item, field)
	}
	return db.setSingleField(table, item, field, data[0])
}

func (db *MDB) setSingleField(table string, item Item, field string, datum Value) error {
	switch datum.Sort() {
	case DBInt:
		_, err := db.base.Exec(fmt.Sprintf(`UPDATE %s SET %s = ? WHERE Id=?;`, table, field), datum.Int(), item)
		return err
	default:
		_, err := db.base.Exec(fmt.Sprintf(`UPDATE %s SET %s = ? WHERE Id=?;`, table, field), datum.String(), item)
		return err
	}
}

func (db *MDB) setListFields(table string, item Item, field string, data []Value) error {
	var err error
	tableName := listFieldToTableName(table, field)
	if !db.TableExists(tableName) {
		return failure("internal error, table %s does not exist (database has been tampered)",
			tableName)
	}
	_, err = db.base.Exec(fmt.Sprintf(`DELETE FROM %s WHERE Owner=?`, tableName), item)
	if err != nil {
		return err
	}
	for i, _ := range data {
		switch data[i].Sort() {
		case DBInt:
			_, err = db.base.Exec(fmt.Sprintf(`INSERT INTO %s(%s,Owner) VALUES(?,?)`, tableName, field),
				data[i].Int(), item)
		default:
			_, err = db.base.Exec(fmt.Sprintf(`INSERT INTO %s(%s,Owner) VALUES(?,?)`, tableName, field),
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
		return nil, failure("table '%s' does not exist", table)
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

// Get the tables in the database.
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

type QuerySort int

const (
	ParseError QuerySort = iota + 1
	Term
	LogicalAnd
	LogicalOr
	LogicalNot
	Clause
	FieldName
	SearchClause
	NoTerm
	EveryTerm
)

type Query struct {
	Sort     QuerySort
	Children []Query
	data     string
}

func FailedQuery(msg string) *Query {
	return &Query{ParseError, nil, msg}
}

func fPrintEscape(s string) string {
	return strings.Replace(s, "%", "%%", -1)
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
	case Clause:
		if (*q).Children == nil {
			return "", failure("empty clause")
		}
		if len((*q).Children) < 2 {
			return "", failure("missing argument in clause")
		}
		if len((*q).Children) > 2 {
			return "", failure("too many arguments in clause")
		}
		if (*q).Children[0].Sort != FieldName {
			return "", failure("first part of a clause must be the field")
		}
		if (*q).Children[1].Sort != Term {
			return "", failure("second part of a clause must be the search term")
		}
		fieldName, err := db.toSqlSearchTerm(&(*q).Children[0], table, fieldDescs, paramStartIdx)
		if err != nil {
			return "", err
		}
		if !db.FieldExists(table, fieldName) {
			return "", failure("field '%s' does not exist in table '%s'", fieldName, table)
		}
		searchTerm, err := db.toSqlSearchTerm(&(*q).Children[1], table, fieldDescs, paramStartIdx)
		*paramStartIdx++
		*fieldDescs = append(*fieldDescs, fieldDesc{fieldName, 1, []bool{true}, *paramStartIdx})
		sort := db.MustGetFieldType(table, fieldName)
		switch sort {
		case DBInt, DBIntList:
			return `CAST(<P` + strconv.Itoa(*paramStartIdx) + `>.` + fmt.Sprintf(`%s AS TEXT) LIKE '%s'`, fieldName, searchTerm), nil
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
			return "", failure("missing argument")
		}
		if len((*q).Children) > 2 {
			return "", failure("too many arguments")
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
			return "", failure("NOT takes only one argument, given %d", len((*q).Children))
		}
		clause, err := db.toSqlSearchTerm(&(*q).Children[0], table, fieldDescs, paramStartIdx)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(`NOT (%s)`, clause), nil
	case NoTerm, EveryTerm:
		if len((*q).Children) == 0 {
			return "", failure("missing argument")
		}
		if len((*q).Children) > 1 {
			return "", failure("NO and EVERY take only one argument, given %d", len((*q).Children))
		}
		if len((*q).Children[0].Children) != 2 {
			return "", failure("ill-formed NO or EVERY clause, expected field name and search term")
		}
		name := (*q).Children[0].Children[0].data
		if !db.IsListField(table, name) {
			return "", failure("not a list field '%s', NO and EVERY can only be applied to list fields", name)
		}
		switch (*q).Sort {
		case NoTerm, EveryTerm:
			*paramStartIdx++
			searchTerm := (*q).Children[0].Children[1].data
			*fieldDescs = append(*fieldDescs, fieldDesc{name, 2, []bool{true, false}, *paramStartIdx})
			paramStr := "<P" + strconv.Itoa(*paramStartIdx) + ">"
			maybeNegated := ""
			if (*q).Sort == EveryTerm {
				maybeNegated = " NOT"
			}
			return fmt.Sprintf("NOT EXISTS (SELECT 1 FROM %s AS "+paramStr+" WHERE "+paramStr+".%s"+maybeNegated+" LIKE '%s' AND %s.Id="+paramStr+".Owner)", listFieldToTableName(table, name), name, searchTerm, table), nil

		default:
			return "", failure("unsupported search modifier %d (version too low?)", int((*q).Sort))
		}
	case FieldName:
		if !validFieldName.MatchString((*q).data) {
			return "", failure("invalid field name '%s'", (*q).data)
		}
		return (*q).data, nil
	case Term:
		return fPrintEscape((*q).data), nil
	default:
		return "", failure("unsupported query element %d (version too low?)", int((*q).Sort))
	}
}

// ToSql returns the sql query for the table, taking into account list fields,
// or returns an error if the query structure is ill-formed.
func (db *MDB) ToSql(table string, query *Query, limit int64) (string, error) {
	if !db.TableExists(table) {
		return "", failure("table '%s' does not exist", table)
	}
	fieldDescs := make([]fieldDesc, 0)
	c := 0
	condition, err := db.toSqlSearchTerm(query, table, &fieldDescs, &c)
	if err != nil {
		return "", err
	}
	for _, field := range fieldDescs {
		if !db.FieldExists(table, field.name) {
			return "", failure("invalid query, %s %s field does not exist", table, field.name)
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
	if limit > 0 {
		return fmt.Sprintf("SELECT DISTINCT %s.Id FROM %s%s WHERE %s LIMIT %d;", table, table, joins, condition, limit), nil
	} else {
		return fmt.Sprintf("SELECT DISTINCT %s.Id FROM %s%s WHERE %s;", table, table, joins, condition), nil
	}
}

// ------------------------------------
// Antlr-based Query expression parser
// ------------------------------------

type mdbListener struct {
	*parser.BaseMdbListener

	stack       *[]Query
	table       string
	parseFailed bool
}

func (l *mdbListener) Push(query *Query) {
	if l.stack == nil {
		s := make([]Query, 1, 100)
		s[0] = *query
		l.stack = &s
	} else {
		*l.stack = append(*l.stack, *query)
	}
}

func (l *mdbListener) Pop() *Query {
	if len(*l.stack) == 0 {
		l.parseFailed = true
		return FailedQuery("stack underflow while parsing expression")
	}
	result := (*l.stack)[len(*l.stack)-1]
	*l.stack = (*l.stack)[:len(*l.stack)-1]
	return &result
}

func (l *mdbListener) ExitSearchclause(c *parser.SearchclauseContext) {
	query := l.Pop()
	if query != nil {
		l.Push(&Query{SearchClause, []Query{*query}, l.table})
	}
}

func (l *mdbListener) ExitField(c *parser.FieldContext) {
	l.Push(&Query{FieldName, nil, c.GetText()})
}

func (l *mdbListener) ExitSearchterm(c *parser.SearchtermContext) {
	var s string
	if c.STRING() != nil {
		s = strings.Trim(c.GetText(), "\"")
	} else {
		s = c.GetText()
	}
	l.Push(&Query{Term, nil, s})
}

func (l *mdbListener) ExitFieldsearch(c *parser.FieldsearchContext) {
	rhs := l.Pop()
	lhs := l.Pop()
	if lhs != nil && rhs != nil {
		l.Push(&Query{Clause, []Query{*lhs, *rhs}, ""})
	}
}

func (l *mdbListener) ExitExpr(c *parser.ExprContext) {
	if c.Relop() != nil {
		s := strings.ToLower(c.Relop().GetText())
		switch s {
		case "and":
			rhs := l.Pop()
			lhs := l.Pop()
			if rhs != nil && lhs != nil {
				l.Push(&Query{LogicalAnd, []Query{*lhs, *rhs}, ""})
			}
		case "or":
			rhs := l.Pop()
			lhs := l.Pop()
			if rhs != nil && lhs != nil {
				l.Push(&Query{LogicalOr, []Query{*lhs, *rhs}, ""})
			}
		default:
			l.parseFailed = true
		}
	} else if c.Unop() != nil {
		switch strings.ToLower(c.Unop().GetText()) {
		case "not":
			arg := l.Pop()
			if arg != nil {
				l.Push(&Query{LogicalNot, []Query{*arg}, ""})
			}
		default:
			l.parseFailed = true
		}
	} else if c.Searchop() != nil {
		switch strings.ToLower(c.Searchop().GetText()) {
		case "no":
			arg := l.Pop()
			if arg != nil {
				l.Push(&Query{NoTerm, []Query{*arg}, ""})
			}
		case "every":
			arg := l.Pop()
			if arg != nil {
				l.Push(&Query{EveryTerm, []Query{*arg}, ""})
			}
		}
	}
}

func (l *mdbListener) ExitUnop(c *parser.UnopContext) {
}

func (l *mdbListener) ExitRelop(c *parser.RelopContext) {
}

func (l *mdbListener) ExitTable(c *parser.TableContext) {
	l.table = c.GetText()
}

func NewMdbListener() *mdbListener {
	return new(mdbListener)
}

//  ParseQuery parses a string consisting of table name+query clauses into a Query
// structure for processing by ToSql.
func ParseQuery(s string) (*Query, error) {
	// Setup the input
	is := antlr.NewInputStream(s)

	// Create the Lexer
	lexer := parser.NewMdbLexer(is)

	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	// // Print the lexer tokens:
	// for {
	// 	t := lexer.NextToken()
	// 	if t.GetTokenType() == antlr.TokenEOF {
	// 		break
	// 	}
	// 	fmt.Printf("%s (%q)\n",
	// 		lexer.SymbolicNames[t.GetTokenType()], t.GetText())
	// }

	// Create the Parser
	p := parser.NewMdbParser(stream)
	p.BuildParseTrees = true
	// Finally parse the expression
	tree := p.Start()
	listener := NewMdbListener()
	antlr.ParseTreeWalkerDefault.Walk(listener, tree)
	if listener.parseFailed == true {
		return nil, failure("failed to parse query '%s'", s)
	}
	if len(*listener.stack) == 0 {
		return nil, failure("incomplete query '%s'", s)
	}
	result := listener.Pop()
	return result, nil
}

// Find items matching the query, return error if the query is ill-formed
// and the items otherwise.
func (db *MDB) Find(query *Query, limit int64) ([]Item, error) {
	result := make([]Item, 0)
	table := (*query).data
	if len((*query).Children) == 0 {
		return result, failure("incomplete query, only table given")
	}
	query = &query.Children[0]
	toExec, err := db.ToSql(table, query, limit)
	// fmt.Println(toExec) // the final query, for debugging
	if err != nil {
		return result, failure("invalid query - %s", err)
	}
	if !db.TableExists(table) {
		return result, failure("invalid query - table '%s' does not exist", table)
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
