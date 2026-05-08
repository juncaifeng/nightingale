package collector

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/prometheus/prompb"
)

func TestAlertMetricsRegistrar_EventToTimeSeries(t *testing.T) {
	registrar := NewAlertMetricsRegistrar("test")

	tests := []struct {
		name     string
		event    *AlertEvent
		expected int
	}{
		{
			name: "firing alert",
			event: &AlertEvent{
				Ident:     "host-01",
				Name:      "HighCPU",
				Severity:  "critical",
				Value:     95.5,
				AlertType: "system",
				Source:    "prometheus",
				StartTime: 1715184000,
				EndTime:   0,
				Labels: map[string]string{
					"environment": "prod",
				},
			},
			expected: 2,
		},
		{
			name: "resolved alert",
			event: &AlertEvent{
				Ident:     "host-02",
				Name:      "DiskFull",
				Severity:  "warning",
				Value:     85.0,
				AlertType: "system",
				Source:    "zabbix",
				StartTime: 1715184000,
				EndTime:   1715187600,
				Labels:    nil,
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := registrar.EventToTimeSeries(tt.event)
			if len(result) != tt.expected {
				t.Errorf("expected %d time series, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestAlertMetricsRegistrar_MetricLabels(t *testing.T) {
	registrar := NewAlertMetricsRegistrar("test")

	event := &AlertEvent{
		Ident:     "test-host",
		Name:      "TestAlert",
		Severity:  "critical",
		Value:     100.0,
		AlertType: "test",
		Source:    "unit-test",
		StartTime: 1715184000,
		EndTime:   0,
		Labels: map[string]string{
			"env":      "testing",
			"region":   "us-west",
			"instance": "i-12345",
		},
	}

	result := registrar.EventToTimeSeries(event)

	if len(result) == 0 {
		t.Fatal("expected at least one time series")
	}

	ts := result[0]

	labelMap := make(map[string]string)
	for _, label := range ts.Labels {
		labelMap[label.Name] = label.Value
	}

	expectedLabels := []string{"__name__", "alert_name", "severity", "alert_type", "source", "ident", "env", "region", "instance"}

	for _, expectedLabel := range expectedLabels {
		if _, exists := labelMap[expectedLabel]; !exists {
			t.Errorf("expected label %s not found", expectedLabel)
		}
	}
}

func TestAlertCollectorHandler_CollectSingleAlert(t *testing.T) {
	gin.SetMode(gin.TestMode)

	registrar := NewAlertMetricsRegistrar("test")
	handler := NewAlertCollectorHandler(registrar, &MockWriter{})

	t.Run("valid single alert", func(t *testing.T) {
		event := AlertEvent{
			Ident:     "host-01",
			Name:      "HighCPU",
			Severity:  "critical",
			Value:     95.5,
			AlertType: "system",
			Source:    "prometheus",
			StartTime: 1715184000,
		}

		body, _ := json.Marshal(event)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/alert", strings.NewReader(string(body)))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CollectSingleAlert(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		if response["status"] != "success" {
			t.Errorf("expected status success, got %v", response["status"])
		}
	})

	t.Run("invalid request body", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/alert", strings.NewReader("invalid json"))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CollectSingleAlert(c)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestAlertCollectorHandler_CollectAlerts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	registrar := NewAlertMetricsRegistrar("test")
	handler := NewAlertCollectorHandler(registrar, &MockWriter{})

	t.Run("valid batch alerts", func(t *testing.T) {
		events := []AlertEvent{
			{
				Ident:     "host-01",
				Name:      "HighCPU",
				Severity:  "critical",
				Value:     95.5,
				AlertType: "system",
				Source:    "prometheus",
				StartTime: 1715184000,
			},
			{
				Ident:     "host-02",
				Name:      "DiskFull",
				Severity:  "warning",
				Value:     85.0,
				AlertType: "system",
				Source:    "zabbix",
				StartTime: 1715184000,
			},
		}

		body, _ := json.Marshal(events)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/alerts", strings.NewReader(string(body)))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CollectAlerts(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		if response["status"] != "success" {
			t.Errorf("expected status success, got %v", response["status"])
		}

		if int(response["total"].(float64)) != 2 {
			t.Errorf("expected total 2, got %v", response["total"])
		}
	})

	t.Run("empty events array", func(t *testing.T) {
		events := []AlertEvent{}

		body, _ := json.Marshal(events)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/alerts", strings.NewReader(string(body)))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CollectAlerts(c)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

type MockWriter struct {
	written []*prompb.TimeSeries
}

func (m *MockWriter) Write(ts *prompb.TimeSeries) {
	m.written = append(m.written, ts)
}
