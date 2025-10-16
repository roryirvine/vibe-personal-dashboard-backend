// Defines parameter types supported in metric queries.
package models

type ParamType string

const (
	ParamTypeString ParamType = "string"
	ParamTypeInt    ParamType = "int"
	ParamTypeFloat  ParamType = "float"
)

func (pt ParamType) IsValid() bool {
	switch pt {
	case ParamTypeString, ParamTypeInt, ParamTypeFloat:
		return true
	}
	return false
}
