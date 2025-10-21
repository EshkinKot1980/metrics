// Модуль аудита полученных метрик.
package audit

import (
	"time"

	"github.com/EshkinKot1980/metrics/internal/server/config"
)

// Событие аудита.
type Event struct {
	TS      int64    `json:"ts"`         // unix timestamp события
	Metrics []string `json:"metrics"`    // наименование полученных метрик
	IP      string   `json:"ip_address"` // IP адрес входящего запроса
}

// Логирует ошибки.
type Logger interface {
	Error(message string, err error)
}

// Подписчик на событие.
type Subscrber interface {
	// Обработка события.
	Handle(e Event)
	// Остановка подписчика при завершении приложения.
	Stop()
}

// Auditor, реализует упрощенную версию паттерна "Наблюдатель".
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

// Создает событие аудита.
func (a *Auditor) Rise(e Event) {
	go func() {
		for _, l := range a.listeners {
			l.Handle(e)
		}
	}()
}

// Выключает аудитор при завершении приложения.
func (a *Auditor) Halt() {
	<-time.After(100 * time.Millisecond)
	for _, l := range a.listeners {
		l.Stop()
	}
}
