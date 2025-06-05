package update

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/EshkinKot1980/metrics/internal/common/models"
)

type Updater interface {
	PutCounter(name string, increment int64) (int64, error)
	PutGauge(name string, value float64) error
}

type Logger interface {
	Error(message string, err error)
}

type UpdateHandler struct {
	updater Updater
	logger  Logger
}

func New(u Updater, l Logger) *UpdateHandler {
	return &UpdateHandler{updater: u, logger: l}
}

func (h *UpdateHandler) Update(w http.ResponseWriter, r *http.Request) {
	var err error
	metric, err := makeMetricFromPath(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch metric.MType {
	case models.TypeGauge:
		err = h.updater.PutGauge(metric.ID, *metric.Value)
	case models.TypeCounter:
		_, err = h.updater.PutCounter(metric.ID, *metric.Delta)
	}

	if err != nil {
		h.logger.Error("failed to save metric", err)
		http.Error(w, "failed to save metric", http.StatusInternalServerError)
		return
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
