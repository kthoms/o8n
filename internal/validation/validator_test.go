package validation

import (
	"reflect"
	"testing"
)

func TestValidateAndParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		typ     string
		want    interface{}
		wantErr bool
	}{
		{"bool true", "true", "bool", true, false},
		{"bool false", "false", "bool", false, false},
		{"bool numeric true", "1", "bool", true, false},
		{"bool invalid", "yes", "bool", nil, true},
		{"int valid", "42", "int", int64(42), false},
		{"int invalid", "4.2", "int", nil, true},
		{"number valid", "3.14", "number", float64(3.14), false},
		{"number invalid", "abc", "number", nil, true},
		{"json valid", "{\"k\": 1}", "json", map[string]interface{}{"k": float64(1)}, false},
		{"text default", "hello", "text", "hello", false},
		{"empty text", "", "text", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateAndParse(tt.input, tt.typ)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got nil, result=%v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// Compare kinds for numeric types
			if tt.want != nil {
				if reflect.TypeOf(tt.want) != reflect.TypeOf(got) {
					t.Fatalf("type mismatch: want %T got %T (value %v)", tt.want, got, got)
				}
				// For maps, use DeepEqual
				if reflect.TypeOf(tt.want).Kind() == reflect.Map {
					if !reflect.DeepEqual(tt.want, got) {
						t.Fatalf("value mismatch: want %#v got %#v", tt.want, got)
					}
				} else if !reflect.DeepEqual(tt.want, got) {
					t.Fatalf("value mismatch: want %#v got %#v", tt.want, got)
				}
			}
		})
	}
}
