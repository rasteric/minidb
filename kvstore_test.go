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
	var tx *Tx
	tmp, _ = ioutil.TempFile("", "minidb-kvstore-testing-*")
	db, err := Open("sqlite3", tmp.Name())
	if err != nil {
		t.Errorf("Open() failed: %s", err)
	}
	// test integer storage
	var i int64
	for i = 1; i < MAXTEST; i++ {
		rvalue := rand.Int63()
		tx, err = db.Begin()
		if err != nil {
			t.Errorf("db.Begin transaction failed: %s", err)
		}
		tx.SetInt(i, rvalue)
		err = tx.Commit()
		if err != nil {
			t.Errorf("tx.Commit failed: %s", err)
		}
		if db.GetInt(i) != rvalue {
			t.Errorf("failure setting int value, given %d, expected %d", db.GetInt(i), rvalue)
		}
		tx, err = db.Begin()
		if err != nil {
			t.Errorf("db.Begin transaction failed: %s", err)
		}
		tx.DeleteInt(i)
		err = tx.Commit()
		if err != nil {
			t.Errorf("tx.Commit failed: %s", err)
		}
	}
	// test string storage
	for i = 1; i < MAXTEST; i++ {
		rvalue := String(40)
		tx, err = db.Begin()
		if err != nil {
			t.Errorf("db.Begin transaction failed: %s", err)
		}
		tx.SetStr(i, rvalue)
		err = tx.Commit()
		if err != nil {
			t.Errorf("tx.Commit failed: %s", err)
		}
		if db.GetStr(i) != rvalue {
			t.Errorf(`failure setting string value, given "%s", expected "%s"`, db.GetStr(i), rvalue)
		}
		tx, err = db.Begin()
		if err != nil {
			t.Errorf("db.Begin transaction failed: %s", err)
		}
		tx.DeleteStr(i)
		err = tx.Commit()
		if err != nil {
			t.Errorf("tx.Commit failed: %s", err)
		}
	}
	// test blob storage
	for i = 1; i < MAXTEST; i++ {
		rvalue := []byte(String(40))
		tx, err = db.Begin()
		if err != nil {
			t.Errorf("db.Begin transaction failed: %s", err)
		}
		tx.SetBlob(i, rvalue)
		err = tx.Commit()
		if err != nil {
			t.Errorf("tx.Commit failed: %s", err)
		}
		if string(db.GetBlob(i)) != string(rvalue) {
			t.Errorf(`failure setting blob value, given "%s", expected "%s"`, string(db.GetBlob(i)), string(rvalue))
		}
		tx, err = db.Begin()
		if err != nil {
			t.Errorf("db.Begin transaction failed: %s", err)
		}
		tx.DeleteBlob(i)
		err = tx.Commit()
		if err != nil {
			t.Errorf("tx.Commit failed: %s", err)
		}
	}
	// test date storage
	for i = 1; i < MAXTEST; i++ {
		rvalue, err := ParseTime(NewRandomDateStr())
		_ = err
		tx, err = db.Begin()
		if err != nil {
			t.Errorf("db.Begin transaction failed: %s", err)
		}
		tx.SetDate(i, rvalue)
		d1 := rvalue.UTC().Format(time.RFC3339)
		err = tx.Commit()
		if err != nil {
			t.Errorf("tx.Commit failed: %s", err)
		}
		d2 := db.GetDate(i).UTC().Format(time.RFC3339)
		if d1 != d2 {
			t.Errorf(`failure setting date value, given "%s", expected "%s"`, d2, d1)
		}
		tx, err = db.Begin()
		if err != nil {
			t.Errorf("db.Begin transaction failed: %s", err)
		}
		tx.DeleteBlob(i)
		err = tx.Commit()
		if err != nil {
			t.Errorf("tx.Commit failed: %s", err)
		}
	}
	db.Close()
	os.Remove(tmp.Name())
}
