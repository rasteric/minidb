package minidb

import (
	"fmt"
	"regexp"
)

type token struct {
	sort    QuerySort
	content []rune
}

type stack struct {
	data []token
	// name string
}

type queue struct {
	data []token
}

func newStack() *stack {
	s := stack{data: []token{}}
	return &s
}

func newQueue() *queue {
	q := queue{data: []token{}}
	return &q
}

func (s *stack) push(v token) {
	// fmt.Printf("*** %s push %s: %s\n", s.name, QuerySortToStr(v.sort), string(v.content))
	s.data = append(s.data, v)
}

func (s *stack) isEmpty() bool {
	return len(s.data) == 0
}

func (s *stack) pop() token {
	v := s.data[len(s.data)-1]
	s.data = s.data[:len(s.data)-1]
	// fmt.Printf("*** %s pop %s: %s\n", s.name, QuerySortToStr(v.sort), string(v.content))
	return v
}

func (s *stack) peek() token {
	return s.data[len(s.data)-1]
}

func (s *stack) count() int {
	return len(s.data)
}

func (s *stack) debugDump() string {
	r := "Stack dump:\n"
	for i := len(s.data) - 1; i >= 0; i-- {
		r = r + fmt.Sprintf("Sort %s, Content '%s'\n", QuerySortToStr(s.data[i].sort), string(s.data[i].content))
	}
	return r
}

func (q *queue) add(v token) {
	q.data = append(q.data, v)
}

func (q *queue) isEmpty() bool {
	return len(q.data) == 0
}

func (q *queue) get() token {
	v := q.data[0]
	q.data = q.data[1:]
	return v
}

func (q *queue) count() int {
	return len(q.data)
}

type pstate struct {
	pos int
	in  []rune
	out *stack
	ops *stack
}

func newPstate(s string) *pstate {
	v := pstate{0, []rune(s), newStack(), newStack()}
	return &v
}

// start the parsing (like a start rule)
func start(state *pstate) error {
	return parseSearchClause(state)
}

// consume one literal token or just lookahead
func lookOrConsume1(state *pstate, consume bool) []rune {
	var isTokenChar = regexp.MustCompile(`^[a-zA-Z_0-9]+$`).MatchString
	i := state.pos
	for i < len(state.in) && state.in[i] == ' ' {
		i++
	}
	from := i
	for i < len(state.in) && isTokenChar(string(state.in[i])) {
		i++
	}
	if consume {
		state.pos = i
	}
	return state.in[from:i]
}

func lookAhead1(state *pstate) []rune {
	return lookOrConsume1(state, false)
}

func consume1(state *pstate) []rune {
	return lookOrConsume1(state, true)
}

// parse parts like "Name=query" or "not every Name=John" or
// "no Data=hello%"
func parseSearch(state *pstate) error {
	peek := lookAhead1(state)
	if string(peek) == "not" {
		state.ops.push(token{content: peek, sort: LogicalNot})
		consume1(state)
		peek = lookAhead1(state)
	}
	switch string(peek) {
	case "every":
		state.ops.push(token{content: peek, sort: EveryTerm})
		consume1(state)
	case "no":
		state.ops.push(token{content: peek, sort: NoTerm})
		consume1(state)
	}
	if err := parseField(state); err != nil {
		return err
	}
	if err := parseInfixOP(state); err != nil {
		return err
	}
	return parseSearchQuery(state)
}

func maybeParseParens(state *pstate) error {
	maybeParen := true
	for maybeParen && state.pos < len(state.in) {
		skipWS(state)
		nc := state.in[state.pos]
		switch nc {
		case '(':
			state.ops.push(token{content: []rune{nc}, sort: LeftParen})
			state.pos++
		case ')':
			/* while the operator at the top of the operator stack is not a left bracket:
			    pop the operator from the operator stack onto the output queue.
			pop the left bracket from the stack. */
			for !state.ops.isEmpty() && state.ops.peek().sort != LeftParen {
				state.out.push(state.ops.pop())
			}
			if state.ops.isEmpty() {
				return Fail(`pos=%d: mismatched right paren (no left paren)`, state.pos)
			}
			state.ops.pop()
			state.pos++
			maybeParseParens(state)
			maybeParseConnective(state)
		default:
			maybeParen = false
		}
	}
	return nil
}

