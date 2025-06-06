package client

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/EshkinKot1980/metrics/internal/agent"
	"github.com/EshkinKot1980/metrics/internal/common/models"
)

const (
	Path        = "/updates"
	ContentType = "application/json"
)

type Storage interface {
	Pull() ([]agent.Counter, []agent.Gauge)
	Put(c []agent.Counter, g []agent.Gauge)
}

type HTTPClient struct {
	storage Storage
	address string
	client  *resty.Client
	mx      sync.Mutex
}

func New(s Storage, serverAddr string) *HTTPClient {
	return &HTTPClient{
		storage: s,
		address: serverAddr,
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
	if !c.mx.TryLock() {
		return
	}
	defer c.mx.Unlock()

	var metric models.Metrics
	counters, gauges := c.storage.Pull()
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

	if len(metrics) == 0 {
		return
	}

	if !c.sendMetric(metrics) {
		c.storage.Put(counters, []agent.Gauge{})
	}
}

func (c *HTTPClient) sendMetric(metric []models.Metrics) bool {
	retries := []int{1, 3, 5}
	i := 0
	for {
		succes, retry := true, false
		req := c.client.R().SetBody(metric)
		resp, err := req.Post(Path)

		if err != nil {
			log.Print(err)
			succes, retry = false, true
		} else if !resp.IsSuccess() {
			log.Print("Code: ", resp.StatusCode(), " Body: ", resp)
			succes = false

			if resp.StatusCode() == 500 {
				retry = true
			}
		}

		if !retry || i >= len(retries) {
			return succes
		}

		interval := time.Duration(retries[i]) * time.Second
		<-time.After(interval)
		i++
	}

}

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
