package interpreter

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	jperror "github.com/jmespath-community/go-jmespath/pkg/error"
	"github.com/jmespath-community/go-jmespath/pkg/parsing"
	"github.com/jmespath-community/go-jmespath/pkg/util"
)

type ExecuteFunc = func(parsing.ASTNode, interface{}, map[string]interface{}) (interface{}, error)

type JpFunction = func(ExecuteFunc, []interface{}) (interface{}, error)

type JpType string

const (
	JpNumber      JpType = "number"
	JpString      JpType = "string"
	JpArray       JpType = "array"
	JpObject      JpType = "object"
	JpArrayArray  JpType = "array[array]"
	JpArrayNumber JpType = "array[number]"
	JpArrayString JpType = "array[string]"
	JpExpref      JpType = "expref"
	JpAny         JpType = "any"
)

type FunctionEntry struct {
	Name      string
	Arguments []ArgSpec
	Handler   JpFunction
}

type ArgSpec struct {
	Types    []JpType
	Variadic bool
	Optional bool
}

type byExprString struct {
	items []interface{}
	keys  []interface{}
}

func (a *byExprString) Len() int {
	return len(a.items)
}

func (a *byExprString) Swap(i, j int) {
	a.items[i], a.items[j] = a.items[j], a.items[i]
	a.keys[i], a.keys[j] = a.keys[j], a.keys[i]
}

func (a *byExprString) Less(i, j int) bool {
	ith := a.keys[i].(string)
	jth := a.keys[j].(string)
	return ith < jth
}

type byExprFloat struct {
	items []interface{}
	keys  []interface{}
}

func (a *byExprFloat) Len() int {
	return len(a.items)
}

func (a *byExprFloat) Swap(i, j int) {
	a.items[i], a.items[j] = a.items[j], a.items[i]
	a.keys[i], a.keys[j] = a.keys[j], a.keys[i]
}

func (a *byExprFloat) Less(i, j int) bool {
	ith := a.keys[i].(float64)
	jth := a.keys[j].(float64)
	return ith < jth
}

type functionCaller struct {
	functionTable map[string]FunctionEntry
}

