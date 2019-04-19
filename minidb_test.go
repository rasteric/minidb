package minidb

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"
)

var db *MDB
var tmpfile *os.File
var tmpfile2 *os.File

func TestToBaseType(t *testing.T) {
	tables := []struct {
		in  FieldType
		out FieldType
	}{
		{DBIntList, DBInt},
		{DBStringList, DBString},
		{DBBlobList, DBBlob},
		{DBDateList, DBDate},
	}
	for _, table := range tables {
		result := ToBaseType(table.in)
		if result != table.out {
			t.Errorf("ToBaseType(%d) incorrect result, expected %d, given %d", table.in, table.out, result)
		}
	}
	if ToBaseType(DBError) != DBError {
		t.Errorf("ToBaseType(%d) incorrect result, expected %d, given %d", DBError, DBError, ToBaseType(DBError))
	}
}

func TestFail(t *testing.T) {
	var result error
	result = Fail("This is a test %d %s %s", 42, "Hello", "World")
	_ = result
}

func TestValue(t *testing.T) {
	// integers
	v := NewInt(2363867)
	if v.Int() != 2363867 {
		t.Errorf("NewInt(2363867) failed.")
	}
	v = NewInt(0)
	if v.Int() != 0 {
		t.Errorf("NewInt(0) failed.")
	}
	v = NewInt(-1)
	if v.Int() != -1 {
		t.Errorf("NewInt(-1) failed.")
	}
	v = NewInt(9223372036854775807)
	if v.Int() != 9223372036854775807 {
		t.Errorf("NewInt(9223372036854775807) failed.")
	}
	v = NewInt(-9223372036854775808)
	if v.Int() != -9223372036854775808 {
		t.Errorf("NewInt(-9223372036854775808) failed.")
	}
	if v.Sort != DBInt {
		t.Errorf("value of type int64 has wrong type, should be %d, given %d", DBInt, v.Sort)
	}
	b := v.Bytes()
	if b == nil {
		t.Errorf("Int64 Value cannot be converted to bytes.")
	}
	if !bytes.Equal(b, []byte("\x00\x00\x00\x00\x00\x00\x00\x80")) {
		t.Errorf("Int64 conversion to Bytes failed, given %q, expected %q", b, []byte("\x00\x00\x00\x00\x00\x00\x00\x80"))
	}
	if v.String() != "-9223372036854775808" {
		t.Errorf("Bytes to String conversion failed in Value.String(), expected '%q', given '%q'", "-9223372036854775808", v.String())
	}

	// strings
	v = NewString("")
	if v.String() != "" {
		t.Errorf("NewString failed.")
	}
	v = NewString("Hello world")
	if v.String() != "Hello world" {
		t.Errorf("NewString failed.")
	}
	v = NewString("Testing «ταБЬℓσ»: 1<2 & 4+1>3, now 20% off!")
	if v.String() != "Testing «ταБЬℓσ»: 1<2 & 4+1>3, now 20% off!" {
		t.Errorf("NewString unicode test failed.")
	}
	if !bytes.Equal(v.Bytes(), []byte(v.String())) {
		t.Errorf("Cannot convert Value of type String to Bytes, given %q, expected %q", []byte(v.String()), v.Bytes())
	}
	if v.Sort != DBString {
		t.Errorf("value of string type has wrong type, should be %d, given %d", DBString, v.Sort)
	}

	// bytes
	v = NewBytes([]byte("Hello world"))
	expect := base64.StdEncoding.EncodeToString(v.Bytes())
	if v.String() != expect {
		t.Errorf("NewBytes failed, given '%s', expected '%s'", v.String(), expect)
	}
	if !bytes.Equal(v.Bytes(), []byte("Hello world")) {
		t.Errorf("NewBytes failed, not the same bytes array is stored than what was passed as an argument")
	}
	v = NewBytes([]byte(""))
	if v.Bytes() == nil {
		t.Errorf("NewBytes failed, empty bytes should be != nil")
	}

	// dates
	d := time.Now()
	v = NewDate(d)
	if d.UTC().Format(time.RFC3339) != v.String() {
		t.Errorf("NewDate failed, expected %s, given %s", d.UTC().Format(time.RFC3339), v.String())
	}
	v = NewDateStr("2018-12-24T18:00:00Z")
	if v.String() != "2018-12-24T18:00:00Z" {
		t.Errorf("NewDateStr failed, expected %s, given %s", "2018-12-24T18:00:00Z", v.String())
	}
	v = NewDateStr("2002-10-02T10:00:00-05:00")
	if v.String() != "2002-10-02T10:00:00-05:00" {
		t.Errorf("NewDateStr failed, expected %s, given %s", "2002-10-02T10:00:00-05:00", v.String())
	}
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("NewDateStr failed, invalid input should have raised panic but passed.")
		}
	}()
	v = NewDateStr("2002-10-02 10:00:00Z")
	v = NewDateStr("2002-10-02T10:00:00")
}

