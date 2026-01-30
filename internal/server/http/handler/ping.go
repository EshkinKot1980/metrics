// Модуль handler реализует обработчики http запросов.
package handler

import (
	"net/http"
)

// Проверяет доступность базы данных.
type DBPinger interface {
	Ping() bool
}

// Обработчик доступности БД.
type PingHandler struct {
	pinger DBPinger
}

func NewPingHandler(p DBPinger) *PingHandler {
	return &PingHandler{pinger: p}
}

// Возвращает http.StatusOK в случае доступности БД,
// http.StatusServiceUnavailable в противном случае.
func (h *PingHandler) Ping(w http.ResponseWriter, r *http.Request) {
	if !h.pinger.Ping() {
		http.Error(w, "", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
}
