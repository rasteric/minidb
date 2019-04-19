package minidb

import (
	"io/ioutil"
	"testing"
)

func TestMultiDB(t *testing.T) {
	var db *MultiDB
	tmpdir, err := ioutil.TempDir("", "multidb")
	if err != nil {
		t.Errorf(`could not create temporary directory for testing`)
	}
	db, err = NewMultiDB(tmpdir, "sqlite3")
	if err != nil {
		t.Errorf(`error creating MultiDB: %s`, err)
	}
	// some test users
	p := DefaultParams()
	salt := GenerateExternalSalt(p)
	key := GenerateKey("a test password", salt, p)
	user1, _, err := db.NewUser("John", "john@test.com", key)
	if err != nil {
		t.Errorf(`could not create new user "John", %s`, err)
	}
	salt2 := GenerateExternalSalt(p)
	key2 := GenerateKey("another password", salt2, p)
	user2, _, err := db.NewUser("Bob", "bob@testing.com", key2)
	if err != nil {
		t.Errorf(`could not create new user "Bob", %s`, err)
	}
	salt3 := GenerateExternalSalt(p)
	key3 := GenerateKey("some password", salt3, p)
	_, errcode, _ := db.NewUser("John", "joey@test.com", key3)
	if errcode != ErrUsernameInUse {
		t.Errorf(`expected errcode=%d for NewUser on existing username, given %d`, ErrUsernameInUse, errcode)
	}
	// The following line is bad practice, you should never re-use salts and passwords.
	_, errcode, _ = db.NewUser("Johnny", "john@test.com", key3)
	if errcode != ErrEmailInUse {
		t.Errorf(`expected errcode=%d for NewUser on existing email, given %d`, ErrEmailInUse, errcode)
	}
	// various methods test
	if !db.ExistingUser(user1.Name()) {
		t.Errorf(`MultiDB.ExistingUser() returned false for existing user`)
	}
	if !db.ExistingUser(user2.Name()) {
		t.Errorf(`MultiDB.ExistingUser() returned false for existing user`)
	}
	// authentication tests
	salt4, errcode, err := db.ExternalSalt("John")
	if err != nil {
		t.Errorf(`MultiDB.ExternalSalt() failed, errcode=%d: %s`, errcode, err)
	}
	key4 := GenerateKey("a test password", salt4, p)
	_, errcode, err = db.Authenticate("John", key4)
	if err != nil {
		t.Errorf(`MultiDB.Authenticate() failed with errcode=%d: %s`, errcode, err)
	}
	// do something with the DB
	user1db, reply, err := db.UserDB(user1)
	if err != nil {
		t.Errorf(`MultiDB.UserDB() failed with errcode=%d: %s`, reply, err)
	}
	user2db, reply, err := db.UserDB(user2)
	if err != nil {
		t.Errorf(`MultiDB.UserDB() failed with errcode=%d: %s`, reply, err)
	}
	user1db.SetStr(1, "test")
	if s := user1db.GetStr(1); s != "test" {
		t.Errorf(`simple get/set failed: %s`, err)
	}
	user2db.SetStr(1, "hello world")
	if s := user2db.GetStr(1); s != "hello world" {
		t.Errorf(`simple get/set failed: %s`, err)
	}
	if s, reply, err := db.UserEmail(user1); s != "john@test.com" {
		t.Errorf(`test user email not stored correctly, errcode=%d, expected "john@test.com", given "%s": %s`, reply, s, err)
	}
	if s, reply, err := db.UserEmail(user2); s != "bob@testing.com" {
		t.Errorf(`test user email not stored correctly, errcode=%d, expected "bob@testing.com", given "%s": %s`, reply, s, err)
	}
	// delete the DB
	if reply, err := db.Delete(); err != nil {
		t.Errorf(`error %d deleting the MultiDB: %s`, reply, err)
	}
}
