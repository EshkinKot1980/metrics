package update

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/EshkinKot1980/metrics/internal/common/models"
)

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

	var (
		delta int64
		err   error
	)
	switch metric.MType {
	case models.TypeGauge:
		err = h.updater.PutGauge(metric.ID, *metric.Value)
		metric.Delta = nil
	case models.TypeCounter:
		delta, err = h.updater.PutCounter(metric.ID, *metric.Delta)
		metric.Delta = &delta
		metric.Value = nil
	}

	if err != nil {
		valueToLong := "value too long for type character varying(32)"
		if strings.Contains(err.Error(), valueToLong) {
			http.Error(w, "id is too long, maximum 32 characters", http.StatusBadRequest)
			return
		}

		h.logger.Error("failed to save metric", err)
		http.Error(w, "failed to save metric", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(metric); err != nil {
		h.logger.Error("failed to write response body", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
