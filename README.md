 # Minidb
*-- a minimalist database for Golang and a key-value database command line tool*

[![GoDoc](https://godoc.org/github.com/rasteric/minidb/go?status.svg)](https://godoc.org/github.com/rasteric/minidb)
[![Go Report Card](https://goreportcard.com/badge/github.com/rasteric/minidb)](https://goreportcard.com/report/github.com/rasteric/minidb)

Minidb is an early version of an SQL database wrapper library and a command line database written in Go. It currently allows you to create tables with "fields", where each field may contain a string, int, blob, or date. It also has types string-list, int-list, blob-list, and date-list. Tables and their fields can then be queried by the command line tool _minidb_. Use the --help command line option for more information about the CLI tool.

The database uses an existing SQL driver and wraps around it. The command line tool uses Sqlite3 and the library is also only tested with Sqlite. I try to avoid using Sqlite-specific constructs but currently do not guarantee that it will work with other SQL databases.

## The Library

You can use `go get github.com/rasteric/minidb` to import the library. The library for Go has two APIs. The direct API provides functions for manipulating the database, most of which work on the basis of an MDB structure. This structure stores the driver and is obtained via the `Open` function. The direct API functions are pretty straightforward wrapper to the underlying SQL database. Although there are many internal error checks, you ought never manipulate the underlying database directly, though.

The indirect API uses `Command` and `Result` structures that provide an additional abstraction layer on top of the direct API. These structures can be marshalled and unmarshalled to JSON, which allows them to be used in client/server architectures. A function `Exec` executes a `Command` and returns a `Result`. For convenience, functions are provided that return commands and take a numerical database id instead of a pointer to MDB as the first argument, but otherwise mirror the direct API exactly. For example, the counterpart to `(db *MDB) GetFields(table string) ([]Field, error)` is `GetFieldsCommand(db CommandDB, table string) *Command`.

The command line tool in the `cmd` directory uses the indirect API to implement inter-process communication between the command line tool `cmd/minidb/minidb` and the local server in `cmd/mdbserve/mdbserve`.

Public functions in the source code are commented unless they are easy to read.

## The Command Line Tool

The command line tool is in `cmd/minidb/minidb` and uses a relative path to the server executable by default. Use the `--help` option to get information about its usage. One important thing that is not documented yet, because it's not fully finished, is the find query syntax. It works like this: You can only search within one table at a time and combine field queries with boolean operators `and`, `or`, and `not`. List fields allow additional prefix-operators `every` and `no`. The actual query format is `fieldname=like-clause`.

### Examples:

`minidb table Person string-list Name string ZIP int Age Email string`

creates a table Person with a Name field that can store a list of strings, a ZIP field that stores a string, an Age field that stores an integer, and an Email field that stores a string.

`minidb new Person`

returns a numeric id for a new Person. 

`minidb set Person 1 Name John Caesar Smith Jr`

sets the name of Person 1 to the list of strings "John", "Caesar", "Smith", "Jr" (since the shell splits up the strings in this way).

`minidb find Person Name=%John%`

look for every person whose name contains the string "John".

`minidb find Person every Name=%r% and ZIP=111%`

look for every Person whose name is such that every name string contains the letter "r" and whose ZIP string starts with "111". This assumes that Person Name is a string-list.

`minidb find Person no Name=John`

assuming that Person Name is of type string-list, this query matches all Persons for which no Name list entry is "John".

`minidb find Person Name=John or Name=Smith%`

find every Person whose Name is exactly "John" (in one of its Name fields, if it is a string-list) or whose name starts with "Smith" (in one of its Name entries, if it is a string-list).

`minidb set-str 1 "Hello world!"`

sets the string with numeric key 1 to "Hello world!"

`minidb get-str 1`

returns "Hello world!" + newline if the previous command has been executed before.

In the key-value interface all keys are integers.


