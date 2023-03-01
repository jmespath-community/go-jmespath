package interpreter

import (
	"errors"
	"fmt"

	jperror "github.com/jmespath-community/go-jmespath/pkg/error"
	"github.com/jmespath-community/go-jmespath/pkg/functions"
	"github.com/jmespath-community/go-jmespath/pkg/parsing"
	"github.com/jmespath-community/go-jmespath/pkg/util"
)

type functionCaller struct {
	functionTable map[string]functions.FunctionEntry
}

func newFunctionCaller(funcs ...functions.FunctionEntry) *functionCaller {
	fTable := map[string]functions.FunctionEntry{}
	for _, f := range funcs {
		fTable[f.Name] = f
	}
	return &functionCaller{
		functionTable: fTable,
	}
}

func resolveArgs(e functions.FunctionEntry, arguments []interface{}) ([]interface{}, error) {
	if len(e.Arguments) == 0 {
		return arguments, nil
	}

	variadic := isVariadic(e.Arguments)
	minExpected := getMinExpected(e.Arguments)
	maxExpected, hasMax := getMaxExpected(e.Arguments)
	count := len(arguments)

	if count < minExpected {
		return nil, jperror.NotEnoughArgumentsSupplied(e.Name, count, minExpected, variadic)
	}

	if hasMax && count > maxExpected {
		return nil, jperror.TooManyArgumentsSupplied(e.Name, count, maxExpected)
	}

	for i, spec := range e.Arguments {
		if !spec.Optional || i <= len(arguments)-1 {
			userArg := arguments[i]
			err := typeCheck(spec, userArg)
			if err != nil {
				return nil, err
			}
		}
	}
	lastIndex := len(e.Arguments) - 1
	lastArg := e.Arguments[lastIndex]
	if lastArg.Variadic {
		for i := len(e.Arguments) - 1; i < len(arguments); i++ {
			userArg := arguments[i]
			err := typeCheck(lastArg, userArg)
			if err != nil {
				return nil, err
			}
		}
	}
	return arguments, nil
}

func isVariadic(arguments []functions.ArgSpec) bool {
	for _, spec := range arguments {
		if spec.Variadic {
			return true
		}
	}
	return false
}

func getMinExpected(arguments []functions.ArgSpec) int {
	expected := 0
	for _, spec := range arguments {
		if !spec.Optional {
			expected++
		}
	}
	return expected
}

func getMaxExpected(arguments []functions.ArgSpec) (int, bool) {
	if isVariadic(arguments) {
		return 0, false
	}
	return len(arguments), true
}

func typeCheck(a functions.ArgSpec, arg interface{}) error {
	for _, t := range a.Types {
		switch t {
		case functions.JpNumber:
			if _, ok := arg.(float64); ok {
				return nil
			}
		case functions.JpString:
			if _, ok := arg.(string); ok {
				return nil
			}
		case functions.JpArray:
			if util.IsSliceType(arg) {
				return nil
			}
		case functions.JpObject:
			if _, ok := arg.(map[string]interface{}); ok {
				return nil
			}
		case functions.JpArrayArray:
			if util.IsSliceType(arg) {
				if _, ok := arg.([]interface{}); ok {
					return nil
				}
			}
		case functions.JpArrayNumber:
			if _, ok := util.ToArrayNum(arg); ok {
				return nil
			}
		case functions.JpArrayString:
			if _, ok := util.ToArrayStr(arg); ok {
				return nil
			}
		case functions.JpAny:
			return nil
		case functions.JpExpref:
			if _, ok := arg.(functions.ExpRef); ok {
				return nil
			}
		}
	}
	return fmt.Errorf("invalid type for: %v, expected: %#v", arg, a.Types)
}

func (f *functionCaller) CallFunction(name string, arguments []interface{}, intr Interpreter) (interface{}, error) {
	entry, ok := f.functionTable[name]
	if !ok {
		return nil, errors.New("unknown function: " + name)
	}
	resolvedArgs, err := resolveArgs(entry, arguments)
	if err != nil {
		return nil, err
	}
	exec := func(node parsing.ASTNode, data interface{}, scope map[string]interface{}) (interface{}, error) {
		intr := intr
		if scope != nil {
			intr = intr.WithScope(scope)
		}
		return intr.Execute(node, data)
	}
	return entry.Handler(exec, resolvedArgs)
}
