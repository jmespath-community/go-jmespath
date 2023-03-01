package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlicePositiveStep(t *testing.T) {
	assert := assert.New(t)
	input := make([]interface{}, 5)
	input[0] = 0
	input[1] = 1
	input[2] = 2
	input[3] = 3
	input[4] = 4
	result, err := Slice(input, []SliceParam{{0, true}, {3, true}, {1, true}})
	assert.Nil(err)
	assert.Equal(input[:3], result)
}

func TestIsFalseJSONTypes(t *testing.T) {
	assert := assert.New(t)
	assert.True(IsFalse(false))
	assert.True(IsFalse(""))
	var empty []interface{}
	assert.True(IsFalse(empty))
	m := make(map[string]interface{})
	assert.True(IsFalse(m))
	assert.True(IsFalse(nil))

}

func TestIsFalseWithUserDefinedStructs(t *testing.T) {
	assert := assert.New(t)
	type nilStructType struct {
		SliceOfPointers []*string
	}
	nilStruct := nilStructType{SliceOfPointers: nil}
	assert.True(IsFalse(nilStruct.SliceOfPointers))

	// A user defined struct will never be false though,
	// even if it's fields are the zero type.
	assert.False(IsFalse(nilStruct))
}

func TestIsFalseWithNilInterface(t *testing.T) {
	assert := assert.New(t)
	var a *int
	var nilInterface interface{} = a
	assert.True(IsFalse(nilInterface))
}

func TestIsFalseWithMapOfUserStructs(t *testing.T) {
	assert := assert.New(t)
	type foo struct {
		Bar string
		Baz string
	}
	m := make(map[int]foo)
	assert.True(IsFalse(m))
}

func TestObjsEqual(t *testing.T) {
	assert := assert.New(t)
	assert.True(ObjsEqual("foo", "foo"))
	assert.True(ObjsEqual(20, 20))
	assert.True(ObjsEqual([]int{1, 2, 3}, []int{1, 2, 3}))
	assert.True(ObjsEqual(nil, nil))
	assert.True(!ObjsEqual(nil, "foo"))
	assert.True(ObjsEqual([]int{}, []int{}))
	assert.True(!ObjsEqual([]int{}, nil))
}
