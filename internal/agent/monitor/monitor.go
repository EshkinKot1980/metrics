package monitor

import (
	"math"
	"math/rand"
	"reflect"
	"runtime"

	"github.com/EshkinKot1980/metrics/internal/agent"
)

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
		agent.Counter{Name: "PollCount", Value: 1},
	}

	m.gauges = make([]agent.Gauge, 0, len(agent.MemStatsFields)+1)

	m.collectMemStats()
	m.gauges = append(
		m.gauges,
		agent.Gauge{Name: "RandomValue", Value: math.MaxFloat32 * rand.Float64()},
	)

	m.updater.Put(m.counters, m.gauges)
}

func (m *Monitor) collectMemStats() {
	var (
		rtm  runtime.MemStats
		gval float64
	)
	runtime.ReadMemStats(&rtm)
	rval := reflect.ValueOf(rtm)

	for _, field := range agent.MemStatsFields {
		ok := true

		switch fv := rval.FieldByName(field); {
		case fv.CanFloat():
			gval = fv.Float()
		case fv.CanUint():
			gval = float64(fv.Uint())
		case fv.CanInt():
			gval = float64(fv.Int())
		default:
			if field == "EnableGC" {
				gval = float64(1)
			} else {
				//TODO: подумать над поведением, возможно ругнуться в лог
				ok = false
			}
		}

		if ok {
			m.gauges = append(m.gauges, agent.Gauge{Name: field, Value: gval})
		}
	}
}
