package models

import "testing"

func TestMetric_Validate(t *testing.T) {
	tests := []struct {
		name    string
		metric  Metric
		wantErr error
	}{
		{
			name: "valid metric without params",
			metric: Metric{
				Name:     "active_users",
				Query:    "SELECT COUNT(*) FROM users",
				MultiRow: false,
				Params:   nil,
			},
			wantErr: nil,
		},
		{
			name: "valid metric with params",
			metric: Metric{
				Name:     "users_by_date",
				Query:    "SELECT * FROM users WHERE created > ?",
				MultiRow: true,
				Params: []ParamDefinition{
					{Name: "start_date", Type: ParamTypeString, Required: true},
				},
			},
			wantErr: nil,
		},
		{
			name: "empty name",
			metric: Metric{
				Name:  "",
				Query: "SELECT 1",
			},
			wantErr: ErrMetricNameEmpty,
		},
		{
			name: "empty query",
			metric: Metric{
				Name:  "test",
				Query: "",
			},
			wantErr: ErrMetricQueryEmpty,
		},
		{
			name: "invalid param",
			metric: Metric{
				Name:  "test",
				Query: "SELECT 1",
				Params: []ParamDefinition{
					{Name: "", Type: ParamTypeString, Required: true},
				},
			},
			wantErr: ErrParamNameEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.metric.Validate()
			if err != tt.wantErr {
				t.Errorf("Metric.Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestMetric_GetParamByName(t *testing.T) {
	metric := Metric{
		Name:  "test",
		Query: "SELECT * FROM users WHERE id = ? AND status = ?",
		Params: []ParamDefinition{
			{Name: "user_id", Type: ParamTypeInt, Required: true},
			{Name: "status", Type: ParamTypeString, Required: false},
		},
	}

	t.Run("existing param", func(t *testing.T) {
		param, found := metric.GetParamByName("user_id")
		if !found {
			t.Error("expected to find user_id param")
		}
		if param.Name != "user_id" || param.Type != ParamTypeInt {
			t.Errorf("got param %+v, want user_id int param", param)
		}
	})

	t.Run("non-existing param", func(t *testing.T) {
		_, found := metric.GetParamByName("nonexistent")
		if found {
			t.Error("expected not to find nonexistent param")
		}
	})
}