func TestGetUserTypeString(t *testing.T) {
	tables := []struct {
		in  FieldType
		out string
	}{
		{DBIntList, "int-list"},
		{DBStringList, "string-list"},
		{DBBlobList, "blob-list"},
		{DBDateList, "date-list"},
		{DBInt, "int"},
		{DBString, "string"},
		{DBBlob, "blob"},
		{DBDate, "date"},
	}
	for _, table := range tables {
		result := GetUserTypeString(table.in)
		if result != table.out {
			t.Errorf("GetUserTypeString(%d) incorrect result, expected %s, given %s", table.in, table.out, result)
		}
	}
	if GetUserTypeString(DBError) != "unknown" {
		t.Errorf("GetUserTypeString(%d) incorrect result, expected 'unknown', given '%s'", DBError, GetUserTypeString(DBError))
	}
}

func TestParseFieldDesc(t *testing.T) {
	tables := []struct {
		in  []string
		out []Field
	}{
		{[]string{"int", "Age"}, []Field{Field{"Age", DBInt}}},
		{[]string{"string", "Name", "string-list", "Address"},
			[]Field{Field{"Name", DBString}, Field{"Address", DBStringList}}},
		{[]string{"date-list", "Meetings"}, []Field{Field{"Meetings", DBDateList}}},
		{[]string{"blob", "foo"}, []Field{Field{"foo", DBBlob}}},
	}
	for _, table := range tables {
		result, err := ParseFieldDesc(table.in)
		if err != nil {
			t.Errorf("ParseFieldDesc() failed on table test")
		}
		for i := range table.out {
			if table.out[i] != result[i] {
				t.Errorf("ParseFieldDesc(%s) incorrect result", table.in)
			}
		}
	}
}

