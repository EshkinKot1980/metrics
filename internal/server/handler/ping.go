package handler

import (
	"net/http"
)

type DBPinger interface {
	Ping() bool
}

type PingHandler struct {
	pinger DBPinger
}

func NewPingHandler(p DBPinger) *PingHandler {
	return &PingHandler{pinger: p}
}

func (h *PingHandler) Ping(w http.ResponseWriter, r *http.Request) {
	if !h.pinger.Ping() {
		http.Error(w, "", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
}