func maybeParseConnective(state *pstate) error {
	// peek for logical and, or
	skipWS(state)
	peek := lookAhead1(state)
	switch string(peek) {
	case "and", "or":
		consume1(state)
		/* if the token is an operator, then:
		   while ((there is a function at the top of the operator stack)
		          or (there is an operator at the top of the operator stack with greater precedence)
		          or (the operator at the top of the operator stack has equal precedence and is left associative))
		         and (the operator at the top of the operator stack is not a left bracket):
		       pop operators from the operator stack onto the output queue.
		   push it onto the operator stack. */
		for !state.ops.isEmpty() && state.ops.peek().sort != LeftParen && state.ops.peek().sort != SearchClause {
			state.out.push(state.ops.pop())
		}
		if string(peek) == "or" {
			state.ops.push(token{sort: LogicalOr, content: peek})
		} else {
			state.ops.push(token{sort: LogicalAnd, content: peek})
		}
		err := parseComplexExpr(state)
		if err != nil {
			return Fail(`pos=%d: missing argument, expected a search clause like "Name=John" as second argument of "%s"; %s`, state.pos, string(peek), err)
		}
	}
	return nil
}

func parseComplexExpr(state *pstate) error {
	if err := maybeParseParens(state); err != nil {
		return err
	}
	if err := parseSearch(state); err != nil {
		return err
	}
	if err := maybeParseConnective(state); err != nil {
		return err
	}
	if err := maybeParseParens(state); err != nil {
		return err
	}
	return nil
}

// parse "Person Name=query" or "Person not every Name=Query"
func parseSearchClause(state *pstate) error {
	state.ops.push(token{sort: SearchClause})
	if err := parseTable(state); err != nil {
		return err
	}
	parseComplexExpr(state)
	return nil
}

// parse the name of a table
func parseTable(state *pstate) error {
	var isTablename = regexp.MustCompile(`^[a-zA-Z_0-9]+$`).MatchString
	skipWS(state)
	start := state.pos
	for state.pos < len(state.in) && isTablename(string(state.in[state.pos])) {
		state.pos++
	}
	if start == state.pos {
		return Fail(`pos=%d: malformed or missing table name`, start)
	}
	state.out.push(token{content: state.in[start:state.pos], sort: TableString})
	return nil
}

// parse the name of a field
func parseField(state *pstate) error {
	var isFieldname = regexp.MustCompile(`^[a-zA-Z_0-9]+$`).MatchString
	skipWS(state)
	start := state.pos
	for state.pos < len(state.in) && isFieldname(string(state.in[state.pos])) {
		state.pos++
	}
	if start == state.pos {
		return Fail(`pos=%d: malformed or missing field name`, start)
	}
	state.out.push(token{content: state.in[start:state.pos], sort: FieldString})
	return nil
}

// parse an infix operator such as "="
func parseInfixOP(state *pstate) error {
	skipWS(state)
	if state.pos >= len(state.in) {
		return Fail(`pos=%d: unexpected end of line`, len(state.in))
	}
	c := state.in[state.pos]
	switch c {
	case '=':
		state.pos++
		state.ops.push(token{content: []rune{'='}, sort: InfixOP})
	default:
		return Fail(`pos=%d: expected an infix operator like "="`, state.pos)
	}
	return nil
}

// parse the query part which may be a string or unquoted, e.g.
// "%hello%" in the query "Person Name=%hello%"
func parseSearchQuery(state *pstate) error {
	var noDelimiter = regexp.MustCompile(`\S`).MatchString
	skipWS(state)
	start := state.pos
	if state.in[state.pos] == '"' {
		return parseString(state)
	}
	for state.pos < len(state.in) && noDelimiter(string(state.in[state.pos])) && state.in[state.pos] != ')' {
		state.pos++
	}
	state.out.push(token{content: state.in[start:state.pos], sort: QueryString})
	return nil
}

