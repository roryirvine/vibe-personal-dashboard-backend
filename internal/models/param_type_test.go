package models

import "testing"

func TestParamType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		paramType ParamType
		want     bool
	}{
		{"string is valid", ParamTypeString, true},
		{"int is valid", ParamTypeInt, true},
		{"float is valid", ParamTypeFloat, true},
		{"invalid type", ParamType("boolean"), false},
		{"empty type", ParamType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.paramType.IsValid(); got != tt.want {
				t.Errorf("ParamType.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