func TestMDB(t *testing.T) {
	db, err := Open("sqlite3", tmpfile.Name())
	if err != nil {
		t.Errorf("Open() failed: %s", err)
	}
	// test code here
	err = db.AddTable("test", []Field{Field{"Name", DBStringList},
		Field{"Email", DBString},
		Field{"Age", DBInt},
		Field{"Scores", DBIntList},
		Field{"Modified", DBDate},
		Field{"Misc", DBBlob},
		Field{"Data", DBBlobList},
		Field{"Schedules", DBDateList},
	})
	if err != nil {
		t.Errorf("MDB.AddTable() failed: %s", err)
	}
	if !db.TableExists("test") {
		t.Errorf("MDB.AddTable() failed or MDB.TableExists() failed, should be true, given false")
	}
	if db.TableExists("humpty") {
		t.Errorf("MDB.TableExists() returned true instead of false for nonexistent table")
	}
	if db.ItemExists("test", 1) {
		t.Errorf("MDB.ItemExists() returned true instead of false for table 'test'")
	}
	item, err := db.NewItem("test")
	if err != nil {
		t.Errorf("MDB.NewItem() failed for table 'test'")
	}
	if !db.ItemExists("test", item) {
		t.Errorf("MDB.ItemExists() returned false for existing item %d, expected true", item)
	}
	for _, field := range []string{"Name", "Email", "Age", "Scores", "Modified", "Misc", "Data", "Schedules"} {
		if !db.FieldExists("test", field) {
			t.Errorf("MDB.FieldExists(%s) is false, but should be true", field)
		}
		var c FieldType
		k := db.MustGetFieldType("test", field)
		switch field {
		case "Name":
			c = DBStringList
		case "Email":
			c = DBString
		case "Age":
			c = DBInt
		case "Scores":
			c = DBIntList
		case "Modified":
			c = DBDate
		case "Misc":
			c = DBBlob
		case "Data":
			c = DBBlobList
		case "Schedules":
			c = DBDateList
		}
		if k != c {
			t.Errorf("MDB.MustGetFieldType() unexpected value, expected %d, given %d", c, k)
		}
		if !db.FieldIsNull("test", item, field) {
			t.Errorf("MDB.FieldIsNull() returned true instead of false for NULL field '%s'", field)
		}
		if !db.FieldIsEmpty("test", item, field) {
			t.Errorf("MDB.FieldIsEmpty() returned true instead of false for NULL field '%s'", field)
		}
	}
	if db.ItemExists("test", 99) {
		t.Errorf("MDB.ItemExists() returns true for fictitious item 99 (did the driver just assign this as first id?)")
	}
	if db.FieldExists("test", "blurbfoo") {
		t.Errorf("MDB.FieldExists() returns true but should return false for nonexistent field")
	}
	if db.FieldExists("schmoo", "Name") {
		t.Errorf("MDB.FieldExists() returns true but should return false for fictitious field in nonexistent table")
	}
	if db.IsListField("test", "Email") {
		t.Errorf("MDB.IsListField() returns true for non-list field")
	}
	if !db.IsListField("test", "Name") {
		t.Errorf("MDB.IsListField() returns false for list field, should be true")
	}
	if db.IsListField("test", "schmoo") {
		t.Errorf("MDB.IsListField() returns true for nonexistent field, should be false")
	}
	if db.IsListField("schmmoo", "what") {
		t.Errorf("MDB.IsListField() returns true for nonexistent field in nonexistent table, should be false")
	}
	if !db.IsEmptyListField("test", 1, "Name") {
		t.Errorf("MDB.IsEmptyListField() returns false for empty list field")
	}
	if db.IsEmptyListField("test", 1, "Email") {
		t.Errorf("MDB.IsEmptyListField() returns true for non list field")
	}
	if !db.FieldExists("test", "Email") {
		t.Errorf("MDB.FieldExists() returns false for an existing field")
	}
	if !db.FieldExists("test", "Name") {
		t.Errorf("MDB.FieldExists() returns false for an existing list field")
	}
	if db.FieldExists("test", "schmoo") {
		t.Errorf("MDB.FieldExists() returns true for a nonexistent field")
	}
	if db.FieldExists("schmoo", "Name") {
		t.Errorf("MDB.FieldExists() returns true for a field in a nonexistent table")
	}
	_, err = db.ParseFieldValues("schmoo", "test", []string{})
	if err == nil {
		t.Errorf("MDB.ParseFieldValues() returns no error for nonexistent table")
	}
	_, err = db.ParseFieldValues("test", "Name", []string{"John", "Theodore", "Smith"})
	if err != nil {
		t.Errorf("MDB.ParseFieldValues() returns error for correct input")
	}
	_, err = db.ParseFieldValues("test", "Name", []string{""})
	if err != nil {
		t.Errorf("MDB.ParseFieldValues() returns error for correct input")
	}
	_, err = db.ParseFieldValues("test", "Name", []string{})
	if err == nil {
		t.Errorf("MDB.ParseFieldValues() returns no error for empty input, should indicate an error")
	}
	_, err = db.ParseFieldValues("test", "Age", []string{"-1234"})
	if err != nil {
		t.Errorf("MDB.ParseFieldValues() returns error for correct numeric input")
	}
	_, err = db.ParseFieldValues("test", "Email", []string{"john", "hello"})
	if err == nil {
		t.Errorf("MDB.ParseFieldValues() returns no error for too long input, should indicate an error")
	}
	_, err = db.ParseFieldValues("test", "Age", []string{"0dude"})
	if err == nil {
		t.Errorf("MDB.ParseFieldValues() returns no error for incorrect numeric input, should indicate an error")
	}
	_, err = db.ParseFieldValues("test", "Scores", []string{"27", "23.3", "7"})
	if err == nil {
		t.Errorf("MDB.ParseFieldValues() returns no error for incorrect numeric input")
	}
	_, err = db.ParseFieldValues("test", "Scores", []string{"27", "23", "7"})
	if err != nil {
		t.Errorf("MDB.ParseFieldValues() returns an error for correct numeric input")
	}
	_, err = db.ParseFieldValues("test", "Misc", []string{"SGVsbG8gd29ybGQh"})
	if err != nil {
		t.Errorf("MDB.ParseFieldValues() returns an error for correct blob input")
	}
	_, err = db.ParseFieldValues("test", "Misc", []string{"SGVXsbG8gd29ybGQh"})
	if err == nil {
		t.Errorf("MDB.ParseFieldValues() returns no error for incorrect blob input")
	}
	_, err = db.ParseFieldValues("test", "Modified", []string{"1999-12-31T01:00:00Z"})
	if err != nil {
		t.Errorf("MDB.ParseFieldValues() returns error for correct date input")
	}
	_, err = db.ParseFieldValues("test", "Modified", []string{"1999-12-31T01:00:00+05:00"})
	if err != nil {
		t.Errorf("MDB.ParseFieldValues() returns error for correct date input")
	}
	_, err = db.ParseFieldValues("test", "Modified", []string{"1999-12-31T01:00:00"})
	if err == nil {
		t.Errorf("MDB.ParseFieldValues() returns no error for incorrect date input")
	}
	_, err = db.ParseFieldValues("test", "Modified", []string{"1999-12-31T01:00:00+"})
	if err == nil {
		t.Errorf("MDB.ParseFieldValues() returns no error for incorrect date input")
	}
	_, err = db.ParseFieldValues("test", "Modified", []string{"1999-12-31 01:00:00Z"})
	if err == nil {
		t.Errorf("MDB.ParseFieldValues() returns no error for incorrect date input")
	}
	if count, _ := db.Count("test"); count > 1 {
		t.Errorf("MDB.Count() > 0 for empty table")
	}
	if count, _ := db.Count("schmoo"); count != 0 {
		t.Errorf("MDB.Count() > 0 for nonextistent table")
	}
	items, _ := db.ListItems("test", 1024)
	if items[0] != item {
		t.Errorf("MDB.ListItems(), expected %d, given %d", item, items[0])
	}
	items, _ = db.ListItems("schmoo", 1024)
	if err == nil {
		t.Errorf("MDB.ListItems() should return error for nonextistent table, given no error")
	}
	err = db.Set("test", item, "Name", []Value{NewString("John"), NewString("Theodore"),
		NewString("Smith")})
	if err != nil {
		t.Errorf("MDB.Set() returns error for correct data")
	}
	err = db.Set("test", item, "Name", []Value{NewString("John"), NewInt(333),
		NewString("Smith")})
	if err == nil {
		t.Errorf("MDB.Set() returns no error for incorrect data")
	}
	values, err := db.Get("test", item, "Name")
	if err != nil {
		t.Errorf("MDB.Get() returns error for valid request: %s", err)
	}
	if len(values) != 3 {
		t.Errorf("MDB.Get() returns wrong slice length for string list of length 3: %d", len(values))
	}
	if values[0].String() != "John" || values[1].String() != "Theodore" || values[2].String() != "Smith" {
		t.Errorf("MDB.Get() returns garbage instead of previously set string list data")
	}
	// setting and getting different types of data
	err = db.Set("test", item, "Age", []Value{NewInt(30)})
	if err != nil {
		t.Errorf("MDB.Set() failed: %s", err)
	}
	values, err = db.Get("test", item, "Age")
	if err != nil {
		t.Errorf("MDB.Get() failed: %s", err)
	}
	if len(values) != 1 {
		t.Errorf("MDG.Get() for single datum returned []Value of length %d", len(values))
	}
	if values[0].Int() != 30 {
		t.Errorf("MDB.Get() failed, given %d, expected %d", values[0].Int(), 30)
	}
	err = db.Set("test", item, "Email", []Value{NewString("Hello world")})
	if err != nil {
		t.Errorf("MDB.Set() failed: %s", err)
	}
	values, err = db.Get("test", item, "Email")
	if len(values) == 0 || values[0].String() != "Hello world" {
		t.Errorf("MDB.Get() failed: %s", err)
	}
	err = db.Set("test", item, "Scores", []Value{NewInt(10), NewInt(20), NewInt(30)})
	if err != nil {
		t.Errorf("MDB.Set() failed: %s", err)
	}
	values, err = db.Get("test", item, "Scores")
	if err != nil {
		t.Errorf("MDB.Get() failed: %s", err)
	}
	if len(values) != 3 {
		t.Errorf("MDB.Get() failed, expected int list of length 3, given length %d", len(values))
	}
	if values[0].Int() != 10 || values[1].Int() != 20 || values[2].Int() != 30 {
		t.Errorf("MDB.Get() failed, returning garbage instead of the expected int slice 10, 20, 30")
	}
	d := time.Now()
	err = db.Set("test", item, "Modified", []Value{NewDate(d)})
	if err != nil {
		t.Errorf("MDB.Set() single date failed: %s", err)
	}
	values, err = db.Get("test", item, "Modified")
	if err != nil {
		t.Errorf("MDB.Get() single date failed: %s", err)
	}
	if len(values) != 1 || d.UTC().Format(time.RFC3339) != values[0].Datetime().UTC().Format(time.RFC3339) {
		t.Errorf("MDB.Get() or previous MDB.Set() failed for date, expected RFC3339 date %s, given %s",
			d.UTC().Format(time.RFC3339), values[0].Datetime().Format(time.RFC3339))
	}
	b := []byte("This is \000a test")
	err = db.Set("test", item, "Misc", []Value{NewBytes(b)})
	if err != nil {
		t.Errorf("MDB.Set() failed for bytes with null in it: %s", err)
	}
	values, err = db.Get("test", item, "Misc")
	if err != nil {
		t.Errorf("MDB.Get() single blob containing null failed: %s", err)
	}
	if len(values) != 1 {
		t.Errorf("MDB.Get() single blob failed, length should be 1, given %d", len(values))
	}
	for i, b2 := range values[0].Bytes() {
		if b[i] != b2 {
			t.Errorf("MDB.Get() single blob or previous MDB.Set() failed, expected blob[%d]=%q, given blob[%d]=%q", i, b2, i, b[i])
			break
		}
	}
	err = db.Set("test", item, "Data", []Value{NewBytes([]byte(""))})
	if err != nil {
		t.Errorf("MDB.Set() failed for empty values")
	}
	values, err = db.Get("test", item, "Data")
	if err != nil {
		t.Errorf("MDB.Get() blob list failed: %s", err)
	}
	if len(values) != 1 {
		t.Errorf("MDB.Get() blob list failed, expected return length 1, given %d", len(values))
	}
	if len(values[0].Bytes()) != 0 {
		t.Errorf("MDB.Get() blob list with one empty blob, expected an empty []byte slice, given something else")
	}
	err = db.Set("test", item, "Schedules", []Value{NewDateStr("2017-02-27T17:31:00Z"),
		NewDateStr("1969-04-30T23:59:00+04:00"), NewDateStr("2140-12-23T18:00:00Z")})
	if err != nil {
		t.Errorf("MDB.Set() failed for list of dates: %s", err)
	}
	values, err = db.Get("test", item, "Schedules")
	if err != nil {
		t.Errorf("MDB.Get() failed for date list: %s", err)
	}
	if len(values) != 3 {
		t.Errorf("MDB.Get() expected date list of length 3, given length %d", len(values))
	}
	v := NewDateStr("2017-02-27T17:31:00Z")
	expected := v.Datetime().UTC().Format(time.RFC3339)
	given := values[0].String()
	if given != expected {
		t.Errorf("MDB.Get() failed for date list entry 0, given %s, expected %s", given, expected)
	}
	v = NewDateStr("1969-04-30T23:59:00+04:00")
	expected = v.Datetime().UTC().Format(time.RFC3339)
	given = values[1].String()
	if given != expected {
		t.Errorf("MDB.Get() failed for date list entry 0, given %s, expected %s", given, expected)
	}
	v = NewDateStr("2140-12-23T18:00:00Z")
	expected = v.Datetime().UTC().Format(time.RFC3339)
	given = values[2].String()
	if given != expected {
		t.Errorf("MDB.Get() failed for date list entry 0, given %s, expected %s", given, expected)
	}
	// GetFields
	fields, err := db.GetFields("test")
	if err != nil {
		t.Errorf("MDB.GetFields() failed: %s", err)
	}
	if len(fields) != 8 {
		t.Errorf("MDB.GetFields() failed, expected %d fields, returned %d", 8, len(fields))
	}
	e := make([]string, 8)
	e[0] = "Name"
	e[1] = "Email"
	e[2] = "Age"
	e[3] = "Scores"
	e[4] = "Modified"
	e[5] = "Misc"
	e[6] = "Data"
	e[7] = "Schedules"
	for i := range fields {
		found := false
		for j := range e {
			if fields[i].Name == e[j] {
				found = true
			}
		}
		if !found {
			t.Errorf("MDB.GetFields() returned an unexpected result")
		}
	}
	q := make([]FieldType, 8)
	q[0] = DBStringList
	q[1] = DBString
	q[2] = DBInt
	q[3] = DBIntList
	q[4] = DBDate
	q[5] = DBBlob
	q[6] = DBBlobList
	q[7] = DBDateList
	for i := range fields {
		found := false
		for j := range q {
			if fields[i].Sort == q[j] {
				found = true
			}
		}
		if !found {
			t.Errorf("MDB.GetFields() returned an unexpected field type")
		}
	}
	// GetTables
	tables := db.GetTables()
	if len(tables) != 1 || tables[0] != "test" {
		t.Errorf("MDB.GetTables() returned garbage")
	}

	// RemoveItem
	if !db.ItemExists("test", item) {
		t.Errorf("MDB.ItemExists() returned false for existing item")
	}
	err = db.RemoveItem("test", item)
	if err != nil {
		t.Errorf("MDB.RemoveItem() failed: %s", err)
	}
	if db.ItemExists("test", item) {
		t.Errorf("MDB.ItemExists() returned true for nonexistent item, should have returned false")
	}

	// close it
	err = db.Close()
	if err != nil {
		t.Errorf("Close() failed: %s", err)
	}
}

