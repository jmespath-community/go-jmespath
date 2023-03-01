package api

import (
	"strconv"

	"github.com/jmespath-community/go-jmespath/pkg/interpreter"
	"github.com/jmespath-community/go-jmespath/pkg/parsing"
)

// JMESPath is the representation of a compiled JMES path query. A JMESPath is
// safe for concurrent use by multiple goroutines.
type JMESPath interface {
	Search(interface{}) (interface{}, error)
}

type jmesPath struct {
	ast parsing.ASTNode
}

// Compile parses a JMESPath expression and returns, if successful, a JMESPath
// object that can be used to match against data.
func Compile(expression string) (JMESPath, error) {
	parser := parsing.NewParser()
	ast, err := parser.Parse(expression)
	if err != nil {
		return nil, err
	}
	return jmesPath{ast: ast}, nil
}

// MustCompile is like Compile but panics if the expression cannot be parsed.
// It simplifies safe initialization of global variables holding compiled
// JMESPaths.
func MustCompile(expression string) JMESPath {
	jmespath, err := Compile(expression)
	if err != nil {
		panic(`jmespath: Compile(` + strconv.Quote(expression) + `): ` + err.Error())
	}
	return jmespath
}

// Search evaluates a JMESPath expression against input data and returns the result.
func (jp jmesPath) Search(data interface{}) (interface{}, error) {
	intr := interpreter.NewInterpreter(data)
	return intr.Execute(jp.ast, data)
}

// Search evaluates a JMESPath expression against input data and returns the result.
func Search(expression string, data interface{}) (interface{}, error) {
	if compiled, err := Compile(expression); err != nil {
		return nil, err
	} else {
		return compiled.Search(data)
	}
}
