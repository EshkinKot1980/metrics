package client

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/EshkinKot1980/metrics/internal/agent"
)

const (
	TypeGauge   = "gauge"
	TypeCounter = "counter"
	PathPrefix  = "update"
	ContentType = "text/plain"
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
			SetBaseURL(serverAddr+"/"+PathPrefix).
			SetHeader("Content-Type", "text/plain"),
	}
}

func (c *HTTPClient) Report() {
	params := make(map[string]string)
	counters, gauges := c.retriever.Pull()

	params["type"] = TypeCounter
	for _, m := range counters {
		params["name"] = m.Name
		params["value"] = fmt.Sprintf("%v", m.Value)
		c.sendMetric(params)
	}

	params["type"] = TypeGauge
	for _, m := range gauges {
		params["name"] = m.Name
		params["value"] = fmt.Sprintf("%v", m.Value)
		c.sendMetric(params)
	}
}

func (c *HTTPClient) sendMetric(params map[string]string) {
	req := c.client.R().SetPathParams(params)
	resp, err := req.Post("/{type}/{name}/{value}")

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
