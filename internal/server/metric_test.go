package server

import (
	"errors"
	"testing"
)

func TestMetric_Validate(t *testing.T) {
	tests := []struct {
		name   string
		metric Metric
		err    error
	}{
		{
			name: "positive_counter",
			metric: Metric{
				Mtype: TypeCounter,
				Name:  "TestCounter",
				Value: "13",
			},
			err: nil,
		},
		{
			name: "negative_counter",
			metric: Metric{
				Mtype: TypeCounter,
				Name:  "TestCounter",
				Value: "3.1415",
			},
			err: errors.New("invalid metric value, counter must be int64, given: 3.1415"),
		},
		{
			name: "positive_gauge",
			metric: Metric{
				Mtype: TypeGauge,
				Name:  "TestGauge",
				Value: "3.1415",
			},
			err: nil,
		},
		{
			name: "negative_gauge",
			metric: Metric{
				Mtype: TypeGauge,
				Name:  "TestGauge",
				Value: "wtf",
			},
			err: errors.New("invalid metric value, gauge must be float64, given: wtf"),
		},
		{
			name: "negative_metric_type",
			metric: Metric{
				Mtype: "unknown",
				Name:  "TestUnknown",
				Value: "1",
			},
			err: errors.New("invalid metric type: unknown"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.metric.Validate()

			if test.err == nil {
				if err != nil {
					t.Errorf("error expected to be: \"%v\"; got: \"%v\"", nil, err)
				}
			} else if err == nil || test.err.Error() != err.Error() {
				t.Errorf("error expected to be: \"%v\"; got: \"%v\"", test.err, err)
			}
		})
	}
}
