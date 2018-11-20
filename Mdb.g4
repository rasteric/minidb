/* ANTLR Grammar for Minidb Query Language */

grammar Mdb;

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
    : DIGIT+
    | ID
    | STRING
    ;

AND
    : 'and'
    ;

OR
    : 'or'
    ;

NOT
    : 'not'
    ;
NO
    : 'no'
    ;

EVERY
    : 'every'
    ;

EQ
    : '='
    ;

fragment VALID_ID_START
    : ('a' .. 'z') | ('A' .. 'Z') | '_'
    ;

fragment VALID_ID_CHAR
    : VALID_ID_START | ('0' .. '9')
    ;

NOT_SPECIAL
    : ~(' ' | '\t' | '\n' | '\r' | '\'' | '"' | ';' | '.' | '=' | '(' | ')')
    ;

ID
    : VALID_ID_START VALID_ID_CHAR*
    ;

DIGIT
    : ('0' .. '9')
    ;

STRING
    : NOT_SPECIAL+
    | '"' ~('\n'|'"')* ('"' )
  /*  | { panic("syntax-error - unterminated string literal") } ) */
    ;

WS
   : [ \r\n\t] + -> skip
;
