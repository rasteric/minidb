// Code generated from Mdb.g4 by ANTLR 4.7.1. DO NOT EDIT.

package parser // Mdb

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

// Suppress unused import errors
var _ = fmt.Printf
var _ = reflect.Copy
var _ = strconv.Itoa

var parserATN = []uint16{
	3, 24715, 42794, 33075, 47597, 16764, 15335, 30598, 22884, 3, 15, 85, 4,
	2, 9, 2, 4, 3, 9, 3, 4, 4, 9, 4, 4, 5, 9, 5, 4, 6, 9, 6, 4, 7, 9, 7, 4,
	8, 9, 8, 4, 9, 9, 9, 4, 10, 9, 10, 4, 11, 9, 11, 4, 12, 9, 12, 4, 13, 9,
	13, 3, 2, 3, 2, 3, 2, 3, 3, 3, 3, 3, 3, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4, 3,
	4, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4, 5, 4, 47, 10, 4, 3,
	4, 3, 4, 3, 4, 3, 4, 7, 4, 53, 10, 4, 12, 4, 14, 4, 56, 11, 4, 3, 5, 3,
	5, 3, 6, 3, 6, 3, 7, 3, 7, 3, 8, 3, 8, 3, 9, 3, 9, 3, 10, 3, 10, 3, 10,
	3, 10, 3, 11, 3, 11, 3, 12, 3, 12, 3, 13, 6, 13, 77, 10, 13, 13, 13, 14,
	13, 78, 3, 13, 3, 13, 5, 13, 83, 10, 13, 3, 13, 2, 3, 6, 14, 2, 4, 6, 8,
	10, 12, 14, 16, 18, 20, 22, 24, 2, 4, 3, 2, 5, 6, 3, 2, 8, 9, 2, 79, 2,
	26, 3, 2, 2, 2, 4, 29, 3, 2, 2, 2, 6, 46, 3, 2, 2, 2, 8, 57, 3, 2, 2, 2,
	10, 59, 3, 2, 2, 2, 12, 61, 3, 2, 2, 2, 14, 63, 3, 2, 2, 2, 16, 65, 3,
	2, 2, 2, 18, 67, 3, 2, 2, 2, 20, 71, 3, 2, 2, 2, 22, 73, 3, 2, 2, 2, 24,
	82, 3, 2, 2, 2, 26, 27, 5, 4, 3, 2, 27, 28, 7, 2, 2, 3, 28, 3, 3, 2, 2,
	2, 29, 30, 5, 22, 12, 2, 30, 31, 5, 6, 4, 2, 31, 5, 3, 2, 2, 2, 32, 33,
	8, 4, 1, 2, 33, 47, 5, 18, 10, 2, 34, 35, 5, 16, 9, 2, 35, 36, 5, 18, 10,
	2, 36, 47, 3, 2, 2, 2, 37, 38, 5, 12, 7, 2, 38, 39, 5, 6, 4, 5, 39, 47,
	3, 2, 2, 2, 40, 41, 5, 8, 5, 2, 41, 42, 5, 6, 4, 2, 42, 43, 5, 14, 8, 2,
	43, 44, 5, 6, 4, 2, 44, 45, 5, 10, 6, 2, 45, 47, 3, 2, 2, 2, 46, 32, 3,
	2, 2, 2, 46, 34, 3, 2, 2, 2, 46, 37, 3, 2, 2, 2, 46, 40, 3, 2, 2, 2, 47,
	54, 3, 2, 2, 2, 48, 49, 12, 4, 2, 2, 49, 50, 5, 14, 8, 2, 50, 51, 5, 6,
	4, 5, 51, 53, 3, 2, 2, 2, 52, 48, 3, 2, 2, 2, 53, 56, 3, 2, 2, 2, 54, 52,
	3, 2, 2, 2, 54, 55, 3, 2, 2, 2, 55, 7, 3, 2, 2, 2, 56, 54, 3, 2, 2, 2,
	57, 58, 7, 3, 2, 2, 58, 9, 3, 2, 2, 2, 59, 60, 7, 4, 2, 2, 60, 11, 3, 2,
	2, 2, 61, 62, 7, 7, 2, 2, 62, 13, 3, 2, 2, 2, 63, 64, 9, 2, 2, 2, 64, 15,
	3, 2, 2, 2, 65, 66, 9, 3, 2, 2, 66, 17, 3, 2, 2, 2, 67, 68, 5, 20, 11,
	2, 68, 69, 7, 10, 2, 2, 69, 70, 5, 24, 13, 2, 70, 19, 3, 2, 2, 2, 71, 72,
	7, 12, 2, 2, 72, 21, 3, 2, 2, 2, 73, 74, 7, 12, 2, 2, 74, 23, 3, 2, 2,
	2, 75, 77, 7, 13, 2, 2, 76, 75, 3, 2, 2, 2, 77, 78, 3, 2, 2, 2, 78, 76,
	3, 2, 2, 2, 78, 79, 3, 2, 2, 2, 79, 83, 3, 2, 2, 2, 80, 83, 7, 12, 2, 2,
	81, 83, 7, 14, 2, 2, 82, 76, 3, 2, 2, 2, 82, 80, 3, 2, 2, 2, 82, 81, 3,
	2, 2, 2, 83, 25, 3, 2, 2, 2, 6, 46, 54, 78, 82,
}
var deserializer = antlr.NewATNDeserializer(nil)
var deserializedATN = deserializer.DeserializeFromUInt16(parserATN)

