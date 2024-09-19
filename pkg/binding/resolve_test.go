package binding

import (
	"reflect"
	"testing"
)

func TestResolve(t *testing.T) {
	type args struct {
		name     string
		bindings Bindings
	}
	tests := []struct {
		name    string
		args    args
		want    any
		wantErr bool
	}{{
		name: "nil",
		args: args{
			name:     "$test",
			bindings: nil,
		},
		want:    nil,
		wantErr: true,
	}, {
		name: "empty",
		args: args{
			name:     "",
			bindings: NewBindings().Register("$test", NewBinding(42)),
		},
		want:    nil,
		wantErr: true,
	}, {
		name: "ok",
		args: args{
			name:     "$test",
			bindings: NewBindings().Register("$test", NewBinding(42)),
		},
		want:    42,
		wantErr: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Resolve(tt.args.name, tt.args.bindings)
			if (err != nil) != tt.wantErr {
				t.Errorf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Resolve() = %v, want %v", got, tt.want)
			}
		})
	}
}
