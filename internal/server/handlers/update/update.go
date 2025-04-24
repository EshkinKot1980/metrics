package update

import (
	"strconv"
	"strings"
	"errors"
	"net/http"
)

const (
	TypeGauge = "gauge"
	TypeCounter = "counter"
)

type Storage interface {
	PutCounter(name string, incriment int64)
	PutGauge(name string, value float64)
}

type metric struct {
	mtype string
	name string
	value string
}

var storage Storage

func New(s Storage) http.Handler {
	storage = s

	return validateHeaders(
		validateData(
			http.HandlerFunc(update),
		),
	)
}

func validateHeaders(next http.Handler) http.Handler {
	fn := func(res http.ResponseWriter, req *http.Request) {
		header := req.Header.Get("content-type")
		if strings.Count(header, "text/plain") == 1 {
			next.ServeHTTP(res, req)
		} else {
			http.Error(
				res,
				"Invalid Content-Type header, header must be \"text/plain\"",
				http.StatusBadRequest,
			)
		}
	}
	
	return http.HandlerFunc(fn)
}


func validateData(next http.Handler) http.Handler {
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
	return metric {
		mtype: req.PathValue("type"),
		name:  req.PathValue("name"),
		value: req.PathValue("value"),
	}
}

func (m metric) validate() error {
	var err error
	
	switch m.mtype {
	case TypeGauge:
		if _, e := strconv.ParseFloat(m.value, 64); e != nil {
			err = errors.New("invalid metric value, gauge must be float64")
		}
	case TypeCounter:
		if _, e := strconv.ParseInt(m.value, 10, 64); e != nil {
			err = errors.New("invalid metric value, counter must be int64")
		}
	default:
		err = errors.New("invalid metric type:" + m.mtype)
	}

	return err
}

func update(res http.ResponseWriter, req *http.Request) {
	switch m := parsePath(req); m.mtype {
	case TypeGauge:
		v, _ := strconv.ParseFloat(m.value, 64)
		storage.PutGauge(m.name, v)		
	case TypeCounter:
		v, _ := strconv.ParseInt(m.value, 10, 64)
		storage.PutCounter(m.name,v)
	}
}
