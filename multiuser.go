package minidb

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/rasteric/packdir"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/blake2b"
)

// Params contain all the parameters that are used by a multiuser database.
type Params struct {
	Argon2Memory       uint32
	Argon2Iterations   uint32
	Argon2Parallelism  uint8
	KeyLength          uint32
	InternalSaltLength uint32
	ExternalSaltLength uint32
}

// DefaultParams returns parameters with reasonable default values that are safe to use.
// Be aware that default parameters may change from release to release to reflect
// updates and changes in security requirements.
func DefaultParams() *Params {
	p := Params{
		KeyLength:          512,
		InternalSaltLength: 256,
		ExternalSaltLength: 256,
		Argon2Memory:       64 * 1024,
		Argon2Iterations:   3,
		Argon2Parallelism:  4}
	return &p
}

func (p *Params) validate() bool {
	if p.KeyLength >= 64 && p.InternalSaltLength >= 32 &&
		p.ExternalSaltLength >= 32 && p.Argon2Memory >= 16*1024 && p.Argon2Iterations >= 2 {
		return true
	}
	return false
}

// User represents a user.
type User struct {
	name string
	id   Item
}

// Name returns the name of the user.
func (u *User) Name() string {
	return u.name
}

// ID returns the ID of the user.
func (u *User) ID() Item {
	return u.id
}

// MultiDB contains all information needed for housekeeping multiple DBs, except for the parameters
// and context-specific information like passwords.
type MultiDB struct {
	basepath string
	driver   string
	system   *MDB
	userdbs  map[Item]*MDB
}

// NewMultiDB returns a new multi user database.
func NewMultiDB(basedir string, driver string) (*MultiDB, error) {
	d := filepath.Clean(basedir)
	if !validDir(d) {
		return nil, Fail(`the base directory "%s" does not exist or has incorrect permissions`, d)
	}
	db := MultiDB{basepath: basedir}
	thedb := &db
	sys, err := Open(driver, thedb.systemDBFile())
	if err != nil {
		return nil, err
	}
	thedb.system = sys
	thedb.driver = driver
	thedb.userdbs = make(map[Item]*MDB)
	return thedb, nil
}

// UserDir returns the given user's directory where the user database is stored.
func (m *MultiDB) UserDir(user *User) string {
	return filepath.Join(m.basepath, user.name)
}

// BaseDir returns the base directory of the multiuser database. This directory contains databases
// for all users.
func (m *MultiDB) BaseDir() string {
	return m.basepath
}

func (m *MultiDB) userFile(user *User, file string) string {
	return filepath.Join(m.UserDir(user), file)
}

func (m *MultiDB) userDBFile(user *User) string {
	return m.userFile(user, "data.sqlite")
}

func (m *MultiDB) systemDBFile() string {
	return filepath.Join(m.BaseDir(), "system.sqlite")
}

func validUserName(name string) bool {
	var validUser = regexp.MustCompile(`^\p{L}+[_0-9\p{L}]*$`)
	return validUser.MatchString(name)
}

func validDir(dir string) bool {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return false
	}
	return true
}