func TestParseQuery(t *testing.T) {
	queries := []string{`test Name=r% or Name=John`,
		"Person Name=Smith or not Name=John",
		"Person Age=47 and Name=John%",
		`Person Name="John"`,
		`Person Email="john@smith.com" and Name="%r%"`,
		`Person every Name=%e%`,
		`Person no Name=John`,
		`Person not every Name=John`,
	}
	for _, query := range queries {
		_, err := ParseQuery(query)
		if err != nil {
			t.Errorf(`ParseQuery("%s") should work but returns error: %s`, query, err)
		}
	}
}

// we create random strings for the MDB.Find test, setting this up first
const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func String(length int) string {
	return StringWithCharset(length, charset)
}

type testin struct {
	item   Item
	Names  []string
	Email  string
	Date   string
	Dates  []string
	Blob   []byte
	Blobs  []string
	Age    int64
	Scores []int64
}

func NewRandomDateStr() string {
	s := fmt.Sprintf("%4d-%02d-%02dT%02d:%02d:%02d+%d:00",
		1800+rand.Intn(400),
		1+rand.Intn(12),
		1+rand.Intn(27),
		1+rand.Intn(24),
		rand.Intn(60),
		rand.Intn(60),
		rand.Intn(24),
	)
	return s
}

const UintSize = 32 << (^uint(0) >> 32 & 1) // 32 or 64

