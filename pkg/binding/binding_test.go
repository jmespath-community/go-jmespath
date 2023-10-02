package binding

import (
	"reflect"
	"testing"
)

func TestNewBinding(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  Binding
	}{{
		name:  "nil",
		value: nil,
		want:  &binding{nil},
	}, {
		name:  "int",
		value: int(42),
		want:  &binding{int(42)},
	}, {
		name:  "string",
		value: "42",
		want:  &binding{"42"},
	}, {
		name:  "array",
		value: []interface{}{"42", 42},
		want:  &binding{[]interface{}{"42", 42}},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBinding(tt.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBinding() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_binding_Value(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		want    interface{}
		wantErr bool
	}{{
		name:    "nil",
		value:   nil,
		want:    nil,
		wantErr: false,
	}, {
		name:    "int",
		value:   int(42),
		want:    int(42),
		wantErr: false,
	}, {
		name:    "string",
		value:   "42",
		want:    "42",
		wantErr: false,
	}, {
		name:    "array",
		value:   []interface{}{"42", 42},
		want:    []interface{}{"42", 42},
		wantErr: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &binding{
				value: tt.value,
			}
			got, err := b.Value()
			if (err != nil) != tt.wantErr {
				t.Errorf("binding.Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("binding.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}
