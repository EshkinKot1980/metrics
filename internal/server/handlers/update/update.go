package update

import (
	"fmt"
	"net/http"
	"strconv"
)

const (
	TypeGauge   = "gauge"
	TypeCounter = "counter"
)

type Storage interface {
	PutCounter(name string, incriment int64)
	PutGauge(name string, value float64)
}

type metric struct {
	mtype string
	name  string
	value string
}

var storage Storage

func New(s Storage) http.HandlerFunc {
	storage = s

	return validateData(http.HandlerFunc(saveMetric))
}

// TODO: выяснить где принято эти валидаторы хранить и вынести в отдельный слой
func validateData(next http.HandlerFunc) http.HandlerFunc {
	fn := func(res http.ResponseWriter, req *http.Request) {
		m := parsePath(req)
		if err := m.validate(); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
		} else {
			next.ServeHTTP(res, req)
		}
	}

	return http.HandlerFunc(fn)
}

func parsePath(req *http.Request) metric {
	return metric{
		mtype: req.PathValue("type"),
		name:  req.PathValue("name"),
		value: req.PathValue("value"),
	}
}

func (m metric) validate() error {
	switch m.mtype {
	case TypeGauge:
		if _, err := strconv.ParseFloat(m.value, 64); err != nil {
			return fmt.Errorf("invalid metric value, gauge must be float64, given: %s", m.value)
		}
	case TypeCounter:
		if _, err := strconv.ParseInt(m.value, 10, 64); err != nil {
			return fmt.Errorf("invalid metric value, counter must be int64, given: %s", m.value)
		}
	default:
		return fmt.Errorf("invalid metric type: %s", m.mtype)
	}

	return nil
}

func saveMetric(res http.ResponseWriter, req *http.Request) {
	switch m := parsePath(req); m.mtype {
	case TypeGauge:
		v, err := strconv.ParseFloat(m.value, 64)
		if err != nil {
			panic(err) // если есть ошибка, значит не сработала проверка выше
		}
		storage.PutGauge(m.name, v)
	case TypeCounter:
		v, err := strconv.ParseInt(m.value, 10, 64)
		if err != nil {
			panic(err) // если есть ошибка, значит не сработала проверка выше
		}
		storage.PutCounter(m.name, v)
	}

	res.WriteHeader(http.StatusOK)
}
