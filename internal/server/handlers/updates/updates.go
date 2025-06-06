package updates

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/EshkinKot1980/metrics/internal/common/models"
)

type Updater interface {
	PutMetrics(ctx context.Context, metrics []models.Metrics) error
}

type Logger interface {
	Error(message string, err error)
}

type BatchUpdateHandler struct {
	updater Updater
	logger  Logger
}

func New(u Updater, l Logger) *BatchUpdateHandler {
	return &BatchUpdateHandler{updater: u, logger: l}
}

func (h *BatchUpdateHandler) Update(w http.ResponseWriter, r *http.Request) {
	var metrics []models.Metrics

	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	for _, m := range metrics {
		if err := m.Validate(); err != nil {
			msg := err.Error() + " {id: " + m.ID + ", type: " + m.MType + "}"
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
	}

	err := h.updater.PutMetrics(r.Context(), metrics)
	if err != nil {
		// TODO: подумать над вынесением этой проверки в models.Metrics
		valueToLong := "value too long for type character varying(32)"
		if strings.Contains(err.Error(), valueToLong) {
			http.Error(w, "id is too long, maximum 32 characters", http.StatusBadRequest)
			return
		}

		h.logger.Error("failed to save metrics", err)
		http.Error(w, "failed to save metrics", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
