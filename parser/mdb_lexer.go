// Code generated from Mdb.g4 by ANTLR 4.7.1. DO NOT EDIT.

package parser

import (
	"fmt"
	"unicode"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

// Suppress unused import error
var _ = fmt.Printf
var _ = unicode.IsLetter

var serializedLexerAtn = []uint16{
	3, 24715, 42794, 33075, 47597, 16764, 15335, 30598, 22884, 2, 15, 99, 8,
	1, 4, 2, 9, 2, 4, 3, 9, 3, 4, 4, 9, 4, 4, 5, 9, 5, 4, 6, 9, 6, 4, 7, 9,
	7, 4, 8, 9, 8, 4, 9, 9, 9, 4, 10, 9, 10, 4, 11, 9, 11, 4, 12, 9, 12, 4,
	13, 9, 13, 4, 14, 9, 14, 4, 15, 9, 15, 4, 16, 9, 16, 3, 2, 3, 2, 3, 3,
	3, 3, 3, 4, 3, 4, 3, 4, 3, 4, 3, 5, 3, 5, 3, 5, 3, 6, 3, 6, 3, 6, 3, 6,
	3, 7, 3, 7, 3, 7, 3, 8, 3, 8, 3, 8, 3, 8, 3, 8, 3, 8, 3, 9, 3, 9, 3, 10,
	5, 10, 61, 10, 10, 3, 11, 3, 11, 5, 11, 65, 10, 11, 3, 12, 3, 12, 3, 13,
	3, 13, 7, 13, 71, 10, 13, 12, 13, 14, 13, 74, 11, 13, 3, 14, 3, 14, 3,
	15, 6, 15, 79, 10, 15, 13, 15, 14, 15, 80, 3, 15, 3, 15, 7, 15, 85, 10,
	15, 12, 15, 14, 15, 88, 11, 15, 3, 15, 5, 15, 91, 10, 15, 3, 16, 6, 16,
	94, 10, 16, 13, 16, 14, 16, 95, 3, 16, 3, 16, 2, 2, 17, 3, 3, 5, 4, 7,
	5, 9, 6, 11, 7, 13, 8, 15, 9, 17, 10, 19, 2, 21, 2, 23, 11, 25, 12, 27,
	13, 29, 14, 31, 15, 3, 2, 6, 5, 2, 67, 92, 97, 97, 99, 124, 10, 2, 11,
	12, 15, 15, 34, 34, 36, 36, 41, 43, 48, 48, 61, 61, 63, 63, 4, 2, 12, 12,
	36, 36, 5, 2, 11, 12, 15, 15, 34, 34, 2, 102, 2, 3, 3, 2, 2, 2, 2, 5, 3,
	2, 2, 2, 2, 7, 3, 2, 2, 2, 2, 9, 3, 2, 2, 2, 2, 11, 3, 2, 2, 2, 2, 13,
	3, 2, 2, 2, 2, 15, 3, 2, 2, 2, 2, 17, 3, 2, 2, 2, 2, 23, 3, 2, 2, 2, 2,
	25, 3, 2, 2, 2, 2, 27, 3, 2, 2, 2, 2, 29, 3, 2, 2, 2, 2, 31, 3, 2, 2, 2,
	3, 33, 3, 2, 2, 2, 5, 35, 3, 2, 2, 2, 7, 37, 3, 2, 2, 2, 9, 41, 3, 2, 2,
	2, 11, 44, 3, 2, 2, 2, 13, 48, 3, 2, 2, 2, 15, 51, 3, 2, 2, 2, 17, 57,
	3, 2, 2, 2, 19, 60, 3, 2, 2, 2, 21, 64, 3, 2, 2, 2, 23, 66, 3, 2, 2, 2,
	25, 68, 3, 2, 2, 2, 27, 75, 3, 2, 2, 2, 29, 90, 3, 2, 2, 2, 31, 93, 3,
	2, 2, 2, 33, 34, 7, 42, 2, 2, 34, 4, 3, 2, 2, 2, 35, 36, 7, 43, 2, 2, 36,
	6, 3, 2, 2, 2, 37, 38, 7, 99, 2, 2, 38, 39, 7, 112, 2, 2, 39, 40, 7, 102,
	2, 2, 40, 8, 3, 2, 2, 2, 41, 42, 7, 113, 2, 2, 42, 43, 7, 116, 2, 2, 43,
	10, 3, 2, 2, 2, 44, 45, 7, 112, 2, 2, 45, 46, 7, 113, 2, 2, 46, 47, 7,
	118, 2, 2, 47, 12, 3, 2, 2, 2, 48, 49, 7, 112, 2, 2, 49, 50, 7, 113, 2,
	2, 50, 14, 3, 2, 2, 2, 51, 52, 7, 103, 2, 2, 52, 53, 7, 120, 2, 2, 53,
	54, 7, 103, 2, 2, 54, 55, 7, 116, 2, 2, 55, 56, 7, 123, 2, 2, 56, 16, 3,
	2, 2, 2, 57, 58, 7, 63, 2, 2, 58, 18, 3, 2, 2, 2, 59, 61, 9, 2, 2, 2, 60,
	59, 3, 2, 2, 2, 61, 20, 3, 2, 2, 2, 62, 65, 5, 19, 10, 2, 63, 65, 4, 50,
	59, 2, 64, 62, 3, 2, 2, 2, 64, 63, 3, 2, 2, 2, 65, 22, 3, 2, 2, 2, 66,
	67, 10, 3, 2, 2, 67, 24, 3, 2, 2, 2, 68, 72, 5, 19, 10, 2, 69, 71, 5, 21,
	11, 2, 70, 69, 3, 2, 2, 2, 71, 74, 3, 2, 2, 2, 72, 70, 3, 2, 2, 2, 72,
	73, 3, 2, 2, 2, 73, 26, 3, 2, 2, 2, 74, 72, 3, 2, 2, 2, 75, 76, 4, 50,
	59, 2, 76, 28, 3, 2, 2, 2, 77, 79, 5, 23, 12, 2, 78, 77, 3, 2, 2, 2, 79,
	80, 3, 2, 2, 2, 80, 78, 3, 2, 2, 2, 80, 81, 3, 2, 2, 2, 81, 91, 3, 2, 2,
	2, 82, 86, 7, 36, 2, 2, 83, 85, 10, 4, 2, 2, 84, 83, 3, 2, 2, 2, 85, 88,
	3, 2, 2, 2, 86, 84, 3, 2, 2, 2, 86, 87, 3, 2, 2, 2, 87, 89, 3, 2, 2, 2,
	88, 86, 3, 2, 2, 2, 89, 91, 7, 36, 2, 2, 90, 78, 3, 2, 2, 2, 90, 82, 3,
	2, 2, 2, 91, 30, 3, 2, 2, 2, 92, 94, 9, 5, 2, 2, 93, 92, 3, 2, 2, 2, 94,
	95, 3, 2, 2, 2, 95, 93, 3, 2, 2, 2, 95, 96, 3, 2, 2, 2, 96, 97, 3, 2, 2,
	2, 97, 98, 8, 16, 2, 2, 98, 32, 3, 2, 2, 2, 10, 2, 60, 64, 72, 80, 86,
	90, 95, 3, 8, 2, 2,
}

var lexerDeserializer = antlr.NewATNDeserializer(nil)
var lexerAtn = lexerDeserializer.DeserializeFromUInt16(serializedLexerAtn)

var lexerChannelNames = []string{
	"DEFAULT_TOKEN_CHANNEL", "HIDDEN",
}

var lexerModeNames = []string{
	"DEFAULT_MODE",
}

var lexerLiteralNames = []string{
	"", "'('", "')'", "'and'", "'or'", "'not'", "'no'", "'every'", "'='",
}

var lexerSymbolicNames = []string{
	"", "", "", "AND", "OR", "NOT", "NO", "EVERY", "EQ", "NOT_SPECIAL", "ID",
	"DIGIT", "STRING", "WS",
}

var lexerRuleNames = []string{
	"T__0", "T__1", "AND", "OR", "NOT", "NO", "EVERY", "EQ", "VALID_ID_START",
	"VALID_ID_CHAR", "NOT_SPECIAL", "ID", "DIGIT", "STRING", "WS",
}

type MdbLexer struct {
	*antlr.BaseLexer
	channelNames []string
	modeNames    []string
	// TODO: EOF string
}

var lexerDecisionToDFA = make([]*antlr.DFA, len(lexerAtn.DecisionToState))

func init() {
	for index, ds := range lexerAtn.DecisionToState {
		lexerDecisionToDFA[index] = antlr.NewDFA(ds, index)
	}
}

func NewMdbLexer(input antlr.CharStream) *MdbLexer {

	l := new(MdbLexer)

	l.BaseLexer = antlr.NewBaseLexer(input)
	l.Interpreter = antlr.NewLexerATNSimulator(l, lexerAtn, lexerDecisionToDFA, antlr.NewPredictionContextCache())

	l.channelNames = lexerChannelNames
	l.modeNames = lexerModeNames
	l.RuleNames = lexerRuleNames
	l.LiteralNames = lexerLiteralNames
	l.SymbolicNames = lexerSymbolicNames
	l.GrammarFileName = "Mdb.g4"
	// TODO: l.EOF = antlr.TokenEOF

	return l
}

// MdbLexer tokens.
const (
	MdbLexerT__0        = 1
	MdbLexerT__1        = 2
	MdbLexerAND         = 3
	MdbLexerOR          = 4
	MdbLexerNOT         = 5
	MdbLexerNO          = 6
	MdbLexerEVERY       = 7
	MdbLexerEQ          = 8
	MdbLexerNOT_SPECIAL = 9
	MdbLexerID          = 10
	MdbLexerDIGIT       = 11
	MdbLexerSTRING      = 12
	MdbLexerWS          = 13
)
