package audit

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/EshkinKot1980/metrics/internal/server/audit/mocks"
	"github.com/EshkinKot1980/metrics/internal/server/config"
)

func TestAuditor_Rise(t *testing.T) {
	t.Run("succes", func(t *testing.T) {
		event := Event{
			TS:      time.Now().Unix(),
			Metrics: []string{"PollCount", "FreeMemory"},
			IP:      "172.19.0.4",
		}
		eventJSON, err := json.Marshal(event)
		require.Nil(t, err, "Audit event encoding")

		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method, "Request Method")

			body, err := io.ReadAll(r.Body)
			require.Nil(t, err, "Request Body reading")
			assert.Equal(t, eventJSON, body, "Request body")

			w.WriteHeader(http.StatusOK)
		}
		server := httptest.NewServer(http.HandlerFunc(handler))
		defer server.Close()

		fileName := t.TempDir() + "/audit.log"

		cfg := &config.Config{
			AuditFile: fileName,
			AuditURL:  server.URL,
		}

		ctrl := gomock.NewController(t)
		logger := mocks.NewMockLogger(ctrl)

		auditor := NewAuditor(cfg, logger)
		defer auditor.Halt()
		auditor.Rise(event)

		<-time.After(100 * time.Millisecond)
		file, err := os.Open(fileName)
		require.Nil(t, err, "Audit file opening")
		defer file.Close()

		content, err := os.ReadFile(fileName)
		require.Nil(t, err, "Audit file reading")
		eol := []byte("\n")
		want := append(eventJSON, eol...)

		assert.Equal(t, want, content, "Audit file")
	})

	t.Run("invalid_file", func(t *testing.T) {

		fileName := t.TempDir() + "/eadOnlyFile"

		file, err := os.OpenFile(fileName, os.O_CREATE, 0444)
		require.Nil(t, err, "Read only dir creating")
		file.Close()

		cfg := &config.Config{AuditFile: fileName}

		ctrl := gomock.NewController(t)
		logger := mocks.NewMockLogger(ctrl)
		logger.EXPECT().Error("failed to create file auditor", gomock.All())

		_ = NewAuditor(cfg, logger)
	})
}
