package update

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/EshkinKot1980/metrics/internal/server"
)

type Updater interface {
	PutCounter(name string, incriment int64)
	PutGauge(name string, value float64)
}

type UpdateHandler struct {
	updater Updater
}

func New(u Updater) *UpdateHandler {
	return &UpdateHandler{updater: u}
}

func (h *UpdateHandler) Update(res http.ResponseWriter, req *http.Request) {
	const op = "server.handlers.update.UpdateHandler.Update"

	switch m := server.ParsePath(req); m.Mtype {
	case server.TypeGauge:
		v, err := strconv.ParseFloat(m.Value, 64)
		if err != nil {
			// TODO: Отправить ERROR в логер
			err = fmt.Errorf("unexpected error in %s: %w", op, err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		h.updater.PutGauge(m.Name, v)
	case server.TypeCounter:
		v, err := strconv.ParseInt(m.Value, 10, 64)
		if err != nil {
			// TODO: Отправить ERROR в логер
			err = fmt.Errorf("unexpected error in %s: %w", op, err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		h.updater.PutCounter(m.Name, v)
	}
	res.WriteHeader(http.StatusOK)
}
