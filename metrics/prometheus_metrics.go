package metrics

import (
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricType defines a custom type for different kinds of metrics.
type MetricType struct {
	name string
}

// String implements stringer for MetricType.
func (m MetricType) String() string {
	return m.name
}

// Predefined MetricType instances for each metric type.
var (
	metricCounter   = MetricType{"Counter"}
	metricGauge     = MetricType{"Gauge"}
	metricHistogram = MetricType{"Histogram"}
)

// MetricCounter returns an instance of MetricType representing a Counter metric.
// This function encapsulates the creation of a Counter metric type.
func MetricCounter() MetricType { return metricCounter }

// MetricGauge returns an instance of MetricType representing a Gauge metric.
func MetricGauge() MetricType { return metricGauge }

// MetricHistogram returns an instance of MetricType representing a Histogram metric.
func MetricHistogram() MetricType { return metricHistogram }

// PrometheusMetrics is a structure that implements the Metrics interface using Prometheus as the backend.
// It stores mappings for different Prometheus metric types (Counter, Gauge, Histogram) and their vector counterparts.
type PrometheusMetrics struct {
	counters      map[string]prometheus.Counter
	counterVecs   map[string]*prometheus.CounterVec // New map for CounterVec objects
	gauges        map[string]prometheus.Gauge
	gaugeVecs     map[string]*prometheus.GaugeVec // New map for CounterVec objects
	histograms    map[string]prometheus.Histogram
	histogramVecs map[string]*prometheus.HistogramVec
	customBuckets map[string][]float64 // Stores custom buckets for histograms
	timers        map[uint64]time.Time // Stores the start time of function/block executions. Used by RecordExecTime.
}

// NewPrometheusMetrics creates and initializes a new instance of PrometheusMetrics.
// This function sets up the internal maps used to store various types of Prometheus metrics,
// including counters, gauges, histograms, and their labeled (vector) versions.
//
// Additionally, it registers a histogram vector named "op_exec_time" with a label also named "op_exec_time".
// This histogram vector is used by the RecordExecTime function to record the execution time of operations.
// The label "op_exec_time" is used to differentiate the execution times of different operations.
//
// Here's an example of how to use it:
//
//	func main() {
//		p := NewPrometheusMetrics()
//		// Now you can use p to record metrics...
//	}
func NewPrometheusMetrics() *PrometheusMetrics {
	p := &PrometheusMetrics{
		counters:      make(map[string]prometheus.Counter),
		counterVecs:   make(map[string]*prometheus.CounterVec),
		gauges:        make(map[string]prometheus.Gauge),
		gaugeVecs:     make(map[string]*prometheus.GaugeVec),
		histograms:    make(map[string]prometheus.Histogram),
		histogramVecs: make(map[string]*prometheus.HistogramVec),
		customBuckets: make(map[string][]float64),
		timers:        make(map[uint64]time.Time), // Initialize the timers map
	}

	// Register a histogram for operation execution times
	histogramVec := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "op_exec_time",
		Help: "Time taken by operations to execute",
	}, []string{"op"}) // the label "op" contains the name of the operation
	prometheus.MustRegister(histogramVec)
	p.histogramVecs["op_exec_time"] = histogramVec

	return p
}

// SetCustomBuckets allows setting custom bucket sizes for histograms.
// requiring finer or broader granularity. The 'name' parameter specifies the metric name, and 'buckets' is a slice
// of float64 values defining the bucket thresholds.
func (p *PrometheusMetrics) SetCustomBuckets(name string, buckets []float64) {
	p.customBuckets[name] = buckets
}

// Register creates and registers a new metric in the Prometheus registry based on the provided type.
// Supported metric types include 'Counter', 'Gauge', and 'Histogram'.
// The method takes the metric 'name', its 'metricType', and a 'help' string describing the metric.
// For 'Histogram' types, it uses custom buckets if they have been set; otherwise, it falls back to default buckets.
func (p *PrometheusMetrics) Register(name string, metricType MetricType, help string) {
	switch metricType {
	case metricCounter:
		// Creating a new Counter metric
		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: help,
		})
		// Registering the Counter with Prometheus
		prometheus.MustRegister(counter)
		// Storing the reference in the counters map
		p.counters[name] = counter

	case metricGauge:
		// Creating a new Gauge metric
		gauge := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: name,
			Help: help,
		})
		// Registering the Gauge with Prometheus
		prometheus.MustRegister(gauge)
		// Storing the reference in the gauges map
		p.gauges[name] = gauge

	case metricHistogram:
		buckets, ok := p.customBuckets[name]
		if !ok {
			buckets = prometheus.DefBuckets // Use default buckets if not specified
		}
		histogram := prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    name,
			Help:    help,
			Buckets: buckets,
		})
		prometheus.MustRegister(histogram)
		p.histograms[name] = histogram
	default:
		// Handle unknown metric type
		log.Printf("Error: Attempted to register unknown metric type '%s' with name '%s'", metricType, name)
	}
}

