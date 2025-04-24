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

func New(s Storage) http.Handler {
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
		body = fmt.Sprintf("%v\r\n", gauge)
	case TypeCounter:
		counter, err = storage.GetCounter(name)
		body = fmt.Sprintf("%v\r\n", counter)
	default:
		err = errors.New("Invalid metric type:" + t)
	}

	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	} else {
		res.Write([]byte(body))
	}
}
