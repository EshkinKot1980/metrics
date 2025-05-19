package client

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/EshkinKot1980/metrics/internal/agent"
	"github.com/EshkinKot1980/metrics/internal/common/models"
)

const (
	Path        = "/update"
	ContentType = "application/json"
)

type Retriever interface {
	Pull() ([]agent.Counter, []agent.Gauge)
}

type HTTPClient struct {
	retriever Retriever
	address   string
	client    *resty.Client
}

func New(r Retriever, serverAddr string) *HTTPClient {
	return &HTTPClient{
		retriever: r,
		address:   serverAddr,
		client: resty.New().
			SetTimeout(time.Duration(1)*time.Second).
			SetBaseURL(serverAddr).
			SetHeader("Content-Type", ContentType),
	}
}

func (c *HTTPClient) Report() {
	// params := make(map[string]string)
	var metric models.Metrics
	counters, gauges := c.retriever.Pull()

	metric.MType = models.TypeCounter
	for _, m := range counters {
		metric.ID = m.Name
		metric.Delta = &m.Value
		c.sendMetric(metric)
	}

	metric.MType = models.TypeGauge
	metric.Delta = nil
	for _, m := range gauges {
		metric.ID = m.Name
		metric.Value = &m.Value
		c.sendMetric(metric)
	}
}

func (c *HTTPClient) sendMetric(metric models.Metrics) {
	req := c.client.R().SetBody(metric)
	resp, err := req.Post(Path)

	if err != nil {
		//TODO: заменить на логер
		fmt.Println(err.Error())
		return
	}

	if !resp.IsSuccess() {
		//TODO: заменить на логер
		fmt.Println(req.URL)
		fmt.Println("Code:", resp.StatusCode(), "Body:", resp)
	}
}
