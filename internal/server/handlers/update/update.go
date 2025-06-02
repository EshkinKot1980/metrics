package update

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/EshkinKot1980/metrics/internal/common/models"
)

type Updater interface {
	PutCounter(name string, increment int64) int64
	PutGauge(name string, value float64)
}

type UpdateHandler struct {
	updater Updater
	logger  Logger
}

func New(u Updater) *UpdateHandler {
	return &UpdateHandler{updater: u}
}

func (h *UpdateHandler) Update(w http.ResponseWriter, r *http.Request) {
	metric, err := makeMetricFromPath(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch metric.MType {
	case models.TypeGauge:
		h.updater.PutGauge(metric.ID, *metric.Value)
	case models.TypeCounter:
		h.updater.PutCounter(metric.ID, *metric.Delta)
	}

	w.WriteHeader(http.StatusOK)
}

func makeMetricFromPath(r *http.Request) (models.Metrics, error) {
	value := r.PathValue("value")
	metric := models.Metrics{
		ID:    r.PathValue("name"),
		MType: r.PathValue("type"),
	}

	switch metric.MType {
	case models.TypeGauge:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return metric, fmt.Errorf("invalid metric value, gauge must be float64, given: %s", value)
		}
		metric.Value = &v
	case models.TypeCounter:
		d, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return metric, fmt.Errorf("invalid metric value, counter must be int64, given: %s", value)
		}
		metric.Delta = &d
	default:
		return metric, fmt.Errorf("invalid metric type: %s", metric.MType)
	}

	return metric, nil
}
