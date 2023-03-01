package jmespath

import (
	"errors"
	"math"
	"reflect"
	"unicode"
	"unicode/utf8"

	"github.com/jmespath-community/go-jmespath/pkg/util"
)

/* This is a tree based interpreter.  It walks the AST and directly
   interprets the AST to search through a JSON document.
*/

type treeInterpreter struct {
	fCall  *functionCaller
	scopes *scopes
}

func newInterpreter(data interface{}) *treeInterpreter {
	interpreter := treeInterpreter{
		scopes: newScopes(),
	}
	interpreter.scopes.pushScope(map[string]interface{}{"$": data})

	interpreter.fCall = newFunctionCaller(interpreter.scopes)

	return &interpreter
}

type expRef struct {
	ref     ASTNode
	context interface{}
}

// Execute takes an ASTNode and input data and interprets the AST directly.
// It will produce the result of applying the JMESPath expression associated
// with the ASTNode to the input data "value".
func (intr *treeInterpreter) Execute(node ASTNode, value interface{}) (interface{}, error) {
	switch node.nodeType {
	case ASTArithmeticUnaryExpression:
		expr, err := intr.Execute(node.children[0], value)
		if err != nil {
			return nil, err
		}
		num, ok := expr.(float64)
		if !ok {
			return nil, nil
		}
		switch node.value {
		case tPlus:
			return num, nil
		case tMinus:
			return -num, nil
		}
	case ASTArithmeticExpression:
		left, err := intr.Execute(node.children[0], value)
		if err != nil {
			return nil, err
		}
		right, err := intr.Execute(node.children[1], value)
		if err != nil {
			return nil, err
		}
		leftNum, ok := left.(float64)
		if !ok {
			return nil, nil
		}
		rightNum, ok := right.(float64)
		if !ok {
			return nil, nil
		}
		switch node.value {
		case tPlus:
			return leftNum + rightNum, nil
		case tMinus:
			return leftNum - rightNum, nil
		case tStar:
			return leftNum * rightNum, nil
		case tMultiply:
			return leftNum * rightNum, nil
		case tDivide:
			return leftNum / rightNum, nil
		case tModulo:
			return math.Mod(leftNum, rightNum), nil
		case tDiv:
			return math.Floor(leftNum / rightNum), nil
		}
	case ASTComparator:
		left, err := intr.Execute(node.children[0], value)
		if err != nil {
			return nil, err
		}
		right, err := intr.Execute(node.children[1], value)
		if err != nil {
			return nil, err
		}
		switch node.value {
		case tEQ:
			return util.ObjsEqual(left, right), nil
		case tNE:
			return !util.ObjsEqual(left, right), nil
		}
		leftNum, ok := left.(float64)
		if !ok {
			return nil, nil
		}
		rightNum, ok := right.(float64)
		if !ok {
			return nil, nil
		}
		switch node.value {
		case tGT:
			return leftNum > rightNum, nil
		case tGTE:
			return leftNum >= rightNum, nil
		case tLT:
			return leftNum < rightNum, nil
		case tLTE:
			return leftNum <= rightNum, nil
		}
	case ASTExpRef:
		return expRef{ref: node.children[0], context: value}, nil
	case ASTFunctionExpression:
		resolvedArgs := []interface{}{}
		for _, arg := range node.children {
			current, err := intr.Execute(arg, value)
			if err != nil {
				return nil, err
			}
			resolvedArgs = append(resolvedArgs, current)
		}
		return intr.fCall.CallFunction(node.value.(string), resolvedArgs, intr)
	case ASTField:
		key := node.value.(string)
		var result interface{}
		if m, ok := value.(map[string]interface{}); ok {
			result = m[key]
			if result != nil {
				return result, nil
			}
		}
		result, err := intr.fieldFromStruct(node.value.(string), value)
		if err != nil {
			return nil, err
		}
		if result != nil {
			return result, nil
		}
		if result, ok := intr.scopes.getValue(key); ok {
			return result, nil
		}
		return nil, nil
	case ASTFilterProjection:
		left, err := intr.Execute(node.children[0], value)
		if err != nil {
			return nil, nil
		}
		sliceType, ok := left.([]interface{})
		if !ok {
			if util.IsSliceType(left) {
				return intr.filterProjectionWithReflection(node, left)
			}
			return nil, nil
		}
		compareNode := node.children[2]
		collected := []interface{}{}
		for _, element := range sliceType {
			result, err := intr.Execute(compareNode, element)
			if err != nil {
				return nil, err
			}
			if !util.IsFalse(result) {
				current, err := intr.Execute(node.children[1], element)
				if err != nil {
					return nil, err
				}
				if current != nil {
					collected = append(collected, current)
				}
			}
		}
		return collected, nil
	case ASTFlatten:
		left, err := intr.Execute(node.children[0], value)
		if err != nil {
			return nil, nil
		}
		sliceType, ok := left.([]interface{})
		if !ok {
			// If we can't type convert to []interface{}, there's
			// a chance this could still work via reflection if we're
			// dealing with user provided types.
			if util.IsSliceType(left) {
				return intr.flattenWithReflection(left)
			}
			return nil, nil
		}
		flattened := []interface{}{}
		for _, element := range sliceType {
			if elementSlice, ok := element.([]interface{}); ok {
				flattened = append(flattened, elementSlice...)
			} else if util.IsSliceType(element) {
				reflectFlat := []interface{}{}
				v := reflect.ValueOf(element)
				for i := 0; i < v.Len(); i++ {
					reflectFlat = append(reflectFlat, v.Index(i).Interface())
				}
				flattened = append(flattened, reflectFlat...)
			} else {
				flattened = append(flattened, element)
			}
		}
		return flattened, nil
	case ASTIdentity, ASTCurrentNode:
		return value, nil
	case ASTRootNode:
		if result, ok := intr.scopes.getValue("$"); ok {
			return result, nil
		}
		return nil, nil
	case ASTIndex:
		if sliceType, ok := value.([]interface{}); ok {
			index := node.value.(int)
			if index < 0 {
				index += len(sliceType)
			}
			if index < len(sliceType) && index >= 0 {
				return sliceType[index], nil
			}
			return nil, nil
		}
		// Otherwise try via reflection.
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Slice {
			index := node.value.(int)
			if index < 0 {
				index += rv.Len()
			}
			if index < rv.Len() && index >= 0 {
				v := rv.Index(index)
				return v.Interface(), nil
			}
		}
		return nil, nil
	case ASTKeyValPair:
		return intr.Execute(node.children[0], value)
	case ASTLiteral:
		return node.value, nil
	case ASTMultiSelectHash:
		if value == nil {
			return nil, nil
		}
		collected := make(map[string]interface{})
		for _, child := range node.children {
			current, err := intr.Execute(child, value)
			if err != nil {
				return nil, err
			}
			key := child.value.(string)
			collected[key] = current
		}
		return collected, nil
	case ASTMultiSelectList:
		collected := []interface{}{}
		for _, child := range node.children {
			current, err := intr.Execute(child, value)
			if err != nil {
				return nil, err
			}
			collected = append(collected, current)
		}
		return collected, nil
	case ASTOrExpression:
		matched, err := intr.Execute(node.children[0], value)
		if err != nil {
			return nil, err
		}
		if util.IsFalse(matched) {
			matched, err = intr.Execute(node.children[1], value)
			if err != nil {
				return nil, err
			}
		}
		return matched, nil
	case ASTAndExpression:
		matched, err := intr.Execute(node.children[0], value)
		if err != nil {
			return nil, err
		}
		if util.IsFalse(matched) {
			return matched, nil
		}
		return intr.Execute(node.children[1], value)
	case ASTNotExpression:
		matched, err := intr.Execute(node.children[0], value)
		if err != nil {
			return nil, err
		}
		if util.IsFalse(matched) {
			return true, nil
		}
		return false, nil
	case ASTPipe:
		result := value
		var err error
		for _, child := range node.children {
			result, err = intr.Execute(child, result)
			if err != nil {
				return nil, err
			}
		}
		return result, nil
	case ASTProjection:

		// projections typically operate on array | slices
		// string slicing produces an ASTProjection whose
		// first child is an ASTIndexExpression whose
		// second child is an ASTSlice

		// we allow execution of the left index-expression
		// to return a string only if the AST has this
		// specific shape

		allowString := false
		first_child := node.children[0]
		if first_child.nodeType == ASTIndexExpression {
			nested_children := first_child.children
			if len(nested_children) > 1 && nested_children[1].nodeType == ASTSlice {
				allowString = true
			}
		}

		left, err := intr.Execute(node.children[0], value)
		if err != nil {
			return nil, err
		}

		sliceType, ok := left.([]interface{})
		if !ok {
			if util.IsSliceType(left) {
				return intr.projectWithReflection(node, left)
			}
			stringType, ok := left.(string)
			if allowString && ok {
				return stringType, nil
			} else {
				return nil, nil
			}
		}
		collected := []interface{}{}
		var current interface{}
		for _, element := range sliceType {
			current, err = intr.Execute(node.children[1], element)
			if err != nil {
				return nil, err
			}
			if current != nil {
				collected = append(collected, current)
			}
		}
		return collected, nil
	case ASTSubexpression, ASTIndexExpression:
		left, err := intr.Execute(node.children[0], value)
		if err != nil {
			return nil, err
		}
		if left == nil {
			return nil, nil
		}
		return intr.Execute(node.children[1], left)
	case ASTSlice:
		parts := node.value.([]*int)
		sliceType, ok := value.([]interface{})
		if !ok {
			if util.IsSliceType(value) {
				return intr.sliceWithReflection(node, value)
			} else {

				// string slices is implemented by slicing
				// the corresponding array of runes and
				// converting the result back to a string

				if stringType, ok := value.(string); ok {
					runeType := []rune(stringType)
					sliceParams := util.MakeSliceParams(parts)
					runes, err := util.Slice(runeType, sliceParams)
					if err != nil {
						return nil, nil
					}
					return string(runes), nil
				} else {
					return nil, nil
				}
			}
		}
		sliceParams := util.MakeSliceParams(parts)
		return util.Slice(sliceType, sliceParams)
	case ASTValueProjection:
		left, err := intr.Execute(node.children[0], value)
		if err != nil {
			return nil, nil
		}
		mapType, ok := left.(map[string]interface{})
		if !ok {
			return nil, nil
		}
		values := []interface{}{}
		for _, value := range mapType {
			values = append(values, value)
		}
		collected := []interface{}{}
		for _, element := range values {
			current, err := intr.Execute(node.children[1], element)
			if err != nil {
				return nil, err
			}
			if current != nil {
				collected = append(collected, current)
			}
		}
		return collected, nil
	}
	return nil, errors.New("Unknown AST node: " + node.nodeType.String())
}

