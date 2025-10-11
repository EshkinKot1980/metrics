package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/EshkinKot1980/metrics/internal/common/models"
	"github.com/EshkinKot1980/metrics/internal/server/service"
)

type RetrieveService interface {
	Fill(metric models.Metrics) (models.Metrics, error)
}

type Logger interface {
	Error(message string, err error)
}

type RetrieveHandler struct {
	service RetrieveService
	logger  Logger
}

func NewRetrieveHandler(s RetrieveService, l Logger) *RetrieveHandler {
	return &RetrieveHandler{service: s, logger: l}
}

func (h *RetrieveHandler) GetByPath(w http.ResponseWriter, r *http.Request) {
	metric := models.Metrics{
		ID:    r.PathValue("name"),
		MType: r.PathValue("type"),
	}

	if err := metric.ValidateType(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metric, err := h.service.Fill(metric)
	if err != nil {
		if errors.Is(err, service.ErrMetricNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(
				w,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
			)
		}
		return
	}

	var body string
	switch metric.MType {
	case models.TypeCounter:
		body = fmt.Sprintf("%v", *metric.Delta)
	case models.TypeGauge:
		body = fmt.Sprintf("%v", *metric.Value)
	}

	_, err = w.Write([]byte(body))
	if err != nil {
		h.logger.Error("failed to write body", err)
	}
}

func (h *RetrieveHandler) GetJSON(w http.ResponseWriter, r *http.Request) {
	var metric models.Metrics

	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := metric.ValidateType(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metric, err := h.service.Fill(metric)
	if err != nil {
		if errors.Is(err, service.ErrMetricNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(metric); err != nil {
		h.logger.Error("failed to write response body", err)
	}
}
