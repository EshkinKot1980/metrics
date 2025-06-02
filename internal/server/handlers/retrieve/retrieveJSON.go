package retrieve

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/EshkinKot1980/metrics/internal/common/models"
)

type ValueJSONHandler struct {
	retriever Retriever
	logger    Logger
}

func NewJSONHandler(r Retriever, l Logger) *ValueJSONHandler {
	return &ValueJSONHandler{retriever: r, logger: l}
}

func (h *ValueJSONHandler) Retrieve(w http.ResponseWriter, r *http.Request) {
	var (
		metric models.Metrics
		value  float64
		delta  int64
		err    error
	)

	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch metric.MType {
	case models.TypeGauge:
		value, err = h.retriever.GetGauge(metric.ID)
		metric.Value = &value
		metric.Delta = nil
	case models.TypeCounter:
		delta, err = h.retriever.GetCounter(metric.ID)
		metric.Delta = &delta
		metric.Value = nil
	default:
		err = errors.New("invalid metric type: " + metric.MType)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(metric); err != nil {
		h.logger.Error("failed to write response body", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
