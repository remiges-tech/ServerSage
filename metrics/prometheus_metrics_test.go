package metrics

import (
	"math"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestRegisterWithLabels(t *testing.T) {
	metrics := NewPrometheusMetrics()

	metrics.RegisterWithLabels("test_metric1", "Counter", "Test metric with labels", []string{"label1", "label2"})

	if _, ok := metrics.counterVecs["test_metric1"]; !ok {
		t.Errorf("Metric 'test_metric' was not registered")
	}
}

func TestRecordWithLabels(t *testing.T) {
	metrics := NewPrometheusMetrics()

	metrics.RegisterWithLabels("test_metric2", "Counter", "Test metric with labels", []string{"label1", "label2"})
	metrics.RecordWithLabels("test_metric", 1.0, "value1", "value2")

	if _, ok := metrics.counterVecs["test_metric2"]; !ok {
		t.Errorf("Metric 'test_metric' was not recorded")
	}
}

func TestTimer(t *testing.T) {
	metrics := NewPrometheusMetrics()

	// Start a new timer
	id := metrics.StartTimer()

	// Simulate a block of code execution by sleeping for 1 second
	sleepDuration := 1 * time.Second
	time.Sleep(sleepDuration)

	// Record the time
	metrics.RecordTime("test_timer", id)

	// Check if the recorded time is approximately equal to the sleep duration
	if histogramVec, ok := metrics.histogramVecs["op_exec_time"]; ok {
		// Collect the metrics
		metricChan := make(chan prometheus.Metric, 1)
		histogramVec.Collect(metricChan)
		metric := <-metricChan

		// Get the histogram from the metric
		dtoMetric := &dto.Metric{}
		metric.Write(dtoMetric)
		histogram := dtoMetric.GetHistogram()

		// Check if the recorded time is approximately equal to the sleep duration
		if histogram.GetSampleCount() == 0 {
			t.Errorf("No observations were recorded")
		} else {
			observedDuration := histogram.GetSampleSum()
			if math.Abs(observedDuration-float64(sleepDuration.Seconds())) > 0.01 {
				t.Errorf("Recorded time is not approximately equal to the sleep duration")
			}
		}
	} else {
		t.Errorf("Histogram 'op_exec_time' was not found")
	}
}