func newFunctionCaller() *functionCaller {
	caller := &functionCaller{}
	caller.functionTable = map[string]FunctionEntry{
		"abs": {
			Name: "abs",
			Arguments: []ArgSpec{
				{Types: []JpType{JpNumber}},
			},
			Handler: jpfAbs,
		},
		"avg": {
			Name: "avg",
			Arguments: []ArgSpec{
				{Types: []JpType{JpArrayNumber}},
			},
			Handler: jpfAvg,
		},
		"ceil": {
			Name: "ceil",
			Arguments: []ArgSpec{
				{Types: []JpType{JpNumber}},
			},
			Handler: jpfCeil,
		},
		"contains": {
			Name: "contains",
			Arguments: []ArgSpec{
				{Types: []JpType{JpArray, JpString}},
				{Types: []JpType{JpAny}},
			},
			Handler: jpfContains,
		},
		"ends_with": {
			Name: "ends_with",
			Arguments: []ArgSpec{
				{Types: []JpType{JpString}},
				{Types: []JpType{JpString}},
			},
			Handler: jpfEndsWith,
		},
		"find_first": {
			Name: "find_first",
			Arguments: []ArgSpec{
				{Types: []JpType{JpString}},
				{Types: []JpType{JpString}},
				{Types: []JpType{JpNumber}, Optional: true},
				{Types: []JpType{JpNumber}, Optional: true},
			},
			Handler: jpfFindFirst,
		},
		"find_last": {
			Name: "find_last",
			Arguments: []ArgSpec{
				{Types: []JpType{JpString}},
				{Types: []JpType{JpString}},
				{Types: []JpType{JpNumber}, Optional: true},
				{Types: []JpType{JpNumber}, Optional: true},
			},
			Handler: jpfFindLast,
		},
		"floor": {
			Name: "floor",
			Arguments: []ArgSpec{
				{Types: []JpType{JpNumber}},
			},
			Handler: jpfFloor,
		},
		"from_items": {
			Name: "from_items",
			Arguments: []ArgSpec{
				{Types: []JpType{JpArrayArray}},
			},
			Handler: jpfFromItems,
		},
		"group_by": {
			Name: "group_by",
			Arguments: []ArgSpec{
				{Types: []JpType{JpArray}},
				{Types: []JpType{JpExpref}},
			},
			Handler: jpfGroupBy,
		},
		"items": {
			Name: "items",
			Arguments: []ArgSpec{
				{Types: []JpType{JpObject}},
			},
			Handler: jpfItems,
		},
		"join": {
			Name: "join",
			Arguments: []ArgSpec{
				{Types: []JpType{JpString}},
				{Types: []JpType{JpArrayString}},
			},
			Handler: jpfJoin,
		},
		"keys": {
			Name: "keys",
			Arguments: []ArgSpec{
				{Types: []JpType{JpObject}},
			},
			Handler: jpfKeys,
		},
		"length": {
			Name: "length",
			Arguments: []ArgSpec{
				{Types: []JpType{JpString, JpArray, JpObject}},
			},
			Handler: jpfLength,
		},
		"lower": {
			Name: "lower",
			Arguments: []ArgSpec{
				{Types: []JpType{JpString}},
			},
			Handler: jpfLower,
		},
		"let": {
			Name: "let",
			Arguments: []ArgSpec{
				{Types: []JpType{JpObject}},
				{Types: []JpType{JpExpref}},
			},
			Handler: jpfLet,
		},
		"map": {
			Name: "amp",
			Arguments: []ArgSpec{
				{Types: []JpType{JpExpref}},
				{Types: []JpType{JpArray}},
			},
			Handler: jpfMap,
		},
		"max": {
			Name: "max",
			Arguments: []ArgSpec{
				{Types: []JpType{JpArrayNumber, JpArrayString}},
			},
			Handler: jpfMax,
		},
		"max_by": {
			Name: "max_by",
			Arguments: []ArgSpec{
				{Types: []JpType{JpArray}},
				{Types: []JpType{JpExpref}},
			},
			Handler: jpfMaxBy,
		},
		"merge": {
			Name: "merge",
			Arguments: []ArgSpec{
				{Types: []JpType{JpObject}, Variadic: true},
			},
			Handler: jpfMerge,
		},
		"min": {
			Name: "min",
			Arguments: []ArgSpec{
				{Types: []JpType{JpArrayNumber, JpArrayString}},
			},
			Handler: jpfMin,
		},
		"min_by": {
			Name: "min_by",
			Arguments: []ArgSpec{
				{Types: []JpType{JpArray}},
				{Types: []JpType{JpExpref}},
			},
			Handler: jpfMinBy,
		},
		"not_null": {
			Name: "not_null",
			Arguments: []ArgSpec{
				{Types: []JpType{JpAny}, Variadic: true},
			},
			Handler: jpfNotNull,
		},
		"pad_left": {
			Name: "pad_left",
			Arguments: []ArgSpec{
				{Types: []JpType{JpString}},
				{Types: []JpType{JpNumber}},
				{Types: []JpType{JpString}, Optional: true},
			},
			Handler: jpfPadLeft,
		},
		"pad_right": {
			Name: "pad_right",
			Arguments: []ArgSpec{
				{Types: []JpType{JpString}},
				{Types: []JpType{JpNumber}},
				{Types: []JpType{JpString}, Optional: true},
			},
			Handler: jpfPadRight,
		},
		"replace": {
			Name: "replace",
			Arguments: []ArgSpec{
				{Types: []JpType{JpString}},
				{Types: []JpType{JpString}},
				{Types: []JpType{JpString}},
				{Types: []JpType{JpNumber}, Optional: true},
			},
			Handler: jpfReplace,
		},
		"reverse": {
			Name: "reverse",
			Arguments: []ArgSpec{
				{Types: []JpType{JpArray, JpString}},
			},
			Handler: jpfReverse,
		},
		"sort": {
			Name: "sort",
			Arguments: []ArgSpec{
				{Types: []JpType{JpArrayString, JpArrayNumber}},
			},
			Handler: jpfSort,
		},
		"sort_by": {
			Name: "sort_by",
			Arguments: []ArgSpec{
				{Types: []JpType{JpArray}},
				{Types: []JpType{JpExpref}},
			},
			Handler: jpfSortBy,
		},
		"split": {
			Name: "split",
			Arguments: []ArgSpec{
				{Types: []JpType{JpString}},
				{Types: []JpType{JpString}},
				{Types: []JpType{JpNumber}, Optional: true},
			},
			Handler: jpfSplit,
		},
		"starts_with": {
			Name: "starts_with",
			Arguments: []ArgSpec{
				{Types: []JpType{JpString}},
				{Types: []JpType{JpString}},
			},
			Handler: jpfStartsWith,
		},
		"sum": {
			Name: "sum",
			Arguments: []ArgSpec{
				{Types: []JpType{JpArrayNumber}},
			},
			Handler: jpfSum,
		},
		"to_array": {
			Name: "to_array",
			Arguments: []ArgSpec{
				{Types: []JpType{JpAny}},
			},
			Handler: jpfToArray,
		},
		"to_number": {
			Name: "to_number",
			Arguments: []ArgSpec{
				{Types: []JpType{JpAny}},
			},
			Handler: jpfToNumber,
		},
		"to_string": {
			Name: "to_string",
			Arguments: []ArgSpec{
				{Types: []JpType{JpAny}},
			},
			Handler: jpfToString,
		},
		"trim": {
			Name: "trim",
			Arguments: []ArgSpec{
				{Types: []JpType{JpString}},
				{Types: []JpType{JpString}, Optional: true},
			},
			Handler: jpfTrim,
		},
		"trim_left": {
			Name: "trim_left",
			Arguments: []ArgSpec{
				{Types: []JpType{JpString}},
				{Types: []JpType{JpString}, Optional: true},
			},
			Handler: jpfTrimLeft,
		},
		"trim_right": {
			Name: "trim_right",
			Arguments: []ArgSpec{
				{Types: []JpType{JpString}},
				{Types: []JpType{JpString}, Optional: true},
			},
			Handler: jpfTrimRight,
		},
		"type": {
			Name: "type",
			Arguments: []ArgSpec{
				{Types: []JpType{JpAny}},
			},
			Handler: jpfType,
		},
		"upper": {
			Name: "upper",
			Arguments: []ArgSpec{
				{Types: []JpType{JpString}},
			},
			Handler: jpfUpper,
		},
		"values": {
			Name: "values",
			Arguments: []ArgSpec{
				{Types: []JpType{JpObject}},
			},
			Handler: jpfValues,
		},
		"zip": {
			Name: "zip",
			Arguments: []ArgSpec{
				{Types: []JpType{JpArray}},
				{Types: []JpType{JpArray}, Variadic: true},
			},
			Handler: jpfZip,
		},
	}
	return caller
}