var literalNames = []string{
	"", "'('", "')'", "'and'", "'or'", "'not'", "'no'", "'every'", "'='",
}
var symbolicNames = []string{
	"", "", "", "AND", "OR", "NOT", "NO", "EVERY", "EQ", "NOT_SPECIAL", "ID",
	"DIGIT", "STRING", "WS",
}

var ruleNames = []string{
	"start", "searchclause", "expr", "lparen", "rparen", "unop", "relop", "searchop",
	"fieldsearch", "field", "table", "searchterm",
}
var decisionToDFA = make([]*antlr.DFA, len(deserializedATN.DecisionToState))

func init() {
	for index, ds := range deserializedATN.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(ds, index)
	}
}

type MdbParser struct {
	*antlr.BaseParser
}

func NewMdbParser(input antlr.TokenStream) *MdbParser {
	this := new(MdbParser)

	this.BaseParser = antlr.NewBaseParser(input)

	this.Interpreter = antlr.NewParserATNSimulator(this, deserializedATN, decisionToDFA, antlr.NewPredictionContextCache())
	this.RuleNames = ruleNames
	this.LiteralNames = literalNames
	this.SymbolicNames = symbolicNames
	this.GrammarFileName = "Mdb.g4"

	return this
}

// MdbParser tokens.
const (
	MdbParserEOF         = antlr.TokenEOF
	MdbParserT__0        = 1
	MdbParserT__1        = 2
	MdbParserAND         = 3
	MdbParserOR          = 4
	MdbParserNOT         = 5
	MdbParserNO          = 6
	MdbParserEVERY       = 7
	MdbParserEQ          = 8
	MdbParserNOT_SPECIAL = 9
	MdbParserID          = 10
	MdbParserDIGIT       = 11
	MdbParserSTRING      = 12
	MdbParserWS          = 13
)

// MdbParser rules.
const (
	MdbParserRULE_start        = 0
	MdbParserRULE_searchclause = 1
	MdbParserRULE_expr         = 2
	MdbParserRULE_lparen       = 3
	MdbParserRULE_rparen       = 4
	MdbParserRULE_unop         = 5
	MdbParserRULE_relop        = 6
	MdbParserRULE_searchop     = 7
	MdbParserRULE_fieldsearch  = 8
	MdbParserRULE_field        = 9
	MdbParserRULE_table        = 10
	MdbParserRULE_searchterm   = 11
)

// IStartContext is an interface to support dynamic dispatch.
type IStartContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsStartContext differentiates from other interfaces.
	IsStartContext()
}

type StartContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyStartContext() *StartContext {
	var p = new(StartContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = MdbParserRULE_start
	return p
}

func (*StartContext) IsStartContext() {}

func NewStartContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *StartContext {
	var p = new(StartContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = MdbParserRULE_start

	return p
}

func (s *StartContext) GetParser() antlr.Parser { return s.parser }

func (s *StartContext) Searchclause() ISearchclauseContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISearchclauseContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ISearchclauseContext)
}

func (s *StartContext) EOF() antlr.TerminalNode {
	return s.GetToken(MdbParserEOF, 0)
}

func (s *StartContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *StartContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *StartContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MdbListener); ok {
		listenerT.EnterStart(s)
	}
}

func (s *StartContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MdbListener); ok {
		listenerT.ExitStart(s)
	}
}

func (p *MdbParser) Start() (localctx IStartContext) {
	localctx = NewStartContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, MdbParserRULE_start)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(24)
		p.Searchclause()
	}
	{
		p.SetState(25)
		p.Match(MdbParserEOF)
	}

	return localctx
}

