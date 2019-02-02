/* ANTLR Grammar for Minidb Query Language */

parser grammar MdbParser;

options {
    tokenVocab=MdbLexer;
}

start
    : searchclause EOF
    ;

searchclause
    : table expr
    ;

expr
    : fieldsearch
    | searchop fieldsearch
    | unop expr
    | expr relop expr
    | lparen expr relop expr rparen
    ;

lparen
    : '('
    ;

rparen
    : ')'
    ;

unop
    : NOT
    ;

relop
    : AND
    | OR
    ;

searchop
    : NO
    | EVERY
    ;

fieldsearch
    : field EQ searchterm
    ;

field
    : ID
    ;

table
    : ID
    ;

searchterm
    : 
    | STRING
    | ID+
    | DIGIT+
    | DIGIT+ ID+ 
    ;


