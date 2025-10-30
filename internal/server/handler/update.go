// Модуль handler реализует обработчики http запросов.
package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/EshkinKot1980/metrics/internal/common/models"
	"github.com/EshkinKot1980/metrics/internal/server/service"
)

// Сервис сохранения метрик.
type UpdateService interface {
	// Сохраняет метрику, возвращает её обновленное состояние.
	Put(metric models.Metrics) (models.Metrics, error)
	// Сохраняет множество метрик.
	PutList(ctx context.Context, metrics []models.Metrics) error
}

// Обработчик для сохранения метрик.
type UpdateHandler struct {
	service UpdateService
	logger  Logger
}

func NewUpdateHandler(s UpdateService, l Logger) *UpdateHandler {
	return &UpdateHandler{service: s, logger: l}
}

// Обновляет метрику из параметров пути GET запроса.
func (h *UpdateHandler) UpdateFromPath(w http.ResponseWriter, r *http.Request) {
	metric, err := models.MakeMetrics(
		r.PathValue("name"),
		r.PathValue("type"),
		r.PathValue("value"),
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = h.service.Put(metric)
	if err != nil {
		msg := http.StatusText(http.StatusInternalServerError)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Обновляет метрику из тела POST запроса в формате JSON.
// В случае успеха возвращает её обновленное значение.
func (h *UpdateHandler) Update(w http.ResponseWriter, r *http.Request) {
	var metric models.Metrics

	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := metric.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metric, err := h.service.Put(metric)
	if err != nil {
		msg := http.StatusText(http.StatusInternalServerError)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(metric); err != nil {
		h.logger.Error("failed to write response body", err)
	}
}

// Обновляет множество метрик из тела POST запроса в формате JSON.
func (h *UpdateHandler) UpdateList(w http.ResponseWriter, r *http.Request) {
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

	addr := strings.Split(r.RemoteAddr, ":")
	ctx := context.WithValue(r.Context(), service.KeyClientIP, addr[0])
	err := h.service.PutList(ctx, metrics)
	if err != nil {
		msg := http.StatusText(http.StatusInternalServerError)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
