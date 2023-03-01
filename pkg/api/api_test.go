package api

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidUncompiledExpressionSearches(t *testing.T) {
	assert := assert.New(t)
	var j = []byte(`{"foo": {"bar": {"baz": [0, 1, 2, 3, 4]}}}`)
	var d interface{}
	err := json.Unmarshal(j, &d)
	assert.Nil(err)
	result, err := Search("foo.bar.baz[2]", d)
	assert.Nil(err)
	assert.Equal(2.0, result)
}

func TestValidPrecompiledExpressionSearches(t *testing.T) {
	assert := assert.New(t)
	data := make(map[string]interface{})
	data["foo"] = "bar"
	precompiled, err := Compile("foo")
	assert.Nil(err)
	result, err := precompiled.Search(data)
	assert.Nil(err)
	assert.Equal("bar", result)
}

func TestInvalidPrecompileErrors(t *testing.T) {
	assert := assert.New(t)
	_, err := Compile("not a valid expression")
	assert.NotNil(err)
}

func TestInvalidMustCompilePanics(t *testing.T) {
	defer func() {
		r := recover()
		assert.NotNil(t, r)
	}()
	MustCompile("not a valid expression")
}

func TestSearch(t *testing.T) {
	type args struct {
		expression string
		data       interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{{
		args: args{
			expression: "not a valid expression",
		},
		wantErr: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			got, err := Search(tt.args.expression, tt.args.data)
			assert.Equal(tt.wantErr, err != nil)
			assert.Equal(tt.want, got)
		})
	}
}

func TestMustCompile(t *testing.T) {
	type args struct {
		expression string
	}
	tests := []struct {
		name      string
		args      args
		wantPanic bool
	}{{
		args: args{
			expression: "not a valid expression",
		},
		wantPanic: true,
	}, {
		args: args{
			expression: "foo.bar.baz[2]",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			if tt.wantPanic {
				defer func() {
					r := recover()
					assert.NotNil(r)
				}()
			}
			got := MustCompile(tt.args.expression)
			assert.NotNil(got)
		})
	}
}
