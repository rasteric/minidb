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
)

type Params struct {
	Argon2_Memory      uint32
	Argon2_Iterations  uint32
	Argon2_Parallelism uint8
	KeyLength          uint32
	SaltLength         uint32
	WeakSaltLength     uint32
}

func DefaultParams() *Params {
	p := Params{
		KeyLength:          512,
		SaltLength:         256,
		WeakSaltLength:     128,
		Argon2_Memory:      64 * 1024,
		Argon2_Iterations:  3,
		Argon2_Parallelism: 4}
	return &p
}

func (p *Params) validate() bool {
	if p.KeyLength >= 64 && p.SaltLength >= 32 &&
		p.WeakSaltLength >= 32 && p.Argon2_Memory >= 16*1024 && p.Argon2_Iterations >= 2 {
		return true
	}
	return false
}

type User struct {
	name string
	id   Item
}

func (u *User) Name() string {
	return u.name
}

func (u *User) ID() Item {
	return u.id
}

type MultiDB struct {
	basepath string
	driver   string
	system   *MDB
	userdbs  map[Item]*MDB
}

// NewMultiDB returns a new multi user database.
func NewMultiDB(basedir string, driver string) (*MultiDB, error) {
	d := filepath.Dir(basedir)
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

func (m *MultiDB) UserDir(user *User) string {
	return filepath.Dir(m.basepath) + string(os.PathSeparator) + user.name
}

func (m *MultiDB) BaseDir() string {
	return filepath.Dir(m.basepath)
}

func (m *MultiDB) userFile(user *User, file string) string {
	return filepath.Dir(m.UserDir(user)) + string(os.PathSeparator) + file
}

func (m *MultiDB) userDBFile(user *User) string {
	return m.userFile(user, "data.sqlite")
}

func (m *MultiDB) systemDBFile() string {
	return m.BaseDir() + string(os.PathSeparator) + "system.sqlite"
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

const (
	OK int = iota + 1
	FAIL
	UsernameInUse
	EmailInUse
	CryptoRandFailure
	InvalidParams
	UnknownUser
	AuthenticationFailed
)

func (m *MultiDB) isExisting(field, query string) bool {
	q, err := ParseQuery(fmt.Sprintf("User %s=%s", field, query))
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
	if len(results) != 1 {
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
func (m *MultiDB) NewUser(username, email, password string, params *Params) (*User, int, error) {
	if err := validateUser(username, m.BaseDir()); err != nil {
		return nil, FAIL, err
	}
	if !params.validate() {
		return nil, InvalidParams, Fail(`invalid parameters`)
	}
	user := User{name: username}
	dirpath := m.UserDir(&user)
	err := CreateDirIfNotExist(dirpath)
	if err != nil {
		return nil, FAIL, err
	}
	if m.system == nil {
		return nil, FAIL, Fail(`the internal database was closed or is locked`)
	}
	if !m.system.TableExists("User") {
		err := m.system.AddTable("User", []Field{Field{Name: "Username", Sort: DBString},
			Field{Name: "Email", Sort: DBString},
			Field{Name: "Key", Sort: DBBlob},
			Field{Name: "WeakSalt", Sort: DBBlob},
			Field{Name: "StrongSalt", Sort: DBBlob},
			Field{Name: "Created", Sort: DBDate},
			Field{Name: "Modified", Sort: DBDate}})
		if err != nil {
			return nil, FAIL, err
		}
	}
	if m.ExistingUser(username) {
		return nil, UsernameInUse, Fail(`user "%s" already exists!`, username)
	}
	if m.ExistingEmail(email) {
		return nil, EmailInUse, Fail(`email "%s" is already in use!`, email)
	}
	// now start adding the user
	tx, err := m.system.base.Begin()
	if err != nil {
		return nil, FAIL, err
	}
	defer tx.Rollback()
	user.id, err = m.system.NewItem("User")
	if err := m.system.Set("User", user.id, "Username", []Value{NewString(username)}); err != nil {
		return nil, FAIL, err
	}
	if err := m.system.Set("User", user.id, "Email", []Value{NewString(email)}); err != nil {
		return nil, FAIL, err
	}
	salt := make([]byte, params.SaltLength)
	n, err := rand.Read(salt)
	if uint32(n) != params.SaltLength || err != nil {
		return nil, CryptoRandFailure, Fail(`random number generator failed to generate salt`)
	}
	if err := m.system.Set("User", user.id, "Salt", []Value{NewBytes(salt)}); err != nil {
		return nil, FAIL, Fail(`could not store salt in multiuser database: %s`, err)
	}
	key := argon2.IDKey([]byte(password),
		salt, params.Argon2_Iterations, params.Argon2_Memory,
		params.Argon2_Parallelism, params.KeyLength)
	if err := m.system.Set("User", user.id, "Key", []Value{NewBytes(key)}); err != nil {
		return nil, FAIL, Fail(`could not store key in multiuser database: %s`, err)
	}
	if err := tx.Commit(); err != nil {
		return nil, FAIL, Fail(`multiuser database error: %s`, err)
	}
	now := NewDate(time.Now())
	if err := m.system.Set("User", user.id, "Created", []Value{now}); err != nil {
		return nil, FAIL, err
	}
	if err := m.system.Set("User", user.id, "Modified", []Value{now}); err != nil {
		return nil, FAIL, err
	}
	return &user, OK, nil
}

// UserSalt is the weak salt associated with a user. It is stored in the database and may
// be used for hashing the password prior to authentication. The weak salt is not used
// for internal key derivation.
func (m *MultiDB) UserSalt(username string) ([]byte, int, error) {
	if !m.ExistingUser(username) {
		return nil, UnknownUser, Fail(`unknown user "%s"`, username)
	}
	id := m.userID(username)
	if id == 0 {
		return nil, UnknownUser, Fail(`unknown user "%s"`, username)
	}
	result, err := m.system.Get("Users", id, "WeakSalt")
	if err != nil || len(result) != 1 {
		return nil, FAIL, Fail(`user "%s" salt not found, the user database might be corrupted`, username)
	}
	return result[0].Bytes(), OK, nil
}

// Authenticate a user by given name and passwords using the given parameters.
// Returns the user and OK if successful, otherwise nil, a numeric error code and the error.
func (m *MultiDB) Authenticate(username, password string, params *Params) (*User, int, error) {
	if err := validateUser(username, m.BaseDir()); err != nil {
		return nil, FAIL, err
	}
	if !m.ExistingUser(username) {
		return nil, UnknownUser, Fail(`user "%s" does not exist`, username)
	}
	if !params.validate() {
		return nil, FAIL, Fail(`invalid parameters`)
	}
	user := User{name: username}
	dirpath := m.UserDir(&user)
	if _, err := os.Stat(dirpath); os.IsNotExist(err) {
		return nil, FAIL, Fail(`user "%s" home directory does not exist: %s`, username, dirpath)
	}
	user.id = m.userID(username)
	if user.id == 0 {
		return nil, FAIL, Fail(`user "%s" does not exist`, username)
	}
	// get the strong salt and hash with it using argon2, compare to stored key
	result, err := m.system.Get("Users", user.id, "StrongSalt")
	if err != nil || len(result) != 1 {
		return nil, FAIL, Fail(`user "%s" strong salt not found, the user database might be corrupted`, username)
	}
	salt := result[0].Bytes()
	if len(salt) != int(params.SaltLength) {
		return nil, InvalidParams, Fail(`invalid params, user "%s" salt length does not match selt length in params, given %d, expected %d`, username, len(salt), params.SaltLength)
	}
	keyA := argon2.IDKey([]byte(password),
		salt, params.Argon2_Iterations, params.Argon2_Memory,
		params.Argon2_Parallelism, params.KeyLength)
	keyresult, err := m.system.Get("Users", user.id, "Key")
	if err != nil || len(keyresult) != 1 {
		return nil, FAIL, Fail(`user "%s" key was not found in user database, the database might be corrupted`, username)
	}
	keyB := keyresult[0].Bytes()
	if !bytes.Equal(keyA, keyB) {
		return nil, AuthenticationFailed, Fail(`authentication failure`)
	}
	return &user, OK, nil
}

// Close the MultiDB, closing the internal housekeeping and all open user databases.
func (m *MultiDB) Close() error {
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
		return Fail(`errors closing multi user DB: %s`, s)
	}
	return nil
}

// UserDB returns the database of the given user.
func (m *MultiDB) UserDB(user *User) (*MDB, error) {
	var err error
	if user.id == 0 {
		return nil, Fail(`user "%s" does not exist`, user.name)
	}
	db := m.userdbs[user.id]
	if db == nil {
		db, err = Open(m.driver, m.userDBFile(user))
		if err != nil {
			return nil, err
		}
	}
	return db, nil
}

// DeleteUserContent deletes a user's content in the multiuser database, i.e., all the user data.
// This action cannot be undone.
func (m *MultiDB) DeleteUserContent(user *User) error {
	db, _ := m.UserDB(user)
	db.Close()
	return removeContents(m.UserDir(user))
}

// DeleteUser deletes a user and all associated user content from a multiuser database.
func (m *MultiDB) DeleteUser(user *User) error {
	// todo
	return nil
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
func (m *MultiDB) Delete() error {
	m.Close()
	m.system.Close()
	return removeContents(m.BaseDir())
}

// Archive stores the user data in a packed zip file but does not close or remove the user.
// This can be used for backups or for archiving.
func (m *MultiDB) ArchiveUser(user *User, archivedir string) error {
	db, err := m.UserDB(user)
	if err != nil {
		return err
	}
	if err := db.Close(); err != nil {
		return err
	}
	source := m.UserDir(user)
	filename := fmt.Sprintf("%s-%d_%s.multidb", user.Name(), int64(user.ID()), time.Now().UTC().Format(time.RFC3339))
	result, err := packdir.Pack(source, filename, archivedir, packdir.GOOD_COMPRESSION, 0)
	if err != nil {
		return err
	}
	if result.ScanErrNum > 0 {
		return Fail(`archiving failed, unable to pack %d files (insufficient permissions?)`, result.ScanErrNum)
	}
	if result.ArchiveErrNum > 0 {
		return Fail(`archiving failed, %d files were not properly archived (insufficient permissions?)`,
			result.ArchiveErrNum)
	}
	delete(m.userdbs, user.ID())
	return nil
}