// ISearchclauseContext is an interface to support dynamic dispatch.
type ISearchclauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsSearchclauseContext differentiates from other interfaces.
	IsSearchclauseContext()
}

type SearchclauseContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySearchclauseContext() *SearchclauseContext {
	var p = new(SearchclauseContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = MdbParserRULE_searchclause
	return p
}

func (*SearchclauseContext) IsSearchclauseContext() {}

func NewSearchclauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SearchclauseContext {
	var p = new(SearchclauseContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = MdbParserRULE_searchclause

	return p
}

func (s *SearchclauseContext) GetParser() antlr.Parser { return s.parser }

func (s *SearchclauseContext) Table() ITableContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ITableContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ITableContext)
}

func (s *SearchclauseContext) Expr() IExprContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IExprContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IExprContext)
}

func (s *SearchclauseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SearchclauseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SearchclauseContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MdbListener); ok {
		listenerT.EnterSearchclause(s)
	}
}

func (s *SearchclauseContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MdbListener); ok {
		listenerT.ExitSearchclause(s)
	}
}

func (p *MdbParser) Searchclause() (localctx ISearchclauseContext) {
	localctx = NewSearchclauseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, MdbParserRULE_searchclause)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(27)
		p.Table()
	}
	{
		p.SetState(28)
		p.expr(0)
	}

	return localctx
}

// IExprContext is an interface to support dynamic dispatch.
type IExprContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsExprContext differentiates from other interfaces.
	IsExprContext()
}

type ExprContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyExprContext() *ExprContext {
	var p = new(ExprContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = MdbParserRULE_expr
	return p
}

func (*ExprContext) IsExprContext() {}

func NewExprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ExprContext {
	var p = new(ExprContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = MdbParserRULE_expr

	return p
}

func (s *ExprContext) GetParser() antlr.Parser { return s.parser }

func (s *ExprContext) Fieldsearch() IFieldsearchContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IFieldsearchContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IFieldsearchContext)
}

func (s *ExprContext) Searchop() ISearchopContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISearchopContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ISearchopContext)
}

func (s *ExprContext) Unop() IUnopContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IUnopContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IUnopContext)
}

func (s *ExprContext) AllExpr() []IExprContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IExprContext)(nil)).Elem())
	var tst = make([]IExprContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IExprContext)
		}
	}

	return tst
}

func (s *ExprContext) Expr(i int) IExprContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IExprContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IExprContext)
}

func (s *ExprContext) Lparen() ILparenContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ILparenContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ILparenContext)
}

func (s *ExprContext) Relop() IRelopContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IRelopContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IRelopContext)
}

func (s *ExprContext) Rparen() IRparenContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IRparenContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IRparenContext)
}

func (s *ExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExprContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MdbListener); ok {
		listenerT.EnterExpr(s)
	}
}

func (s *ExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MdbListener); ok {
		listenerT.ExitExpr(s)
	}
}

func (p *MdbParser) Expr() (localctx IExprContext) {
	return p.expr(0)
}

func (p *MdbParser) expr(_p int) (localctx IExprContext) {
	var _parentctx antlr.ParserRuleContext = p.GetParserRuleContext()
	_parentState := p.GetState()
	localctx = NewExprContext(p, p.GetParserRuleContext(), _parentState)
	var _prevctx IExprContext = localctx
	var _ antlr.ParserRuleContext = _prevctx // TODO: To prevent unused variable warning.
	_startState := 4
	p.EnterRecursionRule(localctx, 4, MdbParserRULE_expr, _p)

	defer func() {
		p.UnrollRecursionContexts(_parentctx)
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(44)
	p.GetErrorHandler().Sync(p)

	switch p.GetTokenStream().LA(1) {
	case MdbParserID:
		{
			p.SetState(31)
			p.Fieldsearch()
		}

	case MdbParserNO, MdbParserEVERY:
		{
			p.SetState(32)
			p.Searchop()
		}
		{
			p.SetState(33)
			p.Fieldsearch()
		}

	case MdbParserNOT:
		{
			p.SetState(35)
			p.Unop()
		}
		{
			p.SetState(36)
			p.expr(3)
		}

	case MdbParserT__0:
		{
			p.SetState(38)
			p.Lparen()
		}
		{
			p.SetState(39)
			p.expr(0)
		}
		{
			p.SetState(40)
			p.Relop()
		}
		{
			p.SetState(41)
			p.expr(0)
		}
		{
			p.SetState(42)
			p.Rparen()
		}

	default:
		panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
	}
	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(52)
	p.GetErrorHandler().Sync(p)
	_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 1, p.GetParserRuleContext())

	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			localctx = NewExprContext(p, _parentctx, _parentState)
			p.PushNewRecursionContext(localctx, _startState, MdbParserRULE_expr)
			p.SetState(46)

			if !(p.Precpred(p.GetParserRuleContext(), 2)) {
				panic(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 2)", ""))
			}
			{
				p.SetState(47)
				p.Relop()
			}
			{
				p.SetState(48)
				p.expr(3)
			}

		}
		p.SetState(54)
		p.GetErrorHandler().Sync(p)
		_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 1, p.GetParserRuleContext())
	}

	return localctx
}

