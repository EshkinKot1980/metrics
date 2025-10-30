package audit

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/EshkinKot1980/metrics/internal/server/audit/mocks"
)

func TestURLauditor_Handle(t *testing.T) {
	event := Event{
		TS:      time.Now().Unix(),
		Metrics: []string{"PollCount", "FreeMemory"},
		IP:      "172.19.0.4",
	}
	eventJSON, err := json.Marshal(event)
	require.Nil(t, err, "Audit event encoding")

	tests := []struct {
		name       string
		netError   bool
		reqBody    []byte
		statusCode int
		event      Event
		lSetup     func(t *testing.T) Logger
	}{
		{
			name:       "succes",
			reqBody:    eventJSON,
			statusCode: http.StatusOK,
			event:      event,
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				return mocks.NewMockLogger(ctrl)
			},
		},
		{
			name:     "network_error",
			netError: true,
			event:    event,
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Error("failed to send audit event", gomock.All())
				return logger
			},
		},
		{
			name:       "not_succes_response",
			reqBody:    eventJSON,
			statusCode: http.StatusBadRequest,
			event:      event,
			lSetup: func(t *testing.T) Logger {
				ctrl := gomock.NewController(t)
				logger := mocks.NewMockLogger(ctrl)
				logger.EXPECT().
					Error("failed to send audit event", gomock.All())
				return logger
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method, "Request Method")

				body, err := io.ReadAll(r.Body)
				require.Nil(t, err, "Request Body reading")
				assert.Equal(t, test.reqBody, body, "Request body")

				w.WriteHeader(test.statusCode)
			}
			server := httptest.NewServer(http.HandlerFunc(handler))
			defer server.Close()

			loger := test.lSetup(t)
			a := NewURLauditor(server.URL, loger)

			if test.netError {
				server.Close()
			}

			a.Handle(test.event)
		})
	}
}
