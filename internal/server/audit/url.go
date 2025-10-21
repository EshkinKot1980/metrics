// Модуль аудита полученных метрик.
package audit

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

// Отправляет данные аудита на URL методом POST, реализует интерфейс Subscrber.
type URLauditor struct {
	url    string
	logger Logger
	client *resty.Client
}

func NewURLauditor(url string, l Logger) *URLauditor {
	return &URLauditor{
		url:    url,
		logger: l,
		client: resty.New().
			SetTimeout(time.Second).
			SetHeader("Content-Type", "application/json"),
	}
}

func (a *URLauditor) Stop() {}

func (a *URLauditor) Handle(e Event) {
	req := a.client.R().SetBody(e)
	resp, err := req.Post(a.url)

	if err != nil {
		a.logger.Error("failed to send audit event", err)
		return
	}

	if !resp.IsSuccess() {
		err := fmt.Errorf("POST %s Code: %d Body: %s", a.url, resp.StatusCode(), string(resp.Body()))
		a.logger.Error("failed to send audit event", err)
	}
}
