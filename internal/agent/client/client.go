package client

import (
	"io"
	"strings"
	"fmt"
	"net/http"
	"time"
	"github.com/EshkinKot1980/metrics/internal/agent/model"
)

const (
	TypeGauge = "gauge"
	TypeCounter = "counter"
	ContentType = "text"
)

type Storage interface {
	Pull() ([]model.Counter, []model.Gauge)
}

var (
	storage Storage
	client *http.Client
	address = "http://localhost:8080"
)

func Run(interval int, s Storage) {
	var i = time.Duration(interval) * time.Second
	storage = s	
	client = &http.Client{
		Timeout: time.Second * 1,
	}

	for {
		<-time.After(i)
		counters, gauges := s.Pull()

		for _, c := range counters {
			url := fmt.Sprintf("%s/update/%s/%s/%v",address, TypeCounter, c.Name, c.Value)
			sendMetric(url)
		}

		for _, g := range gauges {
			url := fmt.Sprintf("%s/update/%s/%s/%v",address, TypeGauge, g.Name, g.Value)
			sendMetric(url)
		}
	}
}

func sendMetric(url string) {
	resp, err := client.Post(url, "text/plain", nil)
		if err !=nil {
			//TODO: заменить на логер
			fmt.Println(err.Error())
		}
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			//TODO: заменить на логер
			fmt.Printf(
				"server error: {url:%s ,code: %v, body: %s}\n",
				url,
				resp.StatusCode,
				strings.TrimSuffix(string(body), "\n"),
			)
		}
		resp.Body.Close()
}