func (e *FunctionEntry) resolveArgs(arguments []interface{}) ([]interface{}, error) {
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
			err := spec.typeCheck(userArg)
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
			err := lastArg.typeCheck(userArg)
			if err != nil {
				return nil, err
			}
		}
	}
	return arguments, nil
}

func isVariadic(arguments []ArgSpec) bool {
	for _, spec := range arguments {
		if spec.Variadic {
			return true
		}
	}
	return false
}

func getMinExpected(arguments []ArgSpec) int {
	expected := 0
	for _, spec := range arguments {
		if !spec.Optional {
			expected++
		}
	}
	return expected
}

func getMaxExpected(arguments []ArgSpec) (int, bool) {
	if isVariadic(arguments) {
		return 0, false
	}
	return len(arguments), true
}

func (a *ArgSpec) typeCheck(arg interface{}) error {
	for _, t := range a.Types {
		switch t {
		case JpNumber:
			if _, ok := arg.(float64); ok {
				return nil
			}
		case JpString:
			if _, ok := arg.(string); ok {
				return nil
			}
		case JpArray:
			if util.IsSliceType(arg) {
				return nil
			}
		case JpObject:
			if _, ok := arg.(map[string]interface{}); ok {
				return nil
			}
		case JpArrayArray:
			if util.IsSliceType(arg) {
				if _, ok := arg.([]interface{}); ok {
					return nil
				}
			}
		case JpArrayNumber:
			if _, ok := util.ToArrayNum(arg); ok {
				return nil
			}
		case JpArrayString:
			if _, ok := util.ToArrayStr(arg); ok {
				return nil
			}
		case JpAny:
			return nil
		case JpExpref:
			if _, ok := arg.(expRef); ok {
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
	resolvedArgs, err := entry.resolveArgs(arguments)
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

func jpfAbs(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	num := arguments[0].(float64)
	return math.Abs(num), nil
}

func jpfAvg(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	// We've already type checked the value so we can safely use
	// type assertions.
	args := arguments[0].([]interface{})
	length := float64(len(args))
	if len(args) == 0 {
		return nil, nil
	}
	numerator := 0.0
	for _, n := range args {
		numerator += n.(float64)
	}
	return numerator / length, nil
}

func jpfCeil(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	val := arguments[0].(float64)
	return math.Ceil(val), nil
}

func jpfContains(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	search := arguments[0]
	el := arguments[1]
	if searchStr, ok := search.(string); ok {
		if elStr, ok := el.(string); ok {
			return strings.Contains(searchStr, elStr), nil
		}
		return false, nil
	}
	// Otherwise this is a generic contains for []interface{}
	general := search.([]interface{})
	for _, item := range general {
		if item == el {
			return true, nil
		}
	}
	return false, nil
}

func jpfEndsWith(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	search := arguments[0].(string)
	suffix := arguments[1].(string)
	return strings.HasSuffix(search, suffix), nil
}

func jpfFindImpl(name string, arguments []interface{}, find func(s string, substr string) int) (interface{}, error) {
	subject := arguments[0].(string)
	substr := arguments[1].(string)

	if len(subject) == 0 || len(substr) == 0 {
		return nil, nil
	}

	start := 0
	startSpecified := len(arguments) > 2
	if startSpecified {
		num, ok := util.ToInteger(arguments[2])
		if !ok {
			return nil, jperror.NotAnInteger(name, "start")
		}
		start = util.Max(0, num)
	}
	end := len(subject)
	endSpecified := len(arguments) > 3
	if endSpecified {
		num, ok := util.ToInteger(arguments[3])
		if !ok {
			return nil, jperror.NotAnInteger(name, "end")
		}
		end = util.Min(num, len(subject))
	}

	offset := find(subject[start:end], substr)

	if offset == -1 {
		return nil, nil
	}

	return float64(start + offset), nil
}

func jpfFindFirst(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	return jpfFindImpl("find_first", arguments, strings.Index)
}

func jpfFindLast(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	return jpfFindImpl("find_last", arguments, strings.LastIndex)
}

func jpfFloor(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	val := arguments[0].(float64)
	return math.Floor(val), nil
}

func jpfFromItems(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	if arr, ok := util.ToArrayArray(arguments[0]); ok {
		result := make(map[string]interface{})
		for _, item := range arr {
			if len(item) != 2 {
				return nil, errors.New("invalid value, each array must contain two elements, a pair of string and value")
			}
			first, ok := item[0].(string)
			if !ok {
				return nil, errors.New("invalid value, each array must contain two elements, a pair of string and value")
			}
			second := item[1]
			result[first] = second
		}
		return result, nil
	}
	return nil, errors.New("invalid type, first argument must be an array of arrays")
}

func jpfGroupBy(exec ExecuteFunc, arguments []interface{}) (interface{}, error) {
	arr := arguments[0].([]interface{})
	exp := arguments[1].(expRef)
	node := exp.ref
	if len(arr) == 0 {
		return nil, nil
	}
	groups := map[string]interface{}{}
	for _, element := range arr {
		spec, err := exec(node, element, nil)
		if err != nil {
			return nil, err
		}
		key, ok := spec.(string)
		if !ok {
			return nil, errors.New("invalid type, the expression must evaluate to a string")
		}
		if _, ok := groups[key]; !ok {
			groups[key] = []interface{}{}
		}
		groups[key] = append(groups[key].([]interface{}), element)
	}
	return groups, nil
}

func jpfItems(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	value := arguments[0].(map[string]interface{})
	arrays := []interface{}{}
	for key, item := range value {
		var element interface{} = []interface{}{key, item}
		arrays = append(arrays, element)
	}

	return arrays, nil
}

func jpfJoin(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	sep := arguments[0].(string)
	// We can't just do arguments[1].([]string), we have to
	// manually convert each item to a string.
	arrayStr := []string{}
	for _, item := range arguments[1].([]interface{}) {
		arrayStr = append(arrayStr, item.(string))
	}
	return strings.Join(arrayStr, sep), nil
}

func jpfKeys(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	arg := arguments[0].(map[string]interface{})
	collected := make([]interface{}, 0, len(arg))
	for key := range arg {
		collected = append(collected, key)
	}
	return collected, nil
}

func jpfLength(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	arg := arguments[0]
	if c, ok := arg.(string); ok {
		return float64(utf8.RuneCountInString(c)), nil
	} else if util.IsSliceType(arg) {
		v := reflect.ValueOf(arg)
		return float64(v.Len()), nil
	} else if c, ok := arg.(map[string]interface{}); ok {
		return float64(len(c)), nil
	}
	return nil, errors.New("could not compute length()")
}

func jpfLower(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	return strings.ToLower(arguments[0].(string)), nil
}

func jpfLet(exec ExecuteFunc, arguments []interface{}) (interface{}, error) {
	scope := arguments[0].(map[string]interface{})
	exp := arguments[1].(expRef)
	node := exp.ref
	context := exp.context

	result, err := exec(node, context, scope)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func jpfMap(exec ExecuteFunc, arguments []interface{}) (interface{}, error) {
	exp := arguments[0].(expRef)
	arr := arguments[1].([]interface{})
	node := exp.ref
	mapped := make([]interface{}, 0, len(arr))
	for _, value := range arr {
		current, err := exec(node, value, nil)
		if err != nil {
			return nil, err
		}
		mapped = append(mapped, current)
	}
	return mapped, nil
}

func jpfMax(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	if items, ok := util.ToArrayNum(arguments[0]); ok {
		if len(items) == 0 {
			return nil, nil
		}
		if len(items) == 1 {
			return items[0], nil
		}
		best := items[0]
		for _, item := range items[1:] {
			if item > best {
				best = item
			}
		}
		return best, nil
	}
	// Otherwise we're dealing with a max() of strings.
	items, _ := util.ToArrayStr(arguments[0])
	if len(items) == 0 {
		return nil, nil
	}
	if len(items) == 1 {
		return items[0], nil
	}
	best := items[0]
	for _, item := range items[1:] {
		if item > best {
			best = item
		}
	}
	return best, nil
}

func jpfMaxBy(exec ExecuteFunc, arguments []interface{}) (interface{}, error) {
	arr := arguments[0].([]interface{})
	exp := arguments[1].(expRef)
	node := exp.ref
	if len(arr) == 0 {
		return nil, nil
	} else if len(arr) == 1 {
		return arr[0], nil
	}
	start, err := exec(node, arr[0], nil)
	if err != nil {
		return nil, err
	}
	switch t := start.(type) {
	case float64:
		bestVal := t
		bestItem := arr[0]
		for _, item := range arr[1:] {
			result, err := exec(node, item, nil)
			if err != nil {
				return nil, err
			}
			current, ok := result.(float64)
			if !ok {
				return nil, errors.New("invalid type, must be number")
			}
			if current > bestVal {
				bestVal = current
				bestItem = item
			}
		}
		return bestItem, nil
	case string:
		bestVal := t
		bestItem := arr[0]
		for _, item := range arr[1:] {
			result, err := exec(node, item, nil)
			if err != nil {
				return nil, err
			}
			current, ok := result.(string)
			if !ok {
				return nil, errors.New("invalid type, must be string")
			}
			if current > bestVal {
				bestVal = current
				bestItem = item
			}
		}
		return bestItem, nil
	default:
		return nil, errors.New("invalid type, must be number of string")
	}
}

func jpfMerge(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	final := make(map[string]interface{})
	for _, m := range arguments {
		mapped := m.(map[string]interface{})
		for key, value := range mapped {
			final[key] = value
		}
	}
	return final, nil
}

func jpfMin(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	if items, ok := util.ToArrayNum(arguments[0]); ok {
		if len(items) == 0 {
			return nil, nil
		}
		if len(items) == 1 {
			return items[0], nil
		}
		best := items[0]
		for _, item := range items[1:] {
			if item < best {
				best = item
			}
		}
		return best, nil
	}
	items, _ := util.ToArrayStr(arguments[0])
	if len(items) == 0 {
		return nil, nil
	}
	if len(items) == 1 {
		return items[0], nil
	}
	best := items[0]
	for _, item := range items[1:] {
		if item < best {
			best = item
		}
	}
	return best, nil
}

func jpfMinBy(exec ExecuteFunc, arguments []interface{}) (interface{}, error) {
	arr := arguments[0].([]interface{})
	exp := arguments[1].(expRef)
	node := exp.ref
	if len(arr) == 0 {
		return nil, nil
	} else if len(arr) == 1 {
		return arr[0], nil
	}
	start, err := exec(node, arr[0], nil)
	if err != nil {
		return nil, err
	}
	if t, ok := start.(float64); ok {
		bestVal := t
		bestItem := arr[0]
		for _, item := range arr[1:] {
			result, err := exec(node, item, nil)
			if err != nil {
				return nil, err
			}
			current, ok := result.(float64)
			if !ok {
				return nil, errors.New("invalid type, must be number")
			}
			if current < bestVal {
				bestVal = current
				bestItem = item
			}
		}
		return bestItem, nil
	} else if t, ok := start.(string); ok {
		bestVal := t
		bestItem := arr[0]
		for _, item := range arr[1:] {
			result, err := exec(node, item, nil)
			if err != nil {
				return nil, err
			}
			current, ok := result.(string)
			if !ok {
				return nil, errors.New("invalid type, must be string")
			}
			if current < bestVal {
				bestVal = current
				bestItem = item
			}
		}
		return bestItem, nil
	} else {
		return nil, errors.New("invalid type, must be number of string")
	}
}

func jpfNotNull(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	for _, arg := range arguments {
		if arg != nil {
			return arg, nil
		}
	}
	return nil, nil
}

func jpfPadImpl(
	name string,
	arguments []interface{},
	pad func(s string, width int, pad string) string,
) (interface{}, error) {
	s := arguments[0].(string)
	width, ok := util.ToPositiveInteger(arguments[1])
	if !ok {
		return nil, jperror.NotAPositiveInteger(name, "width")
	}
	chars := " "
	if len(arguments) > 2 {
		chars = arguments[2].(string)
		if len(chars) > 1 {
			return nil, fmt.Errorf("invalid value, the function '%s' expects its 'pad' argument to be a string of length 1", name)
		}
	}

	return pad(s, width, chars), nil
}

func jpfPadLeft(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	return jpfPadImpl("pad_left", arguments, padLeft)
}

func jpfPadRight(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	return jpfPadImpl("pad_right", arguments, padRight)
}

func padLeft(s string, width int, pad string) string {
	length := util.Max(0, width-len(s))
	padding := strings.Repeat(pad, length)
	result := fmt.Sprintf("%s%s", padding, s)
	return result
}

func padRight(s string, width int, pad string) string {
	length := util.Max(0, width-len(s))
	padding := strings.Repeat(pad, length)
	result := fmt.Sprintf("%s%s", s, padding)
	return result
}

func jpfReplace(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	subject := arguments[0].(string)
	old := arguments[1].(string)
	new := arguments[2].(string)
	count := -1
	if len(arguments) > 3 {
		num, ok := util.ToPositiveInteger(arguments[3])
		if !ok {
			return nil, jperror.NotAPositiveInteger("replace", "count")
		}
		count = num
	}

	return strings.Replace(subject, old, new, count), nil
}

func jpfReverse(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	if s, ok := arguments[0].(string); ok {
		r := []rune(s)
		for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
			r[i], r[j] = r[j], r[i]
		}
		return string(r), nil
	}
	items := arguments[0].([]interface{})
	length := len(items)
	reversed := make([]interface{}, length)
	for i, item := range items {
		reversed[length-(i+1)] = item
	}
	return reversed, nil
}

func jpfSort(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	if items, ok := util.ToArrayNum(arguments[0]); ok {
		d := sort.Float64Slice(items)
		sort.Stable(d)
		final := make([]interface{}, len(d))
		for i, val := range d {
			final[i] = val
		}
		return final, nil
	}
	// Otherwise we're dealing with sort()'ing strings.
	items, _ := util.ToArrayStr(arguments[0])
	d := sort.StringSlice(items)
	sort.Stable(d)
	final := make([]interface{}, len(d))
	for i, val := range d {
		final[i] = val
	}
	return final, nil
}

func jpfSortBy(exec ExecuteFunc, arguments []interface{}) (interface{}, error) {
	arr := arguments[0].([]interface{})
	exp := arguments[1].(expRef)
	node := exp.ref
	if len(arr) == 0 {
		return arr, nil
	} else if len(arr) == 1 {
		return arr, nil
	}
	var sortKeys []interface{}
	for _, item := range arr {
		if value, err := exec(node, item, nil); err != nil {
			return nil, err
		} else {
			sortKeys = append(sortKeys, value)
		}
	}
	if _, ok := sortKeys[0].(float64); ok {
		sortable := &byExprFloat{arr, sortKeys}
		sort.Stable(sortable)
		return arr, nil
	} else if _, ok := sortKeys[0].(string); ok {
		sortable := &byExprString{arr, sortKeys}
		sort.Stable(sortable)
		return arr, nil
	} else {
		return nil, errors.New("invalid type, must be number of string")
	}
}

func jpfSplit(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	s := arguments[0].(string)
	if len(s) == 0 {
		return []interface{}{}, nil
	}

	sep := arguments[1].(string)
	n := 0
	nSpecified := len(arguments) > 2
	if nSpecified {
		num, ok := util.ToPositiveInteger(arguments[2])
		if !ok {
			return nil, jperror.NotAPositiveInteger("split", "count")
		}
		n = num
	}

	if nSpecified && n == 0 {
		result := []interface{}{s}
		return result, nil
	}

	count := -1
	if nSpecified {
		count = n + 1
	}
	splits := strings.SplitN(s, sep, count)

	// convert []string to []interface{} ☹️

	result := []interface{}{}
	for _, split := range splits {
		result = append(result, split)
	}
	return result, nil
}

func jpfStartsWith(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	search := arguments[0].(string)
	prefix := arguments[1].(string)
	return strings.HasPrefix(search, prefix), nil
}

func jpfSum(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	items, _ := util.ToArrayNum(arguments[0])
	sum := 0.0
	for _, item := range items {
		sum += item
	}
	return sum, nil
}

func jpfToArray(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	if _, ok := arguments[0].([]interface{}); ok {
		return arguments[0], nil
	}
	return arguments[:1:1], nil
}

func jpfToString(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	if v, ok := arguments[0].(string); ok {
		return v, nil
	}
	result, err := json.Marshal(arguments[0])
	if err != nil {
		return nil, err
	}
	return string(result), nil
}

func jpfToNumber(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	arg := arguments[0]
	if v, ok := arg.(float64); ok {
		return v, nil
	}
	if v, ok := arg.(string); ok {
		conv, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, nil
		}
		return conv, nil
	}
	if _, ok := arg.([]interface{}); ok {
		return nil, nil
	}
	if _, ok := arg.(map[string]interface{}); ok {
		return nil, nil
	}
	if arg == nil {
		return nil, nil
	}
	if arg == true || arg == false {
		return nil, nil
	}
	return nil, errors.New("unknown type")
}

func jpfTrimImpl(
	arguments []interface{},
	trimSpace func(s string, predicate func(r rune) bool) string,
	trim func(s string, cutset string) string,
) (interface{}, error) {
	s := arguments[0].(string)
	cutset := ""
	if len(arguments) > 1 {
		cutset = arguments[1].(string)
	}

	if len(cutset) == 0 {
		return trimSpace(s, unicode.IsSpace), nil
	}
	return trim(s, cutset), nil
}

func jpfTrim(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	return jpfTrimImpl(arguments, strings.TrimFunc, strings.Trim)
}

func jpfTrimLeft(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	return jpfTrimImpl(arguments, strings.TrimLeftFunc, strings.TrimLeft)
}

func jpfTrimRight(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	return jpfTrimImpl(arguments, strings.TrimRightFunc, strings.TrimRight)
}

func jpfType(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	arg := arguments[0]
	if _, ok := arg.(float64); ok {
		return "number", nil
	}
	if _, ok := arg.(string); ok {
		return "string", nil
	}
	if _, ok := arg.([]interface{}); ok {
		return "array", nil
	}
	if _, ok := arg.(map[string]interface{}); ok {
		return "object", nil
	}
	if arg == nil {
		return "null", nil
	}
	if arg == true || arg == false {
		return "boolean", nil
	}
	return nil, errors.New("unknown type")
}

func jpfUpper(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	return strings.ToUpper(arguments[0].(string)), nil
}

func jpfValues(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	arg := arguments[0].(map[string]interface{})
	collected := make([]interface{}, 0, len(arg))
	for _, value := range arg {
		collected = append(collected, value)
	}
	return collected, nil
}

func jpfZip(_ ExecuteFunc, arguments []interface{}) (interface{}, error) {
	// determine how many items are present
	// for each array in the result

	count := math.MaxInt
	for _, item := range arguments {
		arr := item.([]interface{})
		// TODO: use go1.18 min[T constraints.Ordered] generic function
		count = int(math.Min(float64(count), float64(len(arr))))
	}

	result := []interface{}{}

	for i := 0; i < count; i++ {
		nth := []interface{}{}
		for _, item := range arguments {
			arr := item.([]interface{})
			nth = append(nth, arr[i])
		}
		result = append(result, interface{}(nth))
	}

	return result, nil
}
