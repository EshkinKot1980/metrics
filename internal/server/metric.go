package server

import (
	"fmt"
	"net/http"
	"strconv"
)

const (
	TypeGauge   = "gauge"
	TypeCounter = "counter"
)

type Metric struct {
	Mtype string
	Name  string
	Value string
}

func (m Metric) Validate() error {
	switch m.Mtype {
	case TypeGauge:
		if _, err := strconv.ParseFloat(m.Value, 64); err != nil {
			return fmt.Errorf("invalid metric value, gauge must be float64, given: %s", m.Value)
		}
	case TypeCounter:
		if _, err := strconv.ParseInt(m.Value, 10, 64); err != nil {
			return fmt.Errorf("invalid metric value, counter must be int64, given: %s", m.Value)
		}
	default:
		return fmt.Errorf("invalid metric type: %s", m.Mtype)
	}

	return nil
}

func ParsePath(req *http.Request) Metric {
	return Metric{
		Mtype: req.PathValue("type"),
		Name:  req.PathValue("name"),
		Value: req.PathValue("value"),
	}
}
