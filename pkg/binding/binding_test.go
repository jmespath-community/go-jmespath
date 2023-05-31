package binding

import (
	"reflect"
	"testing"
)

func TestNewBindings(t *testing.T) {
	tests := []struct {
		name string
		want Bindings
	}{{
		want: bindings{},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBindings(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBindings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_bindings_Get(t *testing.T) {
	type fields struct {
		values map[string]interface{}
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    interface{}
		wantErr bool
	}{{
		fields: fields{
			values: nil,
		},
		args: args{
			name: "$root",
		},
		wantErr: true,
	}, {
		fields: fields{
			values: map[string]interface{}{},
		},
		args: args{
			name: "$root",
		},
		wantErr: true,
	}, {
		fields: fields{
			values: map[string]interface{}{
				"$root": 42.0,
			},
		},
		args: args{
			name: "$root",
		},
		want: 42.0,
	}, {
		fields: fields{
			values: map[string]interface{}{
				"$foot": 42.0,
			},
		},
		args: args{
			name: "$root",
		},
		wantErr: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := bindings{
				values: tt.fields.values,
			}
			got, err := b.Get(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("bindings.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("bindings.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_bindings_Register(t *testing.T) {
	type fields struct {
		values map[string]interface{}
	}
	type args struct {
		name  string
		value interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Bindings
	}{{
		fields: fields{
			values: nil,
		},
		args: args{
			name:  "$root",
			value: 42.0,
		},
		want: bindings{
			values: map[string]interface{}{
				"$root": 42.0,
			},
		},
	}, {
		fields: fields{
			values: map[string]interface{}{
				"$root": 21.0,
			},
		},
		args: args{
			name:  "$root",
			value: 42.0,
		},
		want: bindings{
			values: map[string]interface{}{
				"$root": 42.0,
			},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := bindings{
				values: tt.fields.values,
			}
			if got := b.Register(tt.args.name, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("bindings.Register() = %v, want %v", got, tt.want)
			}
		})
	}
}
