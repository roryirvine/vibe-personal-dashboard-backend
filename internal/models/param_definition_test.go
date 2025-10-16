package models

import (
	"testing"
)

func TestParamDefinition_Validate(t *testing.T) {
	tests := []struct {
		name    string
		param   ParamDefinition
		wantErr error
	}{
		{
			name: "valid required string param",
			param: ParamDefinition{
				Name:     "user_id",
				Type:     ParamTypeString,
				Required: true,
			},
			wantErr: nil,
		},
		{
			name: "valid optional int param",
			param: ParamDefinition{
				Name:     "limit",
				Type:     ParamTypeInt,
				Required: false,
			},
			wantErr: nil,
		},
		{
			name: "empty name",
			param: ParamDefinition{
				Name:     "",
				Type:     ParamTypeString,
				Required: true,
			},
			wantErr: ErrParamNameEmpty,
		},
		{
			name: "invalid type",
			param: ParamDefinition{
				Name:     "test",
				Type:     ParamType("boolean"),
				Required: true,
			},
			wantErr: ErrInvalidParamType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.param.Validate()
			if err != tt.wantErr {
				t.Errorf("ParamDefinition.Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
