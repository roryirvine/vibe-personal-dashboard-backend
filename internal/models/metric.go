// Defines metric configuration structure with query and parameter definitions.
package models

import "errors"

var (
	ErrMetricNameEmpty  = errors.New("metric name cannot be empty")
	ErrMetricQueryEmpty = errors.New("metric query cannot be empty")
)

type Metric struct {
	Name     string            `toml:"name"`
	Query    string            `toml:"query"`
	MultiRow bool              `toml:"multi_row"`
	Params   []ParamDefinition `toml:"params"`
}

func (m Metric) Validate() error {
	if m.Name == "" {
		return ErrMetricNameEmpty
	}
	if m.Query == "" {
		return ErrMetricQueryEmpty
	}

	// Validate all parameter definitions
	for _, param := range m.Params {
		if err := param.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (m Metric) GetParamByName(name string) (ParamDefinition, bool) {
	for _, param := range m.Params {
		if param.Name == name {
			return param, true
		}
	}
	return ParamDefinition{}, false
}
