package monitor

import (
	"math"
	"math/rand"
	"runtime"

	"github.com/EshkinKot1980/metrics/internal/agent"
)

const MemStatsFieldsCount = 27

type Updater interface {
	Put(c []agent.Counter, g []agent.Gauge)
}

type Monitor struct {
	updater  Updater
	gauges   []agent.Gauge
	counters []agent.Counter
}

func New(u Updater) *Monitor {
	return &Monitor{updater: u}
}

func (m *Monitor) Poll() {
	m.counters = []agent.Counter{
		{Name: "PollCount", Value: 1},
	}

	m.gauges = make([]agent.Gauge, 0, MemStatsFieldsCount+1)

	m.collectMemStats()
	m.gauges = append(
		m.gauges,
		agent.Gauge{Name: "RandomValue", Value: math.MaxFloat32 * rand.Float64()},
	)

	m.updater.Put(m.counters, m.gauges)
}

func (m *Monitor) collectMemStats() {
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	m.gauges = append(m.gauges, agent.Gauge{Name: "Alloc", Value: float64(rtm.Alloc)})
	m.gauges = append(m.gauges, agent.Gauge{Name: "BuckHashSys", Value: float64(rtm.BuckHashSys)})
	m.gauges = append(m.gauges, agent.Gauge{Name: "Frees", Value: float64(rtm.Frees)})
	m.gauges = append(m.gauges, agent.Gauge{Name: "GCCPUFraction", Value: rtm.GCCPUFraction})
	m.gauges = append(m.gauges, agent.Gauge{Name: "GCSys", Value: float64(rtm.GCSys)})
	m.gauges = append(m.gauges, agent.Gauge{Name: "HeapAlloc", Value: float64(rtm.HeapAlloc)})
	m.gauges = append(m.gauges, agent.Gauge{Name: "HeapIdle", Value: float64(rtm.HeapIdle)})
	m.gauges = append(m.gauges, agent.Gauge{Name: "HeapInuse", Value: float64(rtm.HeapInuse)})
	m.gauges = append(m.gauges, agent.Gauge{Name: "HeapObjects", Value: float64(rtm.HeapObjects)})
	m.gauges = append(m.gauges, agent.Gauge{Name: "HeapReleased", Value: float64(rtm.HeapReleased)})
	m.gauges = append(m.gauges, agent.Gauge{Name: "HeapSys", Value: float64(rtm.HeapSys)})
	m.gauges = append(m.gauges, agent.Gauge{Name: "LastGC", Value: float64(rtm.LastGC)})
	m.gauges = append(m.gauges, agent.Gauge{Name: "Lookups", Value: float64(rtm.Lookups)})
	m.gauges = append(m.gauges, agent.Gauge{Name: "MCacheInuse", Value: float64(rtm.MCacheInuse)})
	m.gauges = append(m.gauges, agent.Gauge{Name: "MCacheSys", Value: float64(rtm.MCacheSys)})
	m.gauges = append(m.gauges, agent.Gauge{Name: "MSpanInuse", Value: float64(rtm.MSpanInuse)})
	m.gauges = append(m.gauges, agent.Gauge{Name: "MSpanSys", Value: float64(rtm.MSpanSys)})
	m.gauges = append(m.gauges, agent.Gauge{Name: "Mallocs", Value: float64(rtm.Mallocs)})
	m.gauges = append(m.gauges, agent.Gauge{Name: "NextGC", Value: float64(rtm.NextGC)})
	m.gauges = append(m.gauges, agent.Gauge{Name: "NumForcedGC", Value: float64(rtm.NumForcedGC)})
	m.gauges = append(m.gauges, agent.Gauge{Name: "NumGC", Value: float64(rtm.NumGC)})
	m.gauges = append(m.gauges, agent.Gauge{Name: "OtherSys", Value: float64(rtm.OtherSys)})
	m.gauges = append(m.gauges, agent.Gauge{Name: "PauseTotalNs", Value: float64(rtm.PauseTotalNs)})
	m.gauges = append(m.gauges, agent.Gauge{Name: "StackInuse", Value: float64(rtm.StackInuse)})
	m.gauges = append(m.gauges, agent.Gauge{Name: "StackSys", Value: float64(rtm.StackSys)})
	m.gauges = append(m.gauges, agent.Gauge{Name: "Sys", Value: float64(rtm.Sys)})
	m.gauges = append(m.gauges, agent.Gauge{Name: "TotalAlloc", Value: float64(rtm.TotalAlloc)})
}
