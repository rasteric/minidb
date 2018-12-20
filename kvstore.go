package minidb

import (
	"database/sql"
	"time"
)

// ------------------------------------------------------------------------------
// Key-Value Store
// ------------------------------------------------------------------------------

// GetInt returns the int64 value for a key, 0 if key doesn't exist.
func (db *MDB) GetInt(key int64) int64 {
	row := db.base.QueryRow(`SELECT Value FROM _KVINT WHERE Id=?`, key)
	var intResult sql.NullInt64
	err := row.Scan(&intResult)
	if err != nil || !intResult.Valid {
		return 0
	}
	return intResult.Int64
}

func (db *MDB) fetchStr(key int64, store string) string {
	row := db.base.QueryRow(`SELECT Value FROM `+store+` WHERE Id=?`, key)
	var strResult sql.NullString
	err := row.Scan(&strResult)
	if err != nil || !strResult.Valid {
		return ""
	}
	return strResult.String
}

// GetStr returns the string value for a key, "" if key doesn't exist.
func (db *MDB) GetStr(key int64) string {
	return db.fetchStr(key, "_KVSTR")
}

// GetBlob returns the blob value for a key, nil if key doesn't exist.
func (db *MDB) GetBlob(key int64) []byte {
	return []byte(db.fetchStr(key, "_KVBLOB"))
}

// GetDate returns the time.Time value of a, January 1, year 1, 00:00:00.000000000 UTC
// if the key doesn't exist.
func (db *MDB) GetDate(key int64) time.Time {
	s := db.fetchStr(key, "_KVDATE")
	t, err := ParseTime(s)
	if err != nil {
		return time.Time{}
	}
	return t
}

// GetDateStr returns a date in RFC3339 string form, the empty string if the key doesn't exist.
func (db *MDB) GetDateStr(key int64) string {
	return db.fetchStr(key, "_KVDATE")
}

// SetInt stores an int64 value by key.
func (db *MDB) SetInt(key int64, value int64) {
	tx, err := db.base.Begin()
	if err != nil {
		return
	}
	_, err = db.base.Exec("DELETE FROM _KVINT WHERE Id=?", key)
	if err != nil {
		tx.Rollback()
		return
	}
	_, err = db.base.Exec("INSERT INTO _KVINT (Id, Value) VALUES (?, ?)", key, value)
	if err != nil {
		tx.Rollback()
	}
	tx.Commit()
}

func (db *MDB) setStrValue(store string, key int64, value string) {
	tx, err := db.base.Begin()
	if err != nil {
		return
	}

	db.base.Exec("DELETE FROM "+store+" WHERE Id=?", key)
	_, err = db.base.Exec("INSERT INTO "+store+" (Id, Value) VALUES (?, ?)", key, value)
	if err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()
}

// SetStr stores a string value by key.
func (db *MDB) SetStr(key int64, value string) {
	db.setStrValue("_KVSTR", key, value)
}

// SetBlob stores a byte array by key.
func (db *MDB) SetBlob(key int64, value []byte) {
	db.setStrValue("_KVBLOB", key, string(value))
}

// SetDate stores a time.Time value by key.
func (db *MDB) SetDate(key int64, value time.Time) {
	db.setStrValue("_KVDATE", key, value.UTC().Format(time.RFC3339))
}

// SetDateStr stores a datetime in RFC3339 format by key. The correctness of the string is not validated.
// Use this function in combination with GetDateStr to prevent unnecessary conversions.
func (db *MDB) SetDateStr(key int64, value string) {
	db.setStrValue("_KVDATE", key, value)
}

func (db *MDB) hasKey(key int64, store string) bool {
	var result int
	err := db.base.QueryRow(`SELECT EXISTS (SELECT 1 FROM `+store+` WHERE Id=?);`, key).Scan(&result)
	if err != nil {
		return false
	}
	return true
}

// HasInt returns true if an int value is stored for the key, false otherwise.
func (db *MDB) HasInt(key int64) bool {
	return db.hasKey(key, "_KVINT")
}

// HasStr returns true of a string value is stored for the key, false otherwise.
func (db *MDB) HasStr(key int64) bool {
	return db.hasKey(key, "_KVSTR")
}

// HasBlob returns true if a byte array value is stored for the key, false otherwise.
func (db *MDB) HasBlob(key int64) bool {
	return db.hasKey(key, "_KVBLOB")
}

// HasDate returns true if a time.Time value is stored for the key, false otherwise.
func (db *MDB) HasDate(key int64) bool {
	return db.hasKey(key, "_KVDATE")
}

func (db *MDB) deleteKV(key int64, store string) {
	db.base.Exec(`DELETE FROM `+store+` WHERE Id=?;`, key)
}

// DeleteInt deletes the key and int value for given key. It has no effect if the key-value pair doesn't exist.
func (db *MDB) DeleteInt(key int64) {
	db.deleteKV(key, "_KVINT")
}

// DeleteStr deletes the key and string value for given key. It has no effect if the key-value pair doesn't exist.
func (db *MDB) DeleteStr(key int64) {
	db.deleteKV(key, "_KVSTR")
}

// DeleteBlob deletes the key and string for given key. It has no effect if the key-value pair doesn't exist.
func (db *MDB) DeleteBlob(key int64) {
	db.deleteKV(key, "_KVBLOB")
}

// DeleteDate deletes the key and date value for given key. It has no effect if the key-value pair doesn't exist.
func (db *MDB) DeleteDate(key int64) {
	db.deleteKV(key, "_KVDATE")
}

func (db *MDB) listKV(store string) []int64 {
	result := make([]int64, 0)
	rows, err := db.base.Query(`SELECT Id FROM ` + store)
	defer rows.Close()
	if err != nil {
		return result
	}
	for rows.Next() {
		var n int64
		if err := rows.Scan(&n); err == nil {
			result = append(result, n)
		}
	}
	return result
}

// ListInt lists all int keys.
func (db *MDB) ListInt() []int64 {
	return db.listKV("_KVINT")
}

// ListStr lists all string keys.
func (db *MDB) ListStr() []int64 {
	return db.listKV("_KVSTR")
}

// ListBlob lists all blob keys.
func (db *MDB) ListBlob() []int64 {
	return db.listKV("_KVBLOB")
}

// ListDate lists all date keys.
func (db *MDB) ListDate() []int64 {
	return db.listKV("_KVDATE")
}
