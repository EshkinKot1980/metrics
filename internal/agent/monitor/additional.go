// Модуль сбора метрик.
package monitor

import (
	"fmt"
	"log"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"

	"github.com/EshkinKot1980/metrics/internal/agent"
)

// Оценочное количество ядер процессоров.
const estimatedCPUcount = 128

// Собирает дополнительные метрики типа датчик при помощи gopsutil:
// "TotalMemory", "FreeMemory", "CPUutilizationN" (N - номер процессора)
type AdditionalMonitor struct {
	updater Updater
	gauges  []agent.Gauge
}

func NewAdditionalMonitor(u Updater) *AdditionalMonitor {
	return &AdditionalMonitor{updater: u}
}

// Сбор метрик.
func (m *AdditionalMonitor) Poll() {
	m.gauges = make([]agent.Gauge, 0, estimatedCPUcount+2)
	m.collectMemUsage()
	m.collectCPUutilization()

	m.updater.Put([]agent.Counter{}, m.gauges)
}

func (m *AdditionalMonitor) collectMemUsage() {
	vm, err := mem.VirtualMemory()
	if err != nil {
		log.Print("can`t get virtual memory: ", err)
		return
	}

	m.gauges = append(m.gauges, agent.Gauge{Name: "TotalMemory", Value: float64(vm.Total)})
	m.gauges = append(m.gauges, agent.Gauge{Name: "FreeMemory", Value: float64(vm.Free)})
}

func (m *AdditionalMonitor) collectCPUutilization() {
	stats, err := cpu.Percent(0, true)
	if err != nil {
		log.Print("can`t get spu ussage: ", err)
		return
	}

	for i, v := range stats {
		gauge := agent.Gauge{
			Name:  fmt.Sprintf("CPUutilization%d", i),
			Value: v,
		}
		m.gauges = append(m.gauges, gauge)
	}
}
