package client

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/EshkinKot1980/metrics/internal/agent"
	"github.com/EshkinKot1980/metrics/internal/common/models"
)

const (
	Path        = "/updates"
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

// TODO: выпилить needCompress после того как доделаю тесты
func New(r Retriever, serverAddr string) *HTTPClient {
	return &HTTPClient{
		retriever: r,
		address:   serverAddr,
		client: resty.New().
			SetTimeout(time.Duration(1)*time.Second).
			SetBaseURL(serverAddr).
			SetHeader("Content-Type", ContentType).
			SetHeader("Accept-Encoding", "gzip").
			SetHeader("Content-Encoding", "gzip").
			OnBeforeRequest(gzipWrapper),
	}
}

func (c *HTTPClient) Report() {
	var metric models.Metrics
	counters, gauges := c.retriever.Pull()
	metrics := make([]models.Metrics, 0, len(counters)+len(gauges))

	metric.MType = models.TypeCounter
	for _, m := range counters {
		metric.ID = m.Name
		metric.Delta = &m.Value
		metrics = append(metrics, metric)
	}

	metric.MType = models.TypeGauge
	metric.Delta = nil
	for _, m := range gauges {
		metric.ID = m.Name
		metric.Value = &m.Value
		metrics = append(metrics, metric)
	}

	c.sendMetric(metrics)
}

func (c *HTTPClient) sendMetric(metric []models.Metrics) {
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

// after countless frustrations and tears i suddenly found the way
func gzipWrapper(c *resty.Client, r *resty.Request) error {
	var body bytes.Buffer

	bodyJSON, err := json.Marshal(r.Body)
	if err != nil {
		return err
	}

	g := gzip.NewWriter(&body)
	if _, err := g.Write(bodyJSON); err != nil {
		return err
	}
	if err := g.Close(); err != nil {
		return err
	}

	r.SetBody(&body)
	return nil
}
