package minidb

import (
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"
)

const MAXTEST = 100

func TestKVStore(t *testing.T) {
	var tmp *os.File
	tmp, _ = ioutil.TempFile("", "minidb-kvstore-testing-*")
	db, err := Open("sqlite3", tmp.Name())
	if err != nil {
		t.Errorf("Open() failed: %s", err)
	}
	// test integer storage
	var i int64
	for i = 1; i < MAXTEST; i++ {
		rvalue := rand.Int63()
		db.SetInt(i, rvalue)
		if db.GetInt(i) != rvalue {
			t.Errorf("failure setting int value, given %d, expected %d", db.GetInt(i), rvalue)
		}
		db.DeleteInt(i)
	}
	// test string storage
	for i = 1; i < MAXTEST; i++ {
		rvalue := String(40)
		db.SetStr(i, rvalue)
		if db.GetStr(i) != rvalue {
			t.Errorf(`failure setting string value, given "%s", expected "%s"`, db.GetStr(i), rvalue)
		}
		db.DeleteStr(i)
	}
	// test blob storage
	for i = 1; i < MAXTEST; i++ {
		rvalue := []byte(String(40))
		db.SetBlob(i, rvalue)
		if string(db.GetBlob(i)) != string(rvalue) {
			t.Errorf(`failure setting blob value, given "%s", expected "%s"`, string(db.GetBlob(i)), string(rvalue))
		}
		db.DeleteBlob(i)
	}
	// test date storage
	for i = 1; i < MAXTEST; i++ {
		rvalue, err := ParseTime(NewRandomDateStr())
		_ = err
		db.SetDate(i, rvalue)
		d1 := rvalue.UTC().Format(time.RFC3339)
		d2 := db.GetDate(i).UTC().Format(time.RFC3339)
		if d1 != d2 {
			t.Errorf(`failure setting date value, given "%s", expected "%s"`, d2, d1)
		}
		db.DeleteBlob(i)
	}
	db.Close()
	os.Remove(tmp.Name())
}
