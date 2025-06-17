package retrieve

import (
	"fmt"
	"net/http"

	"github.com/EshkinKot1980/metrics/internal/common/models"
	"github.com/EshkinKot1980/metrics/internal/server/storage"
)

type Retriever interface {
	GetCounter(name string) (int64, error)
	GetGauge(name string) (float64, error)
}

type Logger interface {
	Error(message string, err error)
}

type ValueHandler struct {
	retriever Retriever
	logger    Logger
}

func New(r Retriever, l Logger) *ValueHandler {
	return &ValueHandler{retriever: r, logger: l}
}

func (h *ValueHandler) Retrieve(w http.ResponseWriter, r *http.Request) {
	var (
		name    = r.PathValue("name")
		gauge   float64
		counter int64
		err     error
		body    string
	)

	switch t := r.PathValue("type"); t {
	case models.TypeGauge:
		gauge, err = h.retriever.GetGauge(name)
		body = fmt.Sprintf("%v", gauge)
	case models.TypeCounter:
		counter, err = h.retriever.GetCounter(name)
		body = fmt.Sprintf("%v", counter)
	default:
		err = models.ErrInvalidMetricType
	}

	if err != nil {
		switch err {
		case storage.ErrCounterNotFound, storage.ErrGaugeNotFound, models.ErrInvalidMetricType:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			h.logger.Error("failed to retrieve metric", err)
			http.Error(w, "failed to retrieve metric", http.StatusInternalServerError)
		}
		return
	}

	_, err = w.Write([]byte(body))
	if err != nil {
		h.logger.Error("unexpected error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
