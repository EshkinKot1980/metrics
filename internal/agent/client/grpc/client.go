package grpc

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/EshkinKot1980/metrics/internal/agent"
	pb "github.com/EshkinKot1980/metrics/internal/common/proto"
)

// Хранилище метрик.
type Storage interface {
	// Забирает метрики из хранилища.
	Pull() ([]agent.Counter, []agent.Gauge)
	// Помещает метрики в хранилище.
	Put(c []agent.Counter, g []agent.Gauge)
}

// Клиент отправляющий метрики на grpc сервер одним запросом.
type Client struct {
	storage Storage
	address string
	agentIP string
}

func NewClient(s Storage, addr string, ip string) *Client {
	return &Client{storage: s, address: addr, agentIP: ip}
}

// Оправляет метрики на сервер.
func (c *Client) Report() {
	counters, gauges := c.storage.Pull()
	metrics := make([]*pb.Metric, 0, len(counters)+len(gauges))

	for _, m := range counters {
		pbm := pb.Metric{}
		pbm.SetId(m.Name)
		pbm.SetType(pb.Metric_COUNTER)
		pbm.SetDelta(m.Value)
		metrics = append(metrics, &pbm)
	}

	for _, m := range gauges {
		pbm := pb.Metric{}
		pbm.SetId(m.Name)
		pbm.SetType(pb.Metric_GAUGE)
		pbm.SetValue(m.Value)
		metrics = append(metrics, &pbm)
	}

	if len(metrics) == 0 {
		return
	}

	if !c.sendMetrics(metrics) {
		c.storage.Put(counters, []agent.Gauge{})
	}
}

func (c *Client) sendMetrics(metrics []*pb.Metric) bool {
	conn, err := grpc.NewClient(c.address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Print(err)
		return false
	}
	defer conn.Close()

	client := pb.NewMetricsClient(conn)
	req := pb.UpdateMetricsRequest_builder{
		Metrics: metrics,
	}.Build()

	md := metadata.Pairs("x-real-ip", c.agentIP)
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	_, err = client.UpdateMetrics(ctx, req)
	if err != nil {
		log.Printf("GRPC: %s,  Code: %d, Error: %s \n", c.address, status.Code(err), err.Error())
		return false
	}

	return true
}