// ILparenContext is an interface to support dynamic dispatch.
type ILparenContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsLparenContext differentiates from other interfaces.
	IsLparenContext()
}

type LparenContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLparenContext() *LparenContext {
	var p = new(LparenContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = MdbParserRULE_lparen
	return p
}

func (*LparenContext) IsLparenContext() {}

func NewLparenContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *LparenContext {
	var p = new(LparenContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = MdbParserRULE_lparen

	return p
}

func (s *LparenContext) GetParser() antlr.Parser { return s.parser }
func (s *LparenContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LparenContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *LparenContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MdbListener); ok {
		listenerT.EnterLparen(s)
	}
}

func (s *LparenContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MdbListener); ok {
		listenerT.ExitLparen(s)
	}
}

func (p *MdbParser) Lparen() (localctx ILparenContext) {
	localctx = NewLparenContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, MdbParserRULE_lparen)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(55)
		p.Match(MdbParserT__0)
	}

	return localctx
}

// IRparenContext is an interface to support dynamic dispatch.
type IRparenContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsRparenContext differentiates from other interfaces.
	IsRparenContext()
}

type RparenContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyRparenContext() *RparenContext {
	var p = new(RparenContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = MdbParserRULE_rparen
	return p
}

func (*RparenContext) IsRparenContext() {}

func NewRparenContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *RparenContext {
	var p = new(RparenContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = MdbParserRULE_rparen

	return p
}

func (s *RparenContext) GetParser() antlr.Parser { return s.parser }
func (s *RparenContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *RparenContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *RparenContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MdbListener); ok {
		listenerT.EnterRparen(s)
	}
}

func (s *RparenContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MdbListener); ok {
		listenerT.ExitRparen(s)
	}
}

func (p *MdbParser) Rparen() (localctx IRparenContext) {
	localctx = NewRparenContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, MdbParserRULE_rparen)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(57)
		p.Match(MdbParserT__1)
	}

	return localctx
}

// IUnopContext is an interface to support dynamic dispatch.
type IUnopContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsUnopContext differentiates from other interfaces.
	IsUnopContext()
}

type UnopContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyUnopContext() *UnopContext {
	var p = new(UnopContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = MdbParserRULE_unop
	return p
}

func (*UnopContext) IsUnopContext() {}

func NewUnopContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *UnopContext {
	var p = new(UnopContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = MdbParserRULE_unop

	return p
}

func (s *UnopContext) GetParser() antlr.Parser { return s.parser }

func (s *UnopContext) NOT() antlr.TerminalNode {
	return s.GetToken(MdbParserNOT, 0)
}

func (s *UnopContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *UnopContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *UnopContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MdbListener); ok {
		listenerT.EnterUnop(s)
	}
}

func (s *UnopContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MdbListener); ok {
		listenerT.ExitUnop(s)
	}
}

func (p *MdbParser) Unop() (localctx IUnopContext) {
	localctx = NewUnopContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 10, MdbParserRULE_unop)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(59)
		p.Match(MdbParserNOT)
	}

	return localctx
}

// IRelopContext is an interface to support dynamic dispatch.
type IRelopContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsRelopContext differentiates from other interfaces.
	IsRelopContext()
}

type RelopContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyRelopContext() *RelopContext {
	var p = new(RelopContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = MdbParserRULE_relop
	return p
}

func (*RelopContext) IsRelopContext() {}

func NewRelopContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *RelopContext {
	var p = new(RelopContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = MdbParserRULE_relop

	return p
}

func (s *RelopContext) GetParser() antlr.Parser { return s.parser }

func (s *RelopContext) AND() antlr.TerminalNode {
	return s.GetToken(MdbParserAND, 0)
}

func (s *RelopContext) OR() antlr.TerminalNode {
	return s.GetToken(MdbParserOR, 0)
}

func (s *RelopContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *RelopContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *RelopContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MdbListener); ok {
		listenerT.EnterRelop(s)
	}
}

func (s *RelopContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MdbListener); ok {
		listenerT.ExitRelop(s)
	}
}

func (p *MdbParser) Relop() (localctx IRelopContext) {
	localctx = NewRelopContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 12, MdbParserRULE_relop)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(61)
		_la = p.GetTokenStream().LA(1)

		if !(_la == MdbParserAND || _la == MdbParserOR) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

	return localctx
}

// ISearchopContext is an interface to support dynamic dispatch.
type ISearchopContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsSearchopContext differentiates from other interfaces.
	IsSearchopContext()
}

type SearchopContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySearchopContext() *SearchopContext {
	var p = new(SearchopContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = MdbParserRULE_searchop
	return p
}

func (*SearchopContext) IsSearchopContext() {}

func NewSearchopContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SearchopContext {
	var p = new(SearchopContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = MdbParserRULE_searchop

	return p
}

func (s *SearchopContext) GetParser() antlr.Parser { return s.parser }

func (s *SearchopContext) NO() antlr.TerminalNode {
	return s.GetToken(MdbParserNO, 0)
}

func (s *SearchopContext) EVERY() antlr.TerminalNode {
	return s.GetToken(MdbParserEVERY, 0)
}

func (s *SearchopContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SearchopContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SearchopContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MdbListener); ok {
		listenerT.EnterSearchop(s)
	}
}

func (s *SearchopContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MdbListener); ok {
		listenerT.ExitSearchop(s)
	}
}

func (p *MdbParser) Searchop() (localctx ISearchopContext) {
	localctx = NewSearchopContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 14, MdbParserRULE_searchop)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(63)
		_la = p.GetTokenStream().LA(1)

		if !(_la == MdbParserNO || _la == MdbParserEVERY) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

	return localctx
}

// IFieldsearchContext is an interface to support dynamic dispatch.
type IFieldsearchContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsFieldsearchContext differentiates from other interfaces.
	IsFieldsearchContext()
}

type FieldsearchContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFieldsearchContext() *FieldsearchContext {
	var p = new(FieldsearchContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = MdbParserRULE_fieldsearch
	return p
}

func (*FieldsearchContext) IsFieldsearchContext() {}

func NewFieldsearchContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *FieldsearchContext {
	var p = new(FieldsearchContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = MdbParserRULE_fieldsearch

	return p
}

func (s *FieldsearchContext) GetParser() antlr.Parser { return s.parser }

func (s *FieldsearchContext) Field() IFieldContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IFieldContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IFieldContext)
}

func (s *FieldsearchContext) EQ() antlr.TerminalNode {
	return s.GetToken(MdbParserEQ, 0)
}

func (s *FieldsearchContext) Searchterm() ISearchtermContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISearchtermContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ISearchtermContext)
}

func (s *FieldsearchContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FieldsearchContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *FieldsearchContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MdbListener); ok {
		listenerT.EnterFieldsearch(s)
	}
}

func (s *FieldsearchContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MdbListener); ok {
		listenerT.ExitFieldsearch(s)
	}
}

func (p *MdbParser) Fieldsearch() (localctx IFieldsearchContext) {
	localctx = NewFieldsearchContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 16, MdbParserRULE_fieldsearch)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(65)
		p.Field()
	}
	{
		p.SetState(66)
		p.Match(MdbParserEQ)
	}
	{
		p.SetState(67)
		p.Searchterm()
	}

	return localctx
}

// IFieldContext is an interface to support dynamic dispatch.
type IFieldContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsFieldContext differentiates from other interfaces.
	IsFieldContext()
}

type FieldContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFieldContext() *FieldContext {
	var p = new(FieldContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = MdbParserRULE_field
	return p
}

func (*FieldContext) IsFieldContext() {}

func NewFieldContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *FieldContext {
	var p = new(FieldContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = MdbParserRULE_field

	return p
}

func (s *FieldContext) GetParser() antlr.Parser { return s.parser }

func (s *FieldContext) ID() antlr.TerminalNode {
	return s.GetToken(MdbParserID, 0)
}

func (s *FieldContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FieldContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *FieldContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MdbListener); ok {
		listenerT.EnterField(s)
	}
}

func (s *FieldContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MdbListener); ok {
		listenerT.ExitField(s)
	}
}

