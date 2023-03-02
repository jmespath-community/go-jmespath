package api

import (
	"strconv"

	"github.com/jmespath-community/go-jmespath/pkg/functions"
	"github.com/jmespath-community/go-jmespath/pkg/interpreter"
	"github.com/jmespath-community/go-jmespath/pkg/parsing"
)

// JMESPath is the representation of a compiled JMES path query. A JMESPath is
// safe for concurrent use by multiple goroutines.
type JMESPath interface {
	Search(interface{}) (interface{}, error)
}

type jmesPath struct {
	node  parsing.ASTNode
	funcs []functions.FunctionEntry
}

func newJMESPath(node parsing.ASTNode, funcs ...functions.FunctionEntry) JMESPath {
	return jmesPath{
		node:  node,
		funcs: funcs,
	}
}

// Compile parses a JMESPath expression and returns, if successful, a JMESPath
// object that can be used to match against data.
func Compile(expression string) (JMESPath, error) {
	parser := parsing.NewParser()
	ast, err := parser.Parse(expression)
	if err != nil {
		return nil, err
	}
	return newJMESPath(ast, functions.GetDefaultFunctions()...), nil
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
	intr := interpreter.NewInterpreter(data, jp.funcs...)
	return intr.Execute(jp.node, data)
}

// Search evaluates a JMESPath expression against input data and returns the result.
func Search(expression string, data interface{}) (interface{}, error) {
	compiled, err := Compile(expression)
	if err != nil {
		return nil, err
	}
	return compiled.Search(data)
}
