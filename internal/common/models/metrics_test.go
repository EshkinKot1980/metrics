package models

import (
	"errors"
	"testing"
)

func TestMetrics_Validate(t *testing.T) {
	delta := int64(13)
	value := 3.1415
	tooLongID := "l"
	for range IDmaxLen {
		tooLongID += "l"
	}

	tests := []struct {
		name   string
		metric Metrics
		err    error
	}{
		{
			name: "positive_counter",
			metric: Metrics{
				ID:    "TestCounter",
				MType: TypeCounter,
				Delta: &delta,
			},
			err: nil,
		},
		{
			name: "negative_counter",
			metric: Metrics{
				ID:    "TestCounter",
				MType: TypeCounter,
			},
			err: errors.New("invalid counter, metric must contain int64 \"delta\" field"),
		},
		{
			name: "positive_gauge",
			metric: Metrics{
				ID:    "TestGauge",
				MType: TypeGauge,
				Value: &value,
			},
			err: nil,
		},
		{
			name: "negative_gauge",
			metric: Metrics{
				ID:    "TestGauge",
				MType: TypeGauge,
			},
			err: errors.New("invalid gauge, metric must contain float64 \"value\" field"),
		},
		{
			name: "negative_metric_type",
			metric: Metrics{
				ID:    "TestUnknown",
				MType: "unknown",
				Delta: &delta,
				Value: &value,
			},
			err: errors.New("invalid metric type"),
		},
		{
			name: "negative_too_long_id",
			metric: Metrics{
				ID:    tooLongID,
				MType: TypeCounter,
				Delta: &delta,
			},
			err: errors.New("id is too long, maximum 32 characters"),
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

func TestMetrics_ValidateType(t *testing.T) {
	tests := []struct {
		name   string
		metric Metrics
		err    error
	}{
		{
			name: "positive_counter",
			metric: Metrics{
				ID:    "TestCounter",
				MType: TypeCounter,
			},
			err: nil,
		},
		{
			name: "positive_gauge",
			metric: Metrics{
				ID:    "TestGauge",
				MType: TypeGauge,
			},
			err: nil,
		},
		{
			name: "negative_metric_type",
			metric: Metrics{
				ID:    "TestUnknown",
				MType: "unknown",
			},
			err: errors.New("invalid metric type"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.metric.ValidateType()

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
