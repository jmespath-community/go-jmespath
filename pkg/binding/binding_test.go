package binding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBinding(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  *binding
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
		value: []any{"42", 42},
		want:  &binding{[]any{"42", 42}},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewBinding(tt.value)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_binding_Value(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		want    any
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
		value:   []any{"42", 42},
		want:    []any{"42", 42},
		wantErr: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subj := NewBinding(tt.value)
			got, err := subj.Value()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_delegate_Value(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		want    any
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
		value:   []any{"42", 42},
		want:    []any{"42", 42},
		wantErr: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subj := NewDelegate(func() (any, error) {
				return tt.value, nil
			})
			got, err := subj.Value()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
