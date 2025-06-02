package ping

import (
	"database/sql"
	"net/http"
)

type PingHandler struct {
	db *sql.DB
}

func New(db *sql.DB) *PingHandler {
	return &PingHandler{db: db}
}

func (h *PingHandler) Ping(w http.ResponseWriter, r *http.Request) {
	if err := h.db.Ping(); err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
