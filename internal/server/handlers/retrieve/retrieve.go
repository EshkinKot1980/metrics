package retrieve

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/EshkinKot1980/metrics/internal/common/models"
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

func (h *ValueHandler) Retrieve(res http.ResponseWriter, req *http.Request) {
	var (
		name    = req.PathValue("name")
		gauge   float64
		counter int64
		err     error
		body    string
	)

	switch t := req.PathValue("type"); t {
	case models.TypeGauge:
		gauge, err = h.retriever.GetGauge(name)
		body = fmt.Sprintf("%v", gauge)
	case models.TypeCounter:
		counter, err = h.retriever.GetCounter(name)
		body = fmt.Sprintf("%v", counter)
	default:
		err = errors.New("invalid metric type: " + t)
	}

	if err != nil {
		http.Error(res, err.Error(), http.StatusNotFound)
		return
	}

	_, err = res.Write([]byte(body))
	if err != nil {
		h.logger.Error("unexpected error", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}
