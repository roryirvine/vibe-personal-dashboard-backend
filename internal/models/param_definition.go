// Defines parameter definitions for metric queries with validation.
package models

import "errors"

var (
	ErrParamNameEmpty    = errors.New("parameter name cannot be empty")
	ErrInvalidParamType  = errors.New("parameter type must be string, int, or float")
)

type ParamDefinition struct {
	Name     string    `toml:"name"`
	Type     ParamType `toml:"type"`
	Required bool      `toml:"required"`
}

func (pd ParamDefinition) Validate() error {
	if pd.Name == "" {
		return ErrParamNameEmpty
	}
	if !pd.Type.IsValid() {
		return ErrInvalidParamType
	}
	return nil
}