// CreateDirIfNotExists creates a directory including all subdirectories needed,
// or returns an error.
func CreateDirIfNotExist(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func validateUser(name string, basedir string) error {
	if !validUserName(name) {
		return Fail(`invalid user name "%s"`, name)
	}
	if !validDir(basedir) {
		return Fail(`the base directory for user "%s" does not exist: %s`, name, basedir)
	}
	src, err := os.Stat(basedir)
	if err != nil {
		return err
	}
	if !src.IsDir() {
		return Fail(`not a directory: %s`, basedir)
	}
	return nil
}

// ErrCode types represent errors instead of error structures.
type ErrCode int

// Error codes returned by the functions.
const (
	ErrAuthenticationFailed ErrCode = iota + 1 // User authentication has failed (wrong password).
	OK                                         // No error has occured.
	ErrUsernameInUse                           // The user name is already being used.
	ErrEmailInUse                              // The email is already being used.
	ErrCryptoRandFailure                       // The random number generator has failed.
	ErrInvalidParams                           // One or more parameters were invalid.
	ErrUnknownUser                             // The user is not known.
	ErrNotEnoughSalt                           // Insufficiently long salt has been supplied.
	ErrInvalidUser                             // The user name or email is invalid.
	ErrDBClosed                                // The internal housekeeping DB is locked, corrupted, or closed.
	ErrDBFail                                  // A database operation has failed.
	ErrFileSystem                              // A directory of file could not be created.
	ErrNoHome                                  // The user's DB home directory does not exist.
	ErrCloseFailed                             // Could not close the user database.
	ErrOpenFailed                              // Could not open the user database.
	ErrPackFail                                // Compressing user data failed.
	ErrInvalidKey                              // A given salted key is invalid (either nil, or other problems).
)

func (m *MultiDB) isExisting(field, query string) bool {
	q, _ := ParseQuery(fmt.Sprintf("User %s=%s", field, query))
	results, err := m.system.Find(q, 1)
	if err != nil || len(results) < 1 {
		return false
	}
	return true
}

func (m *MultiDB) userID(username string) Item {
	q, err := ParseQuery(fmt.Sprintf("User Username=%s", username))
	if err != nil {
		return 0
	}
	results, err := m.system.Find(q, 1)
	if err != nil || len(results) != 1 {
		return 0
	}
	return results[0]
}

// ExistingUser returns true if a user with the given user name exists, false otherwise.
func (m *MultiDB) ExistingUser(username string) bool {
	return m.isExisting("Username", username)
}

// ExistingEmail returns true if a user with this email address exists, false otherwise.
func (m *MultiDB) ExistingEmail(email string) bool {
	return m.isExisting("Email", email)
}

// NewUser creates a new user with given username, email, and password. Based on a strong
// salt that is only used internally and the Argon2 algorithm with the given parameters
// an internal key is created and stored in an internal database. The user and OK are returned
// unless an error has occurred. The integer returned is a numeric error code to make it easier to distinguish
// certain cases: EmailInUse - the email has already been registered, UsernameInUse - a user with the same
// user name has already been registered. Both emails and usernames must be unique and cannot be
// registered twice.
func (m *MultiDB) NewUser(username, email string, key *saltedKey) (*User, ErrCode, error) {
	// validate inputs
	if err := validateUser(username, m.BaseDir()); err != nil {
		return nil, ErrInvalidUser, err
	}
	reply, err := key.validate()
	if err != nil || reply != OK {
		return nil, reply, err
	}
	user := User{name: username}
	if m.system == nil {
		return nil, ErrDBClosed, Fail(`internal DB is nil`)
	}
	// maybe create the user table
	if !m.system.TableExists("User") {
		err := m.system.AddTable("User", []Field{Field{Name: "Username", Sort: DBString},
			Field{Name: "Email", Sort: DBString},
			Field{Name: "Key", Sort: DBBlob},
			Field{Name: "ExternalSalt", Sort: DBBlob},
			Field{Name: "InternalSalt", Sort: DBBlob},
			Field{Name: "Created", Sort: DBDate},
			Field{Name: "Modified", Sort: DBDate}})
		if err != nil {
			return nil, ErrDBFail, err
		}
	}
	// check if user and email exist
	if m.ExistingUser(username) {
		return nil, ErrUsernameInUse, Fail(`user "%s" already exists!`, username)
	}
	if m.ExistingEmail(email) {
		return nil, ErrEmailInUse, Fail(`email "%s" is already in use!`, email)
	}
	// now start adding the user in a transaction
	tx, err := m.system.base.Begin()
	if err != nil {
		return nil, ErrDBFail, err
	}
	defer tx.Rollback()
	user.id, err = m.system.NewItem("User")
	if err != nil {
		return nil, ErrDBFail, err
	}
	if err := m.system.Set("User", user.id, "Username", []Value{NewString(username)}); err != nil {
		return nil, ErrDBFail, err
	}
	if err := m.system.Set("User", user.id, "Email", []Value{NewString(email)}); err != nil {
		return nil, ErrDBFail, err
	}
	salt := make([]byte, key.p.InternalSaltLength)
	n, err := rand.Read(salt)
	if uint32(n) != key.p.InternalSaltLength || err != nil {
		return nil, ErrCryptoRandFailure, Fail(`random number generator failed to generate salt`)
	}
	if err := m.system.Set("User", user.id, "InternalSalt", []Value{NewBytes(salt)}); err != nil {
		return nil, ErrDBFail, Fail(`could not store salt in multiuser database: %s`, err)
	}
	realkey := argon2.IDKey(key.pwd,
		salt, key.p.Argon2Iterations, key.p.Argon2Memory,
		key.p.Argon2Parallelism, key.p.KeyLength)
	if err := m.system.Set("User", user.id, "Key", []Value{NewBytes(realkey)}); err != nil {
		return nil, ErrDBFail, Fail(`could not store key in multiuser database: %s`, err)
	}
	if err := m.system.Set("User", user.id, "ExternalSalt", []Value{NewBytes(key.sel)}); err != nil {
		return nil, ErrDBFail, Fail(`could not store the external salt in multiuser database: %s`, err)
	}
	now := NewDate(time.Now())
	if err := m.system.Set("User", user.id, "Created", []Value{now}); err != nil {
		return nil, ErrDBFail, err
	}
	if err := m.system.Set("User", user.id, "Modified", []Value{now}); err != nil {
		return nil, ErrDBFail, err
	}
	dirpath := m.UserDir(&user)
	err = CreateDirIfNotExist(dirpath)
	if err != nil {
		return nil, ErrFileSystem, err
	}
	if err := tx.Commit(); err != nil {
		return nil, ErrDBFail, Fail(`multiuser database error: %s`, err)
	}
	return &user, OK, nil
}

// ExternalSalt is the salt associated with a user. It is stored in the database and may
// be used for hashing the password prior to authentication. The external salt is not used
// for internal key derivation.
func (m *MultiDB) ExternalSalt(username string) ([]byte, ErrCode, error) {
	if !m.ExistingUser(username) {
		return nil, ErrUnknownUser, Fail(`unknown user "%s"`, username)
	}
	id := m.userID(username)
	if id == 0 {
		return nil, ErrUnknownUser, Fail(`unknown user "%s"`, username)
	}
	result, err := m.system.Get("User", id, "ExternalSalt")
	if err != nil || len(result) != 1 {
		return nil, ErrNotEnoughSalt, Fail(`user "%s" salt not found, the user database might be corrupted`, username)
	}
	return result[0].Bytes(), OK, nil
}

// GenerateExternalSalt returns some new external salt of the length specified in params.
// This salt should be passed to NewUser and can be used for passphrase hashing prior to
// calling NewUser. It is stored in the user database and can be retrieved as ExternalSalt.
func GenerateExternalSalt(params *Params) []byte {
	salt := make([]byte, params.ExternalSaltLength)
	n, err := rand.Read(salt)
	if err != nil || uint32(n) < params.ExternalSaltLength {
		return nil
	}
	return salt
}

type saltedKey struct {
	pwd []byte
	sel []byte
	p   *Params
}

func (key *saltedKey) validate() (ErrCode, error) {
	if key == nil {
		return ErrInvalidKey, Fail(`key is nil`)
	}
	if key.p == nil {
		return ErrInvalidParams, Fail(`key parameters are nil`)
	}
	if key.pwd == nil {
		return ErrInvalidKey, Fail(`key password is empty`)
	}
	if key.sel == nil {
		return ErrNotEnoughSalt, Fail(`key salt is nil`)
	}
	if !key.p.validate() {
		return ErrInvalidParams, Fail(`invalid parameters`)
	}
	if uint32(len(key.sel)) < key.p.ExternalSaltLength {
		return ErrNotEnoughSalt, Fail(`external key salt length is less than required by key params`)
	}
	return OK, nil
}

// GenerateKey takes a password and some salt, and generates a salted key of length 64 bytes.
// Use the ExternalSalt as salt and the original, unaltered password. The function
// use Blake2b-512 for key derivation.
func GenerateKey(password string, salt []byte, params *Params) *saltedKey {
	unsalted := blake2b.Sum512([]byte(password))
	salted := saltedKey{pwd: append(salt, unsalted[:]...), sel: salt, p: params}
	return &salted
}

// Authenticate a user by given name and salted password.
// Returns the user and OK if successful, otherwise nil, a numeric error code and the error.
// Notice that the external salt is not passed to this function. Instead, the password string
// should have been prepared (securely hashed, whitened, etc.) before calling this function
// on the basis of the user's ExternalSalt.
func (m *MultiDB) Authenticate(username string, key *saltedKey) (*User, ErrCode, error) {
	if err := validateUser(username, m.BaseDir()); err != nil {
		return nil, ErrInvalidUser, err
	}
	if !m.ExistingUser(username) {
		return nil, ErrUnknownUser, Fail(`user "%s" does not exist`, username)
	}
	reply, err := key.validate()
	if err != nil || reply != OK {
		return nil, reply, err
	}
	user := User{name: username}
	dirpath := m.UserDir(&user)
	if _, err := os.Stat(dirpath); os.IsNotExist(err) {
		return nil, ErrNoHome, Fail(`user "%s" home directory does not exist: %s`, username, dirpath)
	}
	user.id = m.userID(username)
	if user.id == 0 {
		return nil, ErrUnknownUser, Fail(`user "%s" does not exist`, username)
	}
	// get the strong salt and hash with it using argon2, compare to stored key
	result, err := m.system.Get("User", user.id, "InternalSalt")
	if err != nil || len(result) != 1 {
		return nil, ErrNotEnoughSalt,
			Fail(`user "%s"'s internal salt was not found, the user database might be corrupted`, username)
	}
	salt := result[0].Bytes()
	if len(salt) != int(key.p.InternalSaltLength) {
		return nil, ErrInvalidParams,
			Fail(`invalid params, user "%s"'s internal salt length does not match internal salt length in params, given %d, expected %d`, username, len(salt), key.p.InternalSaltLength)
	}
	keyA := argon2.IDKey(key.pwd,
		salt, key.p.Argon2Iterations, key.p.Argon2Memory,
		key.p.Argon2Parallelism, key.p.KeyLength)
	keyresult, err := m.system.Get("User", user.id, "Key")
	if err != nil || len(keyresult) != 1 {
		return nil, ErrAuthenticationFailed,
			Fail(`user "%s" key was not found in user database, the database might be corrupted`, username)
	}
	keyB := keyresult[0].Bytes()
	if !bytes.Equal(keyA, keyB) {
		return nil, ErrAuthenticationFailed, Fail(`authentication failure`)
	}
	return &user, OK, nil
}

// Close the MultiDB, closing the internal housekeeping and all open user databases.
func (m *MultiDB) Close() (ErrCode, error) {
	errcount := 0
	s := ""
	for _, v := range m.userdbs {
		if v != nil {
			err := v.Close()
			if err != nil {
				s = fmt.Sprintf("%s, %s", s, err.Error())
				errcount++
			}
		}
	}
	for k := range m.userdbs {
		delete(m.userdbs, k)
	}
	if err := m.system.Close(); err != nil {
		s = fmt.Sprintf("%s, %s", s, err.Error())
		errcount++
	}
	if errcount > 0 {
		return ErrCloseFailed, Fail(`errors closing multi user DB: %s`, s)
	}
	return OK, nil
}

// UserDB returns the database of the given user.
func (m *MultiDB) UserDB(user *User) (*MDB, ErrCode, error) {
	var err error
	if user.id == 0 {
		return nil, ErrUnknownUser, Fail(`user "%s" does not exist`, user.name)
	}
	db := m.userdbs[user.id]
	if db == nil {
		db, err = Open(m.driver, m.userDBFile(user))
		if err != nil {
			return nil, ErrOpenFailed, err
		}
	}
	return db, OK, nil
}

// DeleteUserContent deletes a user's content in the multiuser database, i.e., all the user data.
// This action cannot be undone.
func (m *MultiDB) DeleteUserContent(user *User) (ErrCode, error) {
	db, _, _ := m.UserDB(user)
	db.Close()
	if err := removeContents(m.UserDir(user)); err != nil {
		return ErrFileSystem, err
	}
	return OK, nil
}

// DeleteUser deletes a user and all associated user content from a multiuser database.
func (m *MultiDB) DeleteUser(user *User) (ErrCode, error) {
	if err := m.system.RemoveItem("User", user.ID()); err != nil {
		return ErrDBFail, err
	}
	errcode, err := m.DeleteUserContent(user)
	delete(m.userdbs, user.ID())
	if err != nil {
		return errcode, err
	}
	return OK, nil
}

func removeContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

// Delete deletes the whole multiuser db including all housekeeping information and directories.
// This action cannot be undone.
func (m *MultiDB) Delete() (ErrCode, error) {
	m.Close()
	m.system.Close()
	if err := removeContents(m.BaseDir()); err != nil {
		return ErrFileSystem, err
	}
	return OK, nil
}

// ArchiveUser stores the user data in a packed zip file but does not close or remove the user.
// This can be used for backups or for archiving.
func (m *MultiDB) ArchiveUser(user *User, archivedir string) (ErrCode, error) {
	db, reply, err := m.UserDB(user)
	if err != nil {
		return reply, err
	}
	if err := db.Close(); err != nil {
		return ErrCloseFailed, err
	}
	source := m.UserDir(user)
	filename := fmt.Sprintf("%s-%d_%s.multidb", user.Name(), int64(user.ID()), time.Now().UTC().Format(time.RFC3339))
	result, err := packdir.Pack(source, filename, archivedir, packdir.GOOD_COMPRESSION, 0)
	if err != nil {
		return ErrPackFail, err
	}
	if result.ScanErrNum > 0 {
		return ErrFileSystem,
			Fail(`archiving failed, unable to pack %d files (insufficient permissions?)`, result.ScanErrNum)
	}
	if result.ArchiveErrNum > 0 {
		return ErrPackFail,
			Fail(`archiving failed, %d files were not properly archived (insufficient permissions?)`,
				result.ArchiveErrNum)
	}
	delete(m.userdbs, user.ID())
	return OK, nil
}

func (m *MultiDB) UserEmail(user *User) (string, ErrCode, error) {
	if user == nil || !m.ExistingUser(user.name) {
		return "", ErrUnknownUser, Fail(`user "%s" does not exist`, user.name)
	}
	result, err := m.system.Get("User", user.ID(), "Email")
	if err != nil {
		return "", ErrUnknownUser, Fail(`user "%s" does not exist`, user.name)
	}
	if len(result) != 1 {
		return "", ErrDBFail, Fail(`non unique result for user "%s" email`, user.name)
	}
	return result[0].String(), OK, nil
}
