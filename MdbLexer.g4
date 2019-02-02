lexer grammar MdbLexer;

STRING
 : '"' ~[\r\n"]* '"'
 ;

OPAR
 : '('
 ;

CPAR
 : ')'
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
 : '=' -> pushMode(NOT_SPECIAL_MODE)
 ;

ID
 : VALID_ID_START VALID_ID_CHAR*
 ;

DIGIT
 : [0-9]
 ;

WS
 : [ \r\n\t]+ -> skip
 ;

fragment VALID_ID_START
 : [a-zA-Z_]
 ;

fragment VALID_ID_CHAR
 : [a-zA-Z_0-9]
 ;

mode NOT_SPECIAL_MODE;

  OPAR2
   : '(' -> type(OPAR), popMode
   ;

  CPAR2
   : ')' -> type(CPAR), popMode
   ;

  WS2
   : [ \t\r\n] -> skip, popMode
   ;

  NOT_SPECIAL
   : ~[ \t\r\n()]+
   ;