func (intr *treeInterpreter) fieldFromStruct(key string, value interface{}) (interface{}, error) {
	rv := reflect.ValueOf(value)
	first, n := utf8.DecodeRuneInString(key)
	fieldName := string(unicode.ToUpper(first)) + key[n:]
	if rv.Kind() == reflect.Struct {
		v := rv.FieldByName(fieldName)
		if !v.IsValid() {
			return nil, nil
		}
		return v.Interface(), nil
	} else if rv.Kind() == reflect.Ptr {
		// Handle multiple levels of indirection?
		if rv.IsNil() {
			return nil, nil
		}
		rv = rv.Elem()
		v := rv.FieldByName(fieldName)
		if !v.IsValid() {
			return nil, nil
		}
		return v.Interface(), nil
	}
	return nil, nil
}

func (intr *treeInterpreter) flattenWithReflection(value interface{}) (interface{}, error) {
	v := reflect.ValueOf(value)
	flattened := []interface{}{}
	for i := 0; i < v.Len(); i++ {
		element := v.Index(i).Interface()
		if reflect.TypeOf(element).Kind() == reflect.Slice {
			// Then insert the contents of the element
			// slice into the flattened slice,
			// i.e flattened = append(flattened, mySlice...)
			elementV := reflect.ValueOf(element)
			for j := 0; j < elementV.Len(); j++ {
				flattened = append(
					flattened, elementV.Index(j).Interface())
			}
		} else {
			flattened = append(flattened, element)
		}
	}
	return flattened, nil
}

