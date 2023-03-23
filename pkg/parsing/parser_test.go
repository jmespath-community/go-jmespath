package parsing

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsingVariable(t *testing.T) {
	assert := assert.New(t)
	tokens := []token{
		{tokenType: TOKVarref, value: "foo", position: 20, length: 3},
		{tokenType: TOKEOF, position: 19},
	}

	var prettyPrintedLookup = `ASTVariable {
  value: "foo"
}
`
	p := NewParser()
	parsed, _ := p.parseTokens(tokens)
	assert.Equal(prettyPrintedLookup, parsed.PrettyPrint(0))
}

func TestParsingVariableBinding(t *testing.T) {
	assert := assert.New(t)
	tokens := []token{
		{tokenType: TOKVarref, value: "foo", position: 4, length: 4},
		{tokenType: TOKAssign, value: "=", position: 9, length: 1},
		{tokenType: TOKUnquotedIdentifier, value: "foo", position: 11, length: 3},
		{tokenType: TOKEOF, position: 19},
	}

	var prettyPrintedLookup = `ASTBinding {
  children: {
    ASTVariable {
      value: "foo"
    }
    ASTField {
      value: "foo"
    }
  }
}
`
	p := NewParser()
	parsed, _ := p.parseTokens(tokens)
	assert.Equal(prettyPrintedLookup, parsed.PrettyPrint(0))
}

func TestParsingLetExpression(t *testing.T) {
	// let $foo = foo in @
	// 012345678901234567890123
	//           1         2
	assert := assert.New(t)
	tokens := []token{
		{tokenType: TOKUnquotedIdentifier, value: "let", position: 0, length: 3},
		{tokenType: TOKVarref, value: "foo", position: 4, length: 4},
		{tokenType: TOKAssign, value: "=", position: 9, length: 1},
		{tokenType: TOKUnquotedIdentifier, value: "foo", position: 11, length: 3},
		{tokenType: TOKUnquotedIdentifier, value: "in", position: 15, length: 2},
		{tokenType: TOKCurrent, value: "@", position: 18, length: 1},
		{tokenType: TOKEOF, position: 19},
	}

	expected := `ASTLetExpression {
  children: {
    ASTBindings {
      children: {
        ASTBinding {
          children: {
            ASTVariable {
              value: "foo"
            }
            ASTField {
              value: "foo"
            }
          }
        }
      }
    }
    ASTCurrentNode {
    }
  }
}
`
	p := NewParser()
	parsed, _ := p.parseTokens(tokens)
	assert.Equal(expected, parsed.PrettyPrint(0))
}

var parsingErrorTests = []struct {
	expression string
	msg        string
}{
	{"foo.", "Incopmlete expression"},
	{"[foo", "Incopmlete expression"},
	{"]", "Invalid"},
	{")", "Invalid"},
	{"}", "Invalid"},
	{"foo..bar", "Invalid"},
	{`foo."bar`, "Forwards lexer errors"},
	{`{foo: bar`, "Incomplete expression"},
	{`{foo bar}`, "Invalid"},
	{`[foo bar]`, "Invalid"},
	{`foo@`, "Invalid"},
	{`&&&&&&&&&&&&t(`, "Invalid"},
	{`[*][`, "Invalid"},
}

func TestParsingErrors(t *testing.T) {
	assert := assert.New(t)
	parser := NewParser()
	for _, tt := range parsingErrorTests {
		_, err := parser.Parse(tt.expression)
		assert.NotNil(err, fmt.Sprintf("Expected parsing error: %s, for expression: %s", tt.msg, tt.expression))
	}
}

var prettyPrinted = `ASTProjection {
  children: {
    ASTField {
      value: "foo"
    }
    ASTSubexpression {
      children: {
        ASTSubexpression {
          children: {
            ASTField {
              value: "bar"
            }
            ASTField {
              value: "baz"
            }
          }
        }
        ASTField {
          value: "qux"
        }
      }
    }
  }
}
`

var prettyPrintedCompNode = `ASTFilterProjection {
  children: {
    ASTField {
      value: "a"
    }
    ASTIdentity {
    }
    ASTComparator {
      value: TOKLTE
      children: {
        ASTField {
          value: "b"
        }
        ASTField {
          value: "c"
        }
      }
    }
  }
}
`

func TestPrettyPrintedAST(t *testing.T) {
	assert := assert.New(t)
	parser := NewParser()
	parsed, _ := parser.Parse("foo[*].bar.baz.qux")
	assert.Equal(parsed.PrettyPrint(0), prettyPrinted)
}

func TestPrettyPrintedCompNode(t *testing.T) {
	assert := assert.New(t)
	parser := NewParser()
	parsed, _ := parser.Parse("a[?b<=c]")
	assert.Equal(parsed.PrettyPrint(0), prettyPrintedCompNode)
}

func BenchmarkParseIdentifier(b *testing.B) {
	runParseBenchmark(b, exprIdentifier)
}

func BenchmarkParseSubexpression(b *testing.B) {
	runParseBenchmark(b, exprSubexpr)
}

func BenchmarkParseDeeplyNested50(b *testing.B) {
	runParseBenchmark(b, deeplyNested50)
}

func BenchmarkParseDeepNested50Pipe(b *testing.B) {
	runParseBenchmark(b, deeplyNested50Pipe)
}

func BenchmarkParseDeepNested50Index(b *testing.B) {
	runParseBenchmark(b, deeplyNested50Index)
}

func BenchmarkParseQuotedIdentifier(b *testing.B) {
	runParseBenchmark(b, exprQuotedIdentifier)
}

func BenchmarkParseQuotedIdentifierEscapes(b *testing.B) {
	runParseBenchmark(b, quotedIdentifierEscapes)
}

func BenchmarkParseRawStringLiteral(b *testing.B) {
	runParseBenchmark(b, rawStringLiteral)
}

func BenchmarkParseDeepProjection104(b *testing.B) {
	runParseBenchmark(b, deepProjection104)
}

func runParseBenchmark(b *testing.B, expression string) {
	b.Helper()
	assert := assert.New(b)
	parser := NewParser()
	for i := 0; i < b.N; i++ {
		_, err := parser.Parse(expression)
		if err != nil {
			assert.Fail("Failed to parse expression")
		}
	}
}
