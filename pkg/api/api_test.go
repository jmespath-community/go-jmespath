package api

import (
	"encoding/json"
	"testing"

	"github.com/jmespath-community/go-jmespath/pkg/functions"
	"github.com/jmespath-community/go-jmespath/pkg/interpreter"
	"github.com/stretchr/testify/assert"
)

func TestValidUncompiledExpressionSearches(t *testing.T) {
	assert := assert.New(t)
	j := []byte(`{"foo": {"bar": {"baz": [0, 1, 2, 3, 4]}}}`)
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

func jpfEcho(arguments []interface{}) (interface{}, error) {
	return arguments[0], nil
}

func TestSearch(t *testing.T) {
	type Label string

	type args struct {
		expression string
		data       interface{}
		funcs      []functions.FunctionEntry
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
	}, {
		args: args{
			expression: "sort_by(@, &@ *`-1.0`)",
			data:       []interface{}{1.0, 2.0, 3.0, 4.0, 5.0},
		},
		want: []interface{}{5.0, 4.0, 3.0, 2.0, 1.0},
	}, {
		args: args{
			expression: "echo(@)",
			data:       []interface{}{1.0, 2.0, 3.0, 4.0, 5.0},
			funcs: []functions.FunctionEntry{{
				Name:    "echo",
				Handler: jpfEcho,
				Arguments: []functions.ArgSpec{
					{Types: []functions.JpType{functions.JpAny}},
				},
			}},
		},
		want: []interface{}{1.0, 2.0, 3.0, 4.0, 5.0},
	}, {
		args: args{
			expression: "echo(@)",
			data:       "abc",
			funcs: []functions.FunctionEntry{{
				Name:    "echo",
				Handler: jpfEcho,
				Arguments: []functions.ArgSpec{
					{Types: []functions.JpType{functions.JpAny}},
				},
			}},
		},
		want: "abc",
	}, {
		args: args{
			expression: "echo(@)",
			data:       42.0,
			funcs: []functions.FunctionEntry{{
				Name:    "echo",
				Handler: jpfEcho,
				Arguments: []functions.ArgSpec{
					{Types: []functions.JpType{functions.JpAny}},
				},
			}},
		},
		want: 42.0,
	}, {
		args: args{
			expression: "echo(@)",
			data:       42.0,
		},
		wantErr: true,
	}, {
		args: args{
			expression: `@."$".a`,
			data: map[string]interface{}{
				"a": 42.0,
			},
		},
		want: nil,
	}, {
		args: args{
			expression: "`null` | {foo: @}",
		},
		want: map[string]interface{}{
			"foo": nil,
		},
	}, {
		args: args{
			expression: "let $root = @ in $root.a",
			data: map[string]interface{}{
				"a": 42.0,
			},
		},
		want: 42.0,
	}, {
		args: args{
			expression: "contains(@, { foo: 'bar' })",
			data:       []interface{}{map[string]any{}, nil, map[string]any{"foo": "bar"}},
		},
		want: true,
	}, {
		args: args{
			expression: "length(@[?metric.__name__ == 'foo'])",
			data: []struct {
				Metric map[Label]any
			}{{
				Metric: map[Label]any{
					"__name__": "foo",
				},
			}, {
				Metric: map[Label]any{
					"__name__": "bar",
				},
			}},
		},
		want: 1.0,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			var opts []interpreter.Option
			if len(tt.args.funcs) != 0 {
				var f []functions.FunctionEntry
				f = append(f, functions.GetDefaultFunctions()...)
				f = append(f, tt.args.funcs...)
				caller := interpreter.NewFunctionCaller(f...)
				opts = append(opts, interpreter.WithFunctionCaller(caller))
			}
			got, err := Search(tt.args.expression, tt.args.data, opts...)
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