func (intr *treeInterpreter) sliceWithReflection(node ASTNode, value interface{}) (interface{}, error) {
	v := reflect.ValueOf(value)
	parts := node.value.([]*int)
	sliceParams := make([]util.SliceParam, 3)
	for i, part := range parts {
		if part != nil {
			sliceParams[i].Specified = true
			sliceParams[i].N = *part
		}
	}
	final := []interface{}{}
	for i := 0; i < v.Len(); i++ {
		element := v.Index(i).Interface()
		final = append(final, element)
	}
	return util.Slice(final, sliceParams)
}

func (intr *treeInterpreter) filterProjectionWithReflection(node ASTNode, value interface{}) (interface{}, error) {
	compareNode := node.children[2]
	collected := []interface{}{}
	v := reflect.ValueOf(value)
	for i := 0; i < v.Len(); i++ {
		element := v.Index(i).Interface()
		result, err := intr.Execute(compareNode, element)
		if err != nil {
			return nil, err
		}
		if !util.IsFalse(result) {
			current, err := intr.Execute(node.children[1], element)
			if err != nil {
				return nil, err
			}
			if current != nil {
				collected = append(collected, current)
			}
		}
	}
	return collected, nil
}

func (intr *treeInterpreter) projectWithReflection(node ASTNode, value interface{}) (interface{}, error) {
	collected := []interface{}{}
	v := reflect.ValueOf(value)
	for i := 0; i < v.Len(); i++ {
		element := v.Index(i).Interface()
		result, err := intr.Execute(node.children[1], element)
		if err != nil {
			return nil, err
		}
		if result != nil {
			collected = append(collected, result)
		}
	}
	return collected, nil
}
