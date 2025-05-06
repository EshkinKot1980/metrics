package update

import (
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
	switch m := server.ParsePath(req); m.Mtype {
	case server.TypeGauge:
		v, err := strconv.ParseFloat(m.Value, 64)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		h.updater.PutGauge(m.Name, v)
	case server.TypeCounter:
		v, err := strconv.ParseInt(m.Value, 10, 64)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		h.updater.PutCounter(m.Name, v)
	}
	res.WriteHeader(http.StatusOK)
}
