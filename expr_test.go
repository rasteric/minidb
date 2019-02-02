package minidb

import "testing"

func TestParse(t *testing.T) {
	tables := []struct {
		in  string
		out string
	}{
		{"Person Name=John",
			`SearchClause("Person",[InfixOP("=",[FieldString("Name",[]),QueryString("John",[])])])`},
		{"Person Name=%&@John%",
			`SearchClause("Person",[InfixOP("=",[FieldString("Name",[]),QueryString("%&@John%",[])])])`},
		// every, no, not
		{"Person no Name=John",
			`SearchClause("Person",[NoTerm("no",[InfixOP("=",[FieldString("Name",[]),QueryString("John",[])])])])`},
		{"Person every Name=%r%",
			`SearchClause("Person",[EveryTerm("every",[InfixOP("=",[FieldString("Name",[]),QueryString("%r%",[])])])])`},
		{"Person not Name=John",
			`SearchClause("Person",[LogicalNot("not",[InfixOP("=",[FieldString("Name",[]),QueryString("John",[])])])])`},
		{"Person not every Name=John", `SearchClause("Person",[LogicalNot("not",[EveryTerm("every",[InfixOP("=",[FieldString("Name",[]),QueryString("John",[])])])])])`},
		// connectives
		{"Person Name=John and Name=Smith", `SearchClause("Person",[LogicalAnd("and",[InfixOP("=",[FieldString("Name",[]),QueryString("John",[])]),InfixOP("=",[FieldString("Name",[]),QueryString("Smith",[])])])])`},
		{"Person Name=John and Name=Smith or Name=Mueller", `SearchClause("Person",[LogicalOr("or",[LogicalAnd("and",[InfixOP("=",[FieldString("Name",[]),QueryString("John",[])]),InfixOP("=",[FieldString("Name",[]),QueryString("Smith",[])])]),InfixOP("=",[FieldString("Name",[]),QueryString("Mueller",[])])])])`},
		{"Person (Name=John or Name=Bob)",
			`SearchClause("Person",[LogicalOr("or",[InfixOP("=",[FieldString("Name",[]),QueryString("John",[])]),InfixOP("=",[FieldString("Name",[]),QueryString("Bob",[])])])])`},
		{"Person ((Name=John or Name=Bob) and Name=Theodore)",
			`SearchClause("Person",[LogicalAnd("and",[LogicalOr("or",[InfixOP("=",[FieldString("Name",[]),QueryString("John",[])]),InfixOP("=",[FieldString("Name",[]),QueryString("Bob",[])])]),InfixOP("=",[FieldString("Name",[]),QueryString("Theodore",[])])])])`},
	}
	for _, table := range tables {
		result, err := ParseQuery(table.in)
		if err != nil {
			t.Errorf(`Parse("%s") error: %s`, table.in, err)
		} else {
			s := result.DebugDump()
			//	t.Log(s)
			if s != table.out {
				t.Error(`TestParse("` + table.in + `") failed, expected >>` + table.out + `<<, given >>` + s + `<<`)
			}
		}
	}
}
