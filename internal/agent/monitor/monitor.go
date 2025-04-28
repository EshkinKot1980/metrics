package monitor

import (
	"runtime"
	"time"
	"reflect"
	"math"
	"math/rand"

	"github.com/EshkinKot1980/metrics/internal/agent/model"
)

type Storage interface {
	PutCounter(model.Counter)
	PutGauge(model.Gauge)
	PutLock()
	PutUnlock()
}

var (
	gauges   map[string]float64
	counters map[string]int64
	storage Storage
)


func Run(interval int, s Storage) {
	storage = s
	var i = time.Duration(interval) * time.Second
	
	for {
		<-time.After(i)
		collectMetrics()
		storeMetrics()
	}
}

func collectMetrics() {
	//collect gauges
	gauges = make(map[string]float64)
	collectMemStats()
	gauges["RandomValue"] = math.MaxFloat32 * rand.Float64()
	//collect counters
	counters = make(map[string]int64)
	counters["PollCount"] = 1
}

func storeMetrics() {
	storage.PutLock()
	defer storage.PutUnlock()

	for n, v := range gauges {
		storage.PutGauge(model.Gauge{Name: n, Value: v})
	}

	for n, v := range counters {
		storage.PutCounter(model.Counter{Name: n, Value: v})
	} 
}

func collectMemStats() {
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)
	val := reflect.ValueOf(rtm)

	for _, field := range model.MemStatsFields {
		switch fv := val.FieldByName(field); {
		case fv.CanFloat():
			gauges[field] = fv.Float()
		case fv.CanUint():
			gauges[field] = float64(fv.Uint())				
		case fv.CanInt():
			gauges[field] = float64(fv.Int())			
		default:
			//TODO подумать над поведением в случае esle
			if field == "EnableGC" {
				gauges[field] = float64(1)
			}
		}
	}
}