// Record updates the value of a Prometheus metric without labels.
// It is used for recording values for counters, gauges, and histograms based on the metric 'name'.
// The method identifies the correct metric type and performs the appropriate action: 'Add' for counters,
// 'Set' for gauges, and 'Observe' for histograms. The 'value' parameter is the value to record.
func (p *PrometheusMetrics) Record(name string, value float64) {
	if counter, ok := p.counters[name]; ok {
		counter.Add(value)
		return
	}

	if gauge, ok := p.gauges[name]; ok {
		gauge.Set(value)
		return
	}

	if histogram, ok := p.histograms[name]; ok {
		histogram.Observe(value)
		return
	}

}

// RegisterWithLabels creates and registers a new labeled metric.
// This method is similar to 'Register' but for metrics with labels (like CounterVec, GaugeVec, HistogramVec).
// It takes the metric 'name', 'metricType', a 'help' description, and a slice of 'labels' (the label keys).
// For 'HistogramVec', it respects custom buckets if set for the given metric name.
func (p *PrometheusMetrics) RegisterWithLabels(name string, metricType MetricType, help string, labels []string) {
	// Creating a new Counter metric with labels
	switch metricType {
	case metricCounter:
		counterVec := prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: name,
			Help: help,
		}, labels)
		// Registering the Counter with Prometheus
		prometheus.MustRegister(counterVec)
		// Storing the reference in the counters map
		p.counterVecs[name] = counterVec
	case metricGauge:
		// Creating a new Gauge metric with labels
		gaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: name,
			Help: help,
		}, labels)
		// Registering the Gauge with Prometheus
		prometheus.MustRegister(gaugeVec)
		// Storing the reference in the gaugeVecs map
		p.gaugeVecs[name] = gaugeVec
	case metricHistogram:
		// Creating a new Histogram metric with labels
		buckets, ok := p.customBuckets[name]
		if !ok {
			buckets = prometheus.DefBuckets // Use default buckets if not specified
		}
		histogramVec := prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    name,
			Help:    help,
			Buckets: buckets,
		}, labels)
		// Registering the Histogram with Prometheus
		prometheus.MustRegister(histogramVec)
		// Storing the reference in the histogramVecs map
		p.histogramVecs[name] = histogramVec
	}
}

// RecordWithLabels updates the value of a labeled Prometheus metric.
// This method is used for metrics that were registered with labels, such as those created via 'RegisterWithLabels'.
// It finds the appropriate metric based on 'name' and updates it with the given 'value' and 'labelValues'.
// The 'labelValues' are variadic parameters that should match the order and number of labels defined during registration.
func (p *PrometheusMetrics) RecordWithLabels(name string, value float64, labelValues ...string) {
	if counterVec, ok := p.counterVecs[name]; ok {
		counterVec.WithLabelValues(labelValues...).Add(value)
		return
	}

	if gaugeVec, ok := p.gaugeVecs[name]; ok {
		gaugeVec.WithLabelValues(labelValues...).Set(value)
		return
	}

	if histogramVec, ok := p.histogramVecs[name]; ok {
		histogramVec.WithLabelValues(labelValues...).Observe(value)
		return
	}
}

var timerID uint64

// StartTimer is used to start a new timer. This function should be called
// at the beginning of a block of code or function whose execution time you want to measure.
// It returns a unique identifier for the timer, which can be used to stop the timer later.
// The timer uses the current system time as the start time.
// This function is thread-safe as it uses atomic operations to generate unique identifiers for each timer.
//
// Example usage:
//
//	id := metrics.StartTimer()
//	// Some code or function you want to time...
//	metrics.RecordTime("myFunction", id)
func (p *PrometheusMetrics) StartTimer() uint64 {
	id := atomic.AddUint64(&timerID, 1)
	p.timers[id] = time.Now()
	return id
}

// RecordTime is used to stop a timer that was previously started with StartTimer
// and record the elapsed time. This function should be called at the end of
// a block of code or function whose execution time you want to measure.
// The 'name' parameter is a label that is used to differentiate the execution times
// of different blocks of code or functions. The 'id' parameter is the unique identifier
// of the timer, which was returned by StartTimer when the timer was started.
//
// Example usage:
//
//	id := metrics.StartTimer()
//	// Some code or function you want to time...
//	metrics.RecordTime("myOperation", id)
//
// This will record the execution time of the block of code or function under the label "myOperation".
func (p *PrometheusMetrics) RecordTime(name string, id uint64) {
	if start, ok := p.timers[id]; ok {
		elapsed := time.Since(start).Seconds()
		if histogramVec, ok := p.histogramVecs["op_exec_time"]; ok {
			histogramVec.WithLabelValues(name).Observe(elapsed)
		}
		delete(p.timers, id)
	}
}

// StartMetricsServer initializes and starts an HTTP server on the specified 'port' to expose Prometheus metrics.
// This server provides an endpoint for Prometheus to scrape the collected metrics.
// Typically it would be used to start a metrics server in a separate goroutine to keep it running independently.
func (p *PrometheusMetrics) StartMetricsServer(port string) {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":"+port, nil)
}