// parse a string until the end of the string, adding the unquoted string as token
func parseString(state *pstate) error {
	state.pos++
	start := state.pos
	if state.pos < len(state.in) && !(state.in[state.pos] == '"') {
		state.pos++
	}
	state.out.push(token{content: state.in[start:state.pos], sort: SearchClause})
	state.pos++
	return nil
}

func skipWS(state *pstate) {
	for state.pos < len(state.in) && state.in[state.pos] == ' ' {
		state.pos++
	}
}

// convert the operator stack and argument queue to a well-formed Query structure
// with nested arguments - it's a simple AST.
func convert(parse *pstate) (*Query, error) {
	// if we're here, we're expecting something in the output
	if parse.out.isEmpty() {
		return nil, Fail(`syntax error, unexpected end of line`)
	}
	token := parse.out.pop()
	switch token.sort {

	case LogicalAnd, LogicalOr:
		rarg, err := convert(parse)
		if err != nil {
			return nil, err
		}
		larg, err := convert(parse)
		if err != nil {
			return nil, err
		}
		query := Query{Sort: token.sort, Data: string(token.content), Children: []Query{*larg, *rarg}}
		return &query, nil

	case EveryTerm, NoTerm:
		embeddedQuery, err := convert(parse)
		if err != nil {
			return nil, err
		}
		query := Query{Sort: token.sort, Data: string(token.content), Children: []Query{*embeddedQuery}}
		return &query, nil

	case LogicalNot:
		embeddedQuery, err := convert(parse)
		if err != nil {
			return nil, err
		}
		query := Query{Sort: LogicalNot, Data: string(token.content), Children: []Query{*embeddedQuery}}
		return &query, nil

	case SearchClause:
		subclause, err := convert(parse)
		if err != nil {
			return nil, Fail(`invalid or missing search expression: %s`, err)
		}
		table, err := convert(parse)
		if err != nil {
			return nil, Fail(`invalid or missing table name: %s`, err)
		}
		query := Query{Sort: SearchClause, Children: []Query{*subclause},
			Data: table.Data}
		return &query, nil

	case InfixOP:
		if parse.out.count() < 2 {
			return nil, Fail(`syntax error, incomplete query, expected form fieldname=query`)
		}
		query1 := Query{Sort: QueryString, Data: string(parse.out.pop().content)}
		query2 := Query{Sort: FieldString, Data: string(parse.out.pop().content)}
		query3 := Query{Sort: InfixOP, Data: string(token.content), Children: []Query{query2, query1}}
		return &query3, nil

	case FieldString:
		query := Query{Sort: FieldString, Data: string(token.content)}
		return &query, nil

	case QueryString:
		query := Query{Sort: QueryString, Data: string(token.content)}
		return &query, nil

	case TableString:
		query := Query{Sort: TableString, Data: string(token.content)}
		return &query, nil

	default:
		return nil, Fail(`syntax error, unexpected operator type %s`, QuerySortToStr(token.sort))
	}
}

// ParseQuery parses a string representing the part after the table name into a Query structure.
// Implements the classic Shunting Yard algorithm.
func ParseQuery(s string) (*Query, error) {
	state := newPstate(s)

	// fmt.Println()
	// fmt.Println(s)

	if err := start(state); err != nil {
		return nil, err
	}
	// push all remaining operators to the output stack
	for !state.ops.isEmpty() {
		op := state.ops.pop()
		if op.sort == LeftParen {
			return nil, Fail(`syntax error, unmatched left parenthesis`)
		}
		state.out.push(op)
	}
	// fmt.Print(state.out.debugDump())
	query, err := convert(state)
	if err != nil {
		return nil, err
	}
	// fmt.Print(query.DebugDump())
	if !state.ops.isEmpty() || !state.out.isEmpty() {
		return nil, Fail(`syntax error, unexpected input at end of the line\n++OutputStack=%s,\n++Operator stack=%s\n++Partial parse=%s`, state.out.debugDump(), state.ops.debugDump(), query.DebugDump())
	}
	return query, nil
}
