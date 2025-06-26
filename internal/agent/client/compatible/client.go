// Достал из гита клиент, отсылающий запросы на /update, чтобы было куда воткнуть Worker Pool
package compatible

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"log"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/EshkinKot1980/metrics/internal/agent"
	"github.com/EshkinKot1980/metrics/internal/common/models"
)

const (
	Path                  = "/update"
	ContentType           = "application/json"
	estimatedMetricsCount = 256
)

type Retriever interface {
	Pull() ([]agent.Counter, []agent.Gauge)
}

type HTTPClient struct {
	retriever Retriever
	address   string
	queue     chan models.Metrics
	client    *resty.Client
}

func New(r Retriever, serverAddr string, requestsLimit uint64) *HTTPClient {
	c := &HTTPClient{
		retriever: r,
		address:   serverAddr,
		queue:     make(chan models.Metrics, estimatedMetricsCount),
		client: resty.New().
			SetTimeout(time.Duration(1)*time.Second).
			SetBaseURL(serverAddr).
			SetHeader("Content-Type", ContentType).
			SetHeader("Accept-Encoding", "gzip").
			OnBeforeRequest(gzipWrapper),
	}

	c.makeWorkers(requestsLimit)

	return c
}

func (c *HTTPClient) makeWorkers(count uint64) {
	for i := uint64(1); ; i++ {
		go c.sendMetric()

		if i >= count {
			return
		}
	}
}

func (c *HTTPClient) Report() {
	var metric models.Metrics
	counters, gauges := c.retriever.Pull()

	metric.MType = models.TypeCounter
	for _, m := range counters {
		metric.ID = m.Name
		metric.Delta = &m.Value
		c.queue <- metric
	}

	metric.MType = models.TypeGauge
	metric.Delta = nil
	for _, m := range gauges {
		metric.ID = m.Name
		metric.Value = &m.Value
		c.queue <- metric
	}
}

func (c *HTTPClient) sendMetric() {
	for {
		metric := <-c.queue
		req := c.client.R().SetBody(metric)
		resp, err := req.Post(Path)

		if err != nil {
			log.Print(err)
		} else if !resp.IsSuccess() {
			log.Print("POST", c.address, Path, " Code: ", resp.StatusCode(), " Body: ", resp)
		}
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

	r.SetHeader("Content-Encoding", "gzip")
	r.SetBody(&body)
	return nil
}
