package service

import (
	"testing"

	"github.com/roryirvine/vibe-personal-dashboard-backend/internal/models"
)

func TestConvertParamValue(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		paramType models.ParamType
		want      interface{}
		wantErr   bool
	}{
		// String conversions
		{
			name:      "string value",
			value:     "hello",
			paramType: models.ParamTypeString,
			want:      "hello",
			wantErr:   false,
		},
		{
			name:      "string with spaces",
			value:     "hello world",
			paramType: models.ParamTypeString,
			want:      "hello world",
			wantErr:   false,
		},
		{
			name:      "empty string",
			value:     "",
			paramType: models.ParamTypeString,
			want:      "",
			wantErr:   false,
		},
		{
			name:      "string with special chars",
			value:     "2025-01-15",
			paramType: models.ParamTypeString,
			want:      "2025-01-15",
			wantErr:   false,
		},

		// Integer conversions
		{
			name:      "positive integer",
			value:     "42",
			paramType: models.ParamTypeInt,
			want:      int64(42),
			wantErr:   false,
		},
		{
			name:      "zero",
			value:     "0",
			paramType: models.ParamTypeInt,
			want:      int64(0),
			wantErr:   false,
		},
		{
			name:      "negative integer",
			value:     "-100",
			paramType: models.ParamTypeInt,
			want:      int64(-100),
			wantErr:   false,
		},
		{
			name:      "large integer",
			value:     "9223372036854775807",
			paramType: models.ParamTypeInt,
			want:      int64(9223372036854775807),
			wantErr:   false,
		},
		{
			name:      "invalid integer - letters",
			value:     "abc",
			paramType: models.ParamTypeInt,
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "invalid integer - float",
			value:     "3.14",
			paramType: models.ParamTypeInt,
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "invalid integer - overflow",
			value:     "99999999999999999999",
			paramType: models.ParamTypeInt,
			want:      nil,
			wantErr:   true,
		},

		// Float conversions
		{
			name:      "positive float",
			value:     "3.14",
			paramType: models.ParamTypeFloat,
			want:      3.14,
			wantErr:   false,
		},
		{
			name:      "zero float",
			value:     "0.0",
			paramType: models.ParamTypeFloat,
			want:      0.0,
			wantErr:   false,
		},
		{
			name:      "negative float",
			value:     "-2.5",
			paramType: models.ParamTypeFloat,
			want:      -2.5,
			wantErr:   false,
		},
		{
			name:      "integer as float",
			value:     "42",
			paramType: models.ParamTypeFloat,
			want:      42.0,
			wantErr:   false,
		},
		{
			name:      "scientific notation",
			value:     "1e10",
			paramType: models.ParamTypeFloat,
			want:      1e10,
			wantErr:   false,
		},
		{
			name:      "invalid float - letters",
			value:     "abc",
			paramType: models.ParamTypeFloat,
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "invalid float - multiple dots",
			value:     "3.14.15",
			paramType: models.ParamTypeFloat,
			want:      nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertParamValue(tt.value, tt.paramType)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertParamValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && got != tt.want {
				t.Errorf("convertParamValue() = %v (type %T), want %v (type %T)", got, got, tt.want, tt.want)
			}
		})
	}
}
