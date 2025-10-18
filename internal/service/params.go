// Converts URL query parameters to typed values for database queries.
package service

import (
	"fmt"
	"strconv"

	"github.com/roryirvine/vibe-personal-dashboard-backend/internal/models"
)

// convertParamValue converts a string parameter value to the specified type.
// Returns interface{} containing int64, float64, or string depending on paramType.
// Returns an error if the conversion fails.
func convertParamValue(value string, paramType models.ParamType) (interface{}, error) {
	switch paramType {
	case models.ParamTypeString:
		return value, nil

	case models.ParamTypeInt:
		n, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid integer value %q: %w", value, err)
		}
		return n, nil

	case models.ParamTypeFloat:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid float value %q: %w", value, err)
		}
		return f, nil

	default:
		return nil, fmt.Errorf("unsupported parameter type: %s", paramType)
	}
}
