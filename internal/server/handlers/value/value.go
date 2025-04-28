package value

import (
	"fmt"
	"errors"
	"net/http"
)

const (
	TypeGauge = "gauge"
	TypeCounter = "counter"
)

type Storage interface {
	GetCounter(name string) (int64 ,error)
	GetGauge(name string) (float64 ,error)
}

var storage Storage

func New(s Storage) http.HandlerFunc {
	storage = s

	return http.HandlerFunc(value)
}

func value(res http.ResponseWriter, req *http.Request) {
	var (
		name = req.PathValue("name")
		gauge float64
		counter int64
		err error
		body string	
	)
	
	switch t := req.PathValue("type"); t {
	case TypeGauge:
		gauge, err = storage.GetGauge(name)
		body = fmt.Sprintf("%v", gauge)
	case TypeCounter:
		counter, err = storage.GetCounter(name)
		body = fmt.Sprintf("%v", counter)
	default:
		err = errors.New("invalid metric type: " + t)
	}

	if err != nil {
		http.Error(res, err.Error(), http.StatusNotFound)
	} else {
		res.Write([]byte(body))
	}
}