func (p *MdbParser) Field() (localctx IFieldContext) {
	localctx = NewFieldContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 18, MdbParserRULE_field)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(69)
		p.Match(MdbParserID)
	}

	return localctx
}

// ITableContext is an interface to support dynamic dispatch.
type ITableContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsTableContext differentiates from other interfaces.
	IsTableContext()
}

type TableContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTableContext() *TableContext {
	var p = new(TableContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = MdbParserRULE_table
	return p
}

func (*TableContext) IsTableContext() {}

func NewTableContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TableContext {
	var p = new(TableContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = MdbParserRULE_table

	return p
}

func (s *TableContext) GetParser() antlr.Parser { return s.parser }

func (s *TableContext) ID() antlr.TerminalNode {
	return s.GetToken(MdbParserID, 0)
}

func (s *TableContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TableContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TableContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MdbListener); ok {
		listenerT.EnterTable(s)
	}
}

func (s *TableContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MdbListener); ok {
		listenerT.ExitTable(s)
	}
}

func (p *MdbParser) Table() (localctx ITableContext) {
	localctx = NewTableContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 20, MdbParserRULE_table)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(71)
		p.Match(MdbParserID)
	}

	return localctx
}

// ISearchtermContext is an interface to support dynamic dispatch.
type ISearchtermContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsSearchtermContext differentiates from other interfaces.
	IsSearchtermContext()
}

type SearchtermContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySearchtermContext() *SearchtermContext {
	var p = new(SearchtermContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = MdbParserRULE_searchterm
	return p
}

func (*SearchtermContext) IsSearchtermContext() {}

func NewSearchtermContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SearchtermContext {
	var p = new(SearchtermContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = MdbParserRULE_searchterm

	return p
}

func (s *SearchtermContext) GetParser() antlr.Parser { return s.parser }

func (s *SearchtermContext) AllDIGIT() []antlr.TerminalNode {
	return s.GetTokens(MdbParserDIGIT)
}

func (s *SearchtermContext) DIGIT(i int) antlr.TerminalNode {
	return s.GetToken(MdbParserDIGIT, i)
}

func (s *SearchtermContext) ID() antlr.TerminalNode {
	return s.GetToken(MdbParserID, 0)
}

func (s *SearchtermContext) STRING() antlr.TerminalNode {
	return s.GetToken(MdbParserSTRING, 0)
}

func (s *SearchtermContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SearchtermContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SearchtermContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MdbListener); ok {
		listenerT.EnterSearchterm(s)
	}
}

func (s *SearchtermContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(MdbListener); ok {
		listenerT.ExitSearchterm(s)
	}
}

func (p *MdbParser) Searchterm() (localctx ISearchtermContext) {
	localctx = NewSearchtermContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 22, MdbParserRULE_searchterm)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	var _alt int

	p.SetState(80)
	p.GetErrorHandler().Sync(p)

	switch p.GetTokenStream().LA(1) {
	case MdbParserDIGIT:
		p.EnterOuterAlt(localctx, 1)
		p.SetState(74)
		p.GetErrorHandler().Sync(p)
		_alt = 1
		for ok := true; ok; ok = _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
			switch _alt {
			case 1:
				{
					p.SetState(73)
					p.Match(MdbParserDIGIT)
				}

			default:
				panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			}

			p.SetState(76)
			p.GetErrorHandler().Sync(p)
			_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 2, p.GetParserRuleContext())
		}

	case MdbParserID:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(78)
			p.Match(MdbParserID)
		}

	case MdbParserSTRING:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(79)
			p.Match(MdbParserSTRING)
		}

	default:
		panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
	}

	return localctx
}

func (p *MdbParser) Sempred(localctx antlr.RuleContext, ruleIndex, predIndex int) bool {
	switch ruleIndex {
	case 2:
		var t *ExprContext = nil
		if localctx != nil {
			t = localctx.(*ExprContext)
		}
		return p.Expr_Sempred(t, predIndex)

	default:
		panic("No predicate with index: " + fmt.Sprint(ruleIndex))
	}
}

func (p *MdbParser) Expr_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 0:
		return p.Precpred(p.GetParserRuleContext(), 2)

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}
