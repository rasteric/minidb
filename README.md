# minidb
*-- a minimalist DB wrapper for Golang and a key-value database command line tool*

This is the early alpha version of a little command line database written in Go. It currently allows you to create tables with "fields", where each field may contain a string, int, or blob *or* a string-list, int-list, or blob-list. Tables and their fields can then be queried by the command line tool _minidb_. Use the --help command line option for more information about the CLI tool.

Public functions in the source code should be commented unless they are easy to read.

## Future Plans

This is going to be split up into a library that accepts any SQL driver and the command line tool. I plan to use the library in other projects.

Todo: write tests, etc.



