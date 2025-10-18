package audit

import (
	"time"

	"github.com/EshkinKot1980/metrics/internal/server/config"
)

type Event struct {
	TS      int64    `json:"ts"`
	Metrics []string `json:"metrics"`
	IP      string   `json:"ip_address"`
}

type Logger interface {
	Error(message string, err error)
}

type Subscrber interface {
	Handle(e Event)
	Stop()
}

type Auditor struct {
	listeners []Subscrber
}

func NewAuditor(cfg *config.Config, l Logger) *Auditor {
	a := Auditor{}

	if cfg.AuditFile != "" {
		f, err := NewFileAuditor(cfg.AuditFile, l)
		if err != nil {
			l.Error("failed to create file auditor", err)
		} else {
			a.listeners = append(a.listeners, f)
		}
	}

	if cfg.AuditURL != "" {
		a.listeners = append(a.listeners, NewURLauditor(cfg.AuditURL, l))
	}

	return &a
}

func (a *Auditor) Rise(e Event) {
	go func() {
		for _, l := range a.listeners {
			l.Handle(e)
		}
	}()
}

func (a *Auditor) Halt() {
	<-time.After(100 * time.Millisecond)
	for _, l := range a.listeners {
		l.Stop()
	}
}
