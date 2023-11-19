package logger

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// logs collector
type logSink struct {
	logs []string
}

func (l *logSink) Write(p []byte) (n int, err error) {
	l.logs = append(l.logs, string(p))
	return len(p), nil
}

func (l *logSink) Index(i int) string {
	return l.logs[i]
}

// log message to parse
type LogMessage struct {
	Level     string
	Host      string
	URL       string
	Method    string
	MsServed  float64 `json:"ms_served"`
	Status    int
	BytesSent int `json:"bytes_sent"`
	Time      int64
	Caller    string
	Message   string
}

func TestLogRequestResponse(t *testing.T) {
	// create logger to sink
	SetLevel("info")
	sink := &logSink{}
	l := zerolog.New(sink).With().Timestamp().Caller().Logger()
	log = &l

	// request and response body
	body := bytes.NewReader([]byte("this is a test body"))
	bodyLen := body.Len()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resBody, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		w.Write(resBody)
	})

	req := httptest.NewRequest(http.MethodPost, "/", body)
	w := httptest.NewRecorder()
	handler := LogRequestResponse(testHandler)
	handler.ServeHTTP(w, req)
	res := w.Result()
	defer res.Body.Close()

	message := LogMessage{}
	err := json.Unmarshal([]byte(sink.Index(0)), &message)
	require.NoError(t, err, "Expect log message to be parsed")
	assert.Equal(t, "info", message.Level, "Expected level 'info', got '%s'", message.Level)
	assert.Equal(t, "example.com", message.Host, "Expected host 'example.com', got '%s'", message.Host)
	assert.Equal(t, "/", message.URL, "Expected URL '/', got '%s'", message.URL)
	assert.Equal(t, "POST", message.Method, "Expected method 'POST', got '%s'", message.Method)
	assert.Greater(t, message.MsServed, 0.0, "Expected ms_served > 0, got '%f'", message.MsServed)
	assert.Equal(t, 200, message.Status, "Expected status '200', got '%d'", message.Status)
	assert.Equal(t, bodyLen, message.BytesSent, "Expected bytes sent '%d', got '%d'", body.Len(), message.BytesSent)
	assert.Greater(t, message.Time, int64(0), "Expected time > 0, got '%d'", message.Time)
	assert.NotEqualf(t, "", message.Caller, "Expected caller not empty")
	assert.Equal(t, "http request", message.Message, "Expected message 'http request', got '%s'", message.Message)

}
