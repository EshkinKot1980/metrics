package monitor

import (
	"math"
	"math/rand"
	"reflect"
	"runtime"

	"github.com/EshkinKot1980/metrics/internal/agent/model"
)

type Storage interface {
	Put(c []model.Counter, g []model.Gauge)
}

type Monitor struct {
	storage  Storage
	gauges   []model.Gauge
	counters []model.Counter
}

func New(s Storage) *Monitor {
	return &Monitor{storage: s}
}

func (m *Monitor) Poll() {
	m.counters = []model.Counter{
		model.Counter{Name: "PollCount", Value: 1},
	}

	m.gauges = make([]model.Gauge, 0, len(model.MemStatsFields)+1)

	m.collectMemStats()
	m.gauges = append(
		m.gauges,
		model.Gauge{Name: "RandomValue", Value: math.MaxFloat32 * rand.Float64()},
	)

	m.storage.Put(m.counters, m.gauges)
}

func (m *Monitor) collectMemStats() {
	var (
		rtm  runtime.MemStats
		gval float64
	)
	runtime.ReadMemStats(&rtm)
	rval := reflect.ValueOf(rtm)

	for _, field := range model.MemStatsFields {
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
			m.gauges = append(m.gauges, model.Gauge{Name: field, Value: gval})
		}
	}
}
