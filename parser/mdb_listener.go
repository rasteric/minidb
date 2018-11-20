// Code generated from Mdb.g4 by ANTLR 4.7.1. DO NOT EDIT.

package parser // Mdb

import "github.com/antlr/antlr4/runtime/Go/antlr"

// MdbListener is a complete listener for a parse tree produced by MdbParser.
type MdbListener interface {
	antlr.ParseTreeListener

	// EnterStart is called when entering the start production.
	EnterStart(c *StartContext)

	// EnterSearchclause is called when entering the searchclause production.
	EnterSearchclause(c *SearchclauseContext)

	// EnterExpr is called when entering the expr production.
	EnterExpr(c *ExprContext)

	// EnterLparen is called when entering the lparen production.
	EnterLparen(c *LparenContext)

	// EnterRparen is called when entering the rparen production.
	EnterRparen(c *RparenContext)

	// EnterUnop is called when entering the unop production.
	EnterUnop(c *UnopContext)

	// EnterRelop is called when entering the relop production.
	EnterRelop(c *RelopContext)

	// EnterSearchop is called when entering the searchop production.
	EnterSearchop(c *SearchopContext)

	// EnterFieldsearch is called when entering the fieldsearch production.
	EnterFieldsearch(c *FieldsearchContext)

	// EnterField is called when entering the field production.
	EnterField(c *FieldContext)

	// EnterTable is called when entering the table production.
	EnterTable(c *TableContext)

	// EnterSearchterm is called when entering the searchterm production.
	EnterSearchterm(c *SearchtermContext)

	// ExitStart is called when exiting the start production.
	ExitStart(c *StartContext)

	// ExitSearchclause is called when exiting the searchclause production.
	ExitSearchclause(c *SearchclauseContext)

	// ExitExpr is called when exiting the expr production.
	ExitExpr(c *ExprContext)

	// ExitLparen is called when exiting the lparen production.
	ExitLparen(c *LparenContext)

	// ExitRparen is called when exiting the rparen production.
	ExitRparen(c *RparenContext)

	// ExitUnop is called when exiting the unop production.
	ExitUnop(c *UnopContext)

	// ExitRelop is called when exiting the relop production.
	ExitRelop(c *RelopContext)

	// ExitSearchop is called when exiting the searchop production.
	ExitSearchop(c *SearchopContext)

	// ExitFieldsearch is called when exiting the fieldsearch production.
	ExitFieldsearch(c *FieldsearchContext)

	// ExitField is called when exiting the field production.
	ExitField(c *FieldContext)

	// ExitTable is called when exiting the table production.
	ExitTable(c *TableContext)

	// ExitSearchterm is called when exiting the searchterm production.
	ExitSearchterm(c *SearchtermContext)
}
