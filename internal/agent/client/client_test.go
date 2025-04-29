package client

import(
	"strings"
	"strconv"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/EshkinKot1980/metrics/internal/agent/model"
	"github.com/EshkinKot1980/metrics/internal/agent/storage"
)

func testRequest(req *http.Request) func(t *testing.T) {
	return func(t *testing.T) {
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, ContentType, req.Header.Get("Content-Type"))
		
		pathParts := strings.Split(req.URL.Path, "/")
		require.Equal(t, 5 , len(pathParts))
		assert.Equal(t, PathPrefix,  pathParts[1])
		require.Contains(t, []string{TypeCounter, TypeGauge}, pathParts[2])

		switch pathParts[2] {
		case TypeGauge:
			 _, err := strconv.ParseFloat(pathParts[4], 64)
			assert.Nil(t, err)
		case TypeCounter:
			_, err := strconv.ParseInt(pathParts[4], 10, 64)
			assert.Nil(t, err)
		}
	}
}

func TestReport(t *testing.T) {
	server := httptest.NewServer(makeHadler(t))
	defer server.Close()

	st := storage.New()
	initStorage(st)
	c := New(st, server.URL)
	c.Report()
}

func makeHadler(t *testing.T) http.Handler {
	fn := func(res http.ResponseWriter, req *http.Request) {
		name :=  "report test PATH:" + req.URL.Path
		t.Run(name, testRequest(req))
	}

	return http.HandlerFunc(fn)
}

func initStorage(s *storage.MemoryStorage) {
	s.Put(
		[]model.Counter{
			model.Counter{Name: "TestCounter", Value: 13},
			model.Counter{Name: "Visitord", Value: 256},
		},
		[]model.Gauge{
			model.Gauge{Name: "ConstE", Value: 2.71828},
			model.Gauge{Name: "TTL", Value: 3.14e50},
		},
	)
}
