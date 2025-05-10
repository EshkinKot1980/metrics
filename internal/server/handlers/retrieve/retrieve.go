package retrieve

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/EshkinKot1980/metrics/internal/server"
)

type Retriever interface {
	GetCounter(name string) (int64, error)
	GetGauge(name string) (float64, error)
}

type ValueHandler struct {
	retriever Retriever
}

func New(r Retriever) *ValueHandler {
	return &ValueHandler{retriever: r}
}

func (h *ValueHandler) Retrieve(res http.ResponseWriter, req *http.Request) {
	const op = "server.handlers.update.ValueHandler.Retrieve"
	var (
		name    = req.PathValue("name")
		gauge   float64
		counter int64
		err     error
		body    string
	)

	switch t := req.PathValue("type"); t {
	case server.TypeGauge:
		gauge, err = h.retriever.GetGauge(name)
		body = fmt.Sprintf("%v", gauge)
	case server.TypeCounter:
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
		// TODO: Отправить ERROR в логер
		err = fmt.Errorf("unexpected error in %s: %w", op, err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}
