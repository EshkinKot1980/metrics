package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/EshkinKot1980/metrics/internal/common/models"
)

func ExampleRetrieveHandler_GetByPath() {
	service := newMockService()
	logger := newStubLogger()
	handler := NewRetrieveHandler(service, logger)

	req := httptest.NewRequest(http.MethodGet, "/value/counter/TestCounter", nil)
	// необходимо указать, так как httptest не поддерживает шаблоны для путей
	req.SetPathValue("type", "counter")
	req.SetPathValue("name", "TestCounter")

	// В случае использования роутера chi.Mux либо другого роутера совместимого с net/http
	// r.Get("/value/{type}/{name}", retriever.GetByPath)

	w := httptest.NewRecorder()
	handler.GetByPath(w, req)
	res := w.Result()
	defer res.Body.Close()

	fmt.Println(res.StatusCode)

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("failed to read body: ", err)
		return
	}
	body := strings.TrimSuffix(string(resBody), "\n")
	fmt.Println(body)

	// Output:
	// 200
	// 13
}

func ExampleRetrieveHandler_GetJSON() {
	service := newMockService()
	logger := newStubLogger()
	handler := NewRetrieveHandler(service, logger)

	reqBody := []byte(`{"id":"TestGauge","type":"gauge"}`)
	req := httptest.NewRequest(http.MethodPost, "/value", bytes.NewBuffer(reqBody))

	// В случае использования роутера chi.Mux
	// r.Post("/value", handler.GetJSON)

	w := httptest.NewRecorder()
	handler.GetJSON(w, req)
	res := w.Result()
	defer res.Body.Close()

	fmt.Println(res.StatusCode)

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("failed to read body: ", err)
		return
	}
	body := strings.TrimSuffix(string(resBody), "\n")
	fmt.Println(body)

	// Output:
	// 200
	// {"id":"TestGauge","type":"gauge","value":3.14}
}

func ExampleUpdateHandler_UpdateFromPath() {
	service := newMockService()
	logger := newStubLogger()
	handler := NewUpdateHandler(service, logger)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/TestGauge/3.14", nil)
	// необходимо указать, так как httptest не поддерживает шаблоны для путей
	req.SetPathValue("type", "gauge")
	req.SetPathValue("name", "TestGauge")
	req.SetPathValue("value", "3.14")

	// В случае использования роутера chi.Mux либо другого роутера совместимого с net/http
	// r.Post("/update/{type}/{name}/{value}", handler.UpdateFromPath)

	w := httptest.NewRecorder()
	handler.UpdateFromPath(w, req)
	res := w.Result()
	defer res.Body.Close()

	fmt.Println(res.StatusCode)

	// Output: 200
}

func ExampleUpdateHandler_Update() {
	service := newMockService()
	logger := newStubLogger()
	handler := NewUpdateHandler(service, logger)

	reqBody := []byte(`{"id":"TestCounter","type":"counter","delta":1}`)
	r := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBuffer(reqBody))
	r.Header.Set("Content-Type", "application/json")

	// В случае использования роутера chi.Mux
	// r.Post("/update", handler.Update)

	w := httptest.NewRecorder()
	handler.Update(w, r)
	res := w.Result()
	defer res.Body.Close()

	fmt.Println(res.StatusCode)

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("failed to read body: ", err)
		return
	}
	body := strings.TrimSuffix(string(resBody), "\n")
	fmt.Println(body)

	// Output:
	// 200
	// {"id":"TestCounter","type":"counter","delta":14}
}

func ExampleUpdateHandler_UpdateList() {
	service := newMockService()
	logger := newStubLogger()
	handler := NewUpdateHandler(service, logger)

	reqBody := []byte(`[
						{"id":"TestCounter","type":"counter","delta":1},
						{"id":"TestGauge","type":"gauge","value":3.14}
					  ]`)
	r := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewBuffer(reqBody))
	r.Header.Set("Content-Type", "application/json")

	// В случае использования роутера chi.Mux
	// r.Post("/updates", handler.UpdateList)

	w := httptest.NewRecorder()
	handler.UpdateList(w, r)
	res := w.Result()
	defer res.Body.Close()

	fmt.Println(res.StatusCode)

	// Output: 200
}

type mockService struct{}

func newMockService() *mockService {
	return &mockService{}
}

func (s *mockService) Fill(metric models.Metrics) (models.Metrics, error) {
	var delta int64 = 13
	value := 3.14

	switch metric.MType {
	case models.TypeCounter:
		metric.Delta = &delta
	case models.TypeGauge:
		metric.Value = &value
	}

	return metric, nil
}

func (s *mockService) Put(metric models.Metrics) (models.Metrics, error) {
	if metric.MType == models.TypeCounter {
		*metric.Delta += 13
	}

	return metric, nil
}

func (s *mockService) PutList(ctx context.Context, metrics []models.Metrics) error {
	return nil
}

type stubLogger struct{}

func newStubLogger() *stubLogger {
	return &stubLogger{}
}

func (l *stubLogger) Error(message string, err error) {}
