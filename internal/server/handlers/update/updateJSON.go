package update

import (
	"encoding/json"
	"net/http"

	"github.com/EshkinKot1980/metrics/internal/common/models"
)

type Logger interface {
	Error(message string, err error)
}

type UpdateJSONHandler struct {
	updater Updater
	logger  Logger
}

func NewJSONHandler(u Updater, l Logger) *UpdateJSONHandler {
	return &UpdateJSONHandler{updater: u, logger: l}
}

func (h *UpdateJSONHandler) Update(w http.ResponseWriter, r *http.Request) {
	var metric models.Metrics

	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := metric.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch metric.MType {
	case models.TypeGauge:
		h.updater.PutGauge(metric.ID, *metric.Value)
		metric.Delta = nil
	case models.TypeCounter:
		delta := h.updater.PutCounter(metric.ID, *metric.Delta)
		metric.Delta = &delta
		metric.Value = nil
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(metric); err != nil {
		h.logger.Error("failed to write response body", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
