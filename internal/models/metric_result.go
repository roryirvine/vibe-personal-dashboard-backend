// Defines the API response structure for metric results.
package models

type MetricResult struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}
