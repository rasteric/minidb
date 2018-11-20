// Code generated from Mdb.g4 by ANTLR 4.7.1. DO NOT EDIT.

package parser // Mdb

import "github.com/antlr/antlr4/runtime/Go/antlr"

// BaseMdbListener is a complete listener for a parse tree produced by MdbParser.
type BaseMdbListener struct{}

var _ MdbListener = &BaseMdbListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseMdbListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseMdbListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseMdbListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseMdbListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterStart is called when production start is entered.
func (s *BaseMdbListener) EnterStart(ctx *StartContext) {}

// ExitStart is called when production start is exited.
func (s *BaseMdbListener) ExitStart(ctx *StartContext) {}

// EnterSearchclause is called when production searchclause is entered.
func (s *BaseMdbListener) EnterSearchclause(ctx *SearchclauseContext) {}

// ExitSearchclause is called when production searchclause is exited.
func (s *BaseMdbListener) ExitSearchclause(ctx *SearchclauseContext) {}

// EnterExpr is called when production expr is entered.
func (s *BaseMdbListener) EnterExpr(ctx *ExprContext) {}

// ExitExpr is called when production expr is exited.
func (s *BaseMdbListener) ExitExpr(ctx *ExprContext) {}

// EnterLparen is called when production lparen is entered.
func (s *BaseMdbListener) EnterLparen(ctx *LparenContext) {}

// ExitLparen is called when production lparen is exited.
func (s *BaseMdbListener) ExitLparen(ctx *LparenContext) {}

// EnterRparen is called when production rparen is entered.
func (s *BaseMdbListener) EnterRparen(ctx *RparenContext) {}

// ExitRparen is called when production rparen is exited.
func (s *BaseMdbListener) ExitRparen(ctx *RparenContext) {}

// EnterUnop is called when production unop is entered.
func (s *BaseMdbListener) EnterUnop(ctx *UnopContext) {}

// ExitUnop is called when production unop is exited.
func (s *BaseMdbListener) ExitUnop(ctx *UnopContext) {}

// EnterRelop is called when production relop is entered.
func (s *BaseMdbListener) EnterRelop(ctx *RelopContext) {}

// ExitRelop is called when production relop is exited.
func (s *BaseMdbListener) ExitRelop(ctx *RelopContext) {}

// EnterSearchop is called when production searchop is entered.
func (s *BaseMdbListener) EnterSearchop(ctx *SearchopContext) {}

// ExitSearchop is called when production searchop is exited.
func (s *BaseMdbListener) ExitSearchop(ctx *SearchopContext) {}

// EnterFieldsearch is called when production fieldsearch is entered.
func (s *BaseMdbListener) EnterFieldsearch(ctx *FieldsearchContext) {}

// ExitFieldsearch is called when production fieldsearch is exited.
func (s *BaseMdbListener) ExitFieldsearch(ctx *FieldsearchContext) {}

// EnterField is called when production field is entered.
func (s *BaseMdbListener) EnterField(ctx *FieldContext) {}

// ExitField is called when production field is exited.
func (s *BaseMdbListener) ExitField(ctx *FieldContext) {}

// EnterTable is called when production table is entered.
func (s *BaseMdbListener) EnterTable(ctx *TableContext) {}

// ExitTable is called when production table is exited.
func (s *BaseMdbListener) ExitTable(ctx *TableContext) {}

// EnterSearchterm is called when production searchterm is entered.
func (s *BaseMdbListener) EnterSearchterm(ctx *SearchtermContext) {}

// ExitSearchterm is called when production searchterm is exited.
func (s *BaseMdbListener) ExitSearchterm(ctx *SearchtermContext) {}