const (
	MaxInt  = 1<<(UintSize-1) - 1 // 1<<31 - 1 or 1<<63 - 1
	MinInt  = -MaxInt - 1         // -1 << 31 or -1 << 63
	MaxUint = 1<<UintSize - 1     // 1<<32 - 1 or 1<<64 - 1
)

func NewRandomTestin(n Item) testin {
	var ti testin
	ti.item = n
	ns := make([]string, 3)
	for i := range ns {
		ns[i] = String(rand.Intn(20) + 1)
	}
	ti.Names = ns
	ti.Email = String(rand.Intn(20) + 3)
	ti.Date = NewRandomDateStr()
	dates := make([]string, 3)
	for i := range dates {
		dates[i] = NewRandomDateStr()
	}
	ti.Dates = dates
	ti.Blob = []byte(String(rand.Intn(20) + 3))
	ti.Blobs = make([]string, 3)
	for i := range ti.Blobs {
		ti.Blobs[i] = String(rand.Intn(20) + 1)
	}
	ti.Age = int64(rand.Intn(100))
	ti.Scores = make([]int64, 3)
	for i := range ti.Scores {
		ti.Scores[i] = int64(rand.Intn(MaxInt))
	}
	return ti
}

func TestFind(t *testing.T) {
	db, err := Open("sqlite3", tmpfile.Name())
	if err != nil {
		t.Errorf("Open() failed: %s", err)
	}
	// we first create a number of test inputs
	var in []testin
	const maxtest = 30
	in = make([]testin, maxtest)
	for i := range in {
		item, err := db.NewItem("test")
		if err != nil {
			t.Errorf("MDB.NewItem() failed: %s", err)
		}
		in[i] = NewRandomTestin(item)
		err = db.Set("test", in[i].item, "Name", []Value{NewString(in[i].Names[0]),
			NewString(in[i].Names[1]), NewString(in[i].Names[2])})
		if err != nil {
			t.Errorf("MDB.Set() failed: %s", err)
		}
		err = db.Set("test", in[i].item, "Email", []Value{NewString(in[i].Email)})
		if err != nil {
			t.Errorf("MDB.Set() failed: %s", err)
		}
		err = db.Set("test", in[i].item, "Modified", []Value{NewDateStr(in[i].Date)})
		if err != nil {
			t.Errorf("MDB.Set() failed: %s", err)
		}
		err = db.Set("test", in[i].item, "Schedules", []Value{NewDateStr(in[i].Dates[0]),
			NewDateStr(in[i].Dates[1]), NewDateStr(in[i].Dates[2])})
		if err != nil {
			t.Errorf("MDB.Set() failed: %s", err)
		}
		err = db.Set("test", in[i].item, "Misc", []Value{NewBytes([]byte(in[i].Blob))})
		if err != nil {
			t.Errorf("MDB.Set() failed: %s", err)
		}
		err = db.Set("test", in[i].item, "Data", []Value{NewBytes([]byte(in[i].Blobs[0])),
			NewBytes([]byte(in[i].Blobs[1])), NewBytes([]byte(in[i].Blobs[2]))})
		if err != nil {
			t.Errorf("MDB.Set() failed: %s", err)
		}
		err = db.Set("test", in[i].item, "Age", []Value{NewInt(in[i].Age)})
		if err != nil {
			t.Errorf("MDB.Set() failed: %s", err)
		}
		err = db.Set("test", in[i].item, "Scores", []Value{NewInt(in[i].Scores[0]),
			NewInt(in[i].Scores[1]), NewInt(in[i].Scores[2])})
		if err != nil {
			t.Errorf("MDB.Set() failed: %s", err)
		}
	}
	// now query everything directly first
	for _, v := range in {
		// Names string list field
		for j := range v.Names {
			query := `test Name=` + v.Names[j]
			q, err := ParseQuery(query)
			if err != nil {
				t.Errorf(`ParseQuery("%s") failed: %s`, query, err)
			}
			results, err := db.Find(q, 200)
			if err != nil {
				t.Errorf(`MDB.Find() failed for query "%s", should have succeeded`, query)
			}
			found := false
			for k := range results {
				if results[k] == v.item {
					found = true
				}
			}
			if !found {
				t.Errorf(`MDB.Find() failed to find Name for query "%s", should have succeeded`, query)
			}
		}
		// Age int field
		query := fmt.Sprintf("test Age=%d", v.Age)
		q, err := ParseQuery(query)
		if err != nil {
			t.Errorf(`ParseQuery("%s") failed: %s`, query, err)
		}
		results, err := db.Find(q, 200)
		if err != nil {
			t.Errorf(`MDB.Find() failed for query "%s", should have succeeded`, query)
		}
		found := false
		for k := range results {
			if results[k] == v.item {
				found = true
			}
		}
		if !found {
			t.Errorf(`MDB.Find() failed to find Name for query "%s", should have succeeded`, query)
		}
		// Email string field
		query = fmt.Sprintf("test Email=%s", v.Email)
		q, err = ParseQuery(query)
		if err != nil {
			t.Errorf(`ParseQuery("%s") failed: %s`, query, err)
		}
		results, err = db.Find(q, 200)
		if err != nil {
			t.Errorf(`MDB.Find() failed for query "%s", should have succeeded`, query)
		}
		found = false
		for k := range results {
			if results[k] == v.item {
				found = true
			}
		}
		if !found {
			t.Errorf(`MDB.Find() failed to find Name for query "%s", should have succeeded`, query)
		}
		// Scores int list field
		// Names string list field
		for j := range v.Scores {
			query := fmt.Sprintf(`test Scores=%d`, v.Scores[j])
			q, err := ParseQuery(query)
			if err != nil {
				t.Errorf(`ParseQuery("%s") failed: %s`, query, err)
			}
			results, err := db.Find(q, 200)
			if err != nil {
				t.Errorf(`MDB.Find() failed for query "%s", should have succeeded`, query)
			}
			found := false
			for k := range results {
				if results[k] == v.item {
					found = true
				}
			}
			if !found {
				t.Errorf(`MDB.Find() failed to find Name for query "%s", should have succeeded`, query)
			}
		}
		// Data blob list field
		for j := range v.Blobs {
			query := fmt.Sprintf("test Data=%s", v.Blobs[j])
			q, err := ParseQuery(query)
			if err != nil {
				t.Errorf(`ParseQuery("%s") failed: %s`, query, err)
			}
			results, err := db.Find(q, 200)
			if err != nil {
				t.Errorf(`MDB.Find() failed for query "%s", should have succeeded: %s`, query, err)
			}
			found := false
			for k := range results {
				if results[k] == v.item {
					found = true
				}
			}
			if !found {
				t.Errorf(`MDB.Find() failed to find Blob for query "%s", should have succeeded`, query)
			}
		}
		// Misc blob field
		query = fmt.Sprintf("test Misc=%s", v.Blob)
		q, err = ParseQuery(query)
		if err != nil {
			t.Errorf(`ParseQuery("%s") failed: %s`, query, err)
		}
		results, err = db.Find(q, 200)
		if err != nil {
			t.Errorf(`MDB.Find() failed for query "%s", should have succeeded`, query)
		}
		found = false
		for k := range results {
			if results[k] == v.item {
				found = true
			}
		}
		if !found {
			t.Errorf(`MDB.Find() failed to find single Blob for query "%s", should have succeeded`, query)
		}
	}
	db.Close()
}

func TestBackup(t *testing.T) {
	db, err := Open("sqlite3", tmpfile.Name())
	if err != nil {
		t.Errorf("Open() failed: %s", err)
	}
	defer db.Close()
	if err := db.Backup(tmpfile2.Name()); err != nil {
		t.Errorf(`MDB.Backup() failed: %s`, err)
	}
}

func setup() {
	tmpfile, _ = ioutil.TempFile("", "minidb-testing-*")
	tmpfile2, _ = ioutil.TempFile("", "minidb-testing-*")
}

func teardown() {
	if db != nil {
		db.Close()
	}
	if tmpfile != nil {
		os.Remove(tmpfile.Name())
	}
	if tmpfile2 != nil {
		os.Remove(tmpfile.Name())
	}
}

func TestMain(m *testing.M) {
	setup()
	retCode := m.Run()
	teardown()
	os.Exit(retCode)
}
