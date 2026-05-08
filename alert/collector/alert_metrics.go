package collector

import (
	"github.com/prometheus/prometheus/prompb"
)

type AlertEvent struct {
	Ident        string            `json:"ident"`
	Name         string            `json:"name"`
	Severity     string            `json:"severity"`
	Value        float64           `json:"value"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartTime    int64             `json:"start_time"`
	EndTime      int64             `json:"end_time"`
	AlertType    string            `json:"alert_type"`
	Source       string            `json:"source"`
}

type AlertMetricsRegistrar struct {
	metricPrefix string
	builtInMetrics map[string]bool
}

func NewAlertMetricsRegistrar(prefix string) *AlertMetricsRegistrar {
	return &AlertMetricsRegistrar{
		metricPrefix: prefix,
		builtInMetrics: map[string]bool{
			"alert_count":    true,
			"alert_firing":   true,
			"alert_resolved": true,
			"alert_pending":  true,
		},
	}
}

func (r *AlertMetricsRegistrar) EventToTimeSeries(event *AlertEvent) []*prompb.TimeSeries {
	var ts []*prompb.TimeSeries

	countTS := r.createAlertCountMetric(event)
	if countTS != nil {
		ts = append(ts, countTS)
	}

	if event.EndTime == 0 {
		statusTS := r.createAlertStatusMetric(event)
		if statusTS != nil {
			ts = append(ts, statusTS)
		}
	} else {
		resolvedTS := r.createAlertResolvedMetric(event)
		if resolvedTS != nil {
			ts = append(ts, resolvedTS)
		}
	}

	return ts
}

func (r *AlertMetricsRegistrar) createAlertCountMetric(event *AlertEvent) *prompb.TimeSeries {
	labels := []prompb.Label{
		{Name: "__name__", Value: r.metricPrefix + "_alert_count_total"},
		{Name: "alert_name", Value: event.Name},
		{Name: "severity", Value: event.Severity},
		{Name: "alert_type", Value: event.AlertType},
		{Name: "source", Value: event.Source},
		{Name: "ident", Value: event.Ident},
	}

	for k, v := range event.Labels {
		labels = append(labels, prompb.Label{Name: k, Value: v})
	}

	return &prompb.TimeSeries{
		Labels: labels,
		Samples: []prompb.Sample{
			{Value: 1, Timestamp: event.StartTime * 1000},
		},
	}
}

func (r *AlertMetricsRegistrar) createAlertStatusMetric(event *AlertEvent) *prompb.TimeSeries {
	labels := []prompb.Label{
		{Name: "__name__", Value: r.metricPrefix + "_alert_firing"},
		{Name: "alert_name", Value: event.Name},
		{Name: "severity", Value: event.Severity},
		{Name: "alert_type", Value: event.AlertType},
		{Name: "source", Value: event.Source},
		{Name: "ident", Value: event.Ident},
	}

	for k, v := range event.Labels {
		labels = append(labels, prompb.Label{Name: k, Value: v})
	}

	return &prompb.TimeSeries{
		Labels: labels,
		Samples: []prompb.Sample{
			{Value: event.Value, Timestamp: event.StartTime * 1000},
		},
	}
}

func (r *AlertMetricsRegistrar) createAlertResolvedMetric(event *AlertEvent) *prompb.TimeSeries {
	labels := []prompb.Label{
		{Name: "__name__", Value: r.metricPrefix + "_alert_resolved_total"},
		{Name: "alert_name", Value: event.Name},
		{Name: "severity", Value: event.Severity},
		{Name: "alert_type", Value: event.AlertType},
		{Name: "source", Value: event.Source},
		{Name: "ident", Value: event.Ident},
	}

	for k, v := range event.Labels {
		labels = append(labels, prompb.Label{Name: k, Value: v})
	}

	duration := event.EndTime - event.StartTime
	return &prompb.TimeSeries{
		Labels: labels,
		Samples: []prompb.Sample{
			{Value: float64(duration), Timestamp: event.EndTime * 1000},
		},
	}
}
