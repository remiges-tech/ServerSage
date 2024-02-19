package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/remiges-tech/serversage/metrics"
)

func main() {
	// Create a new PrometheusMetrics instance
	p := metrics.NewPrometheusMetrics()

	// Register a new counter for counting iterations
	p.Register("iterationCount", metrics.MetricCounter(), "Number of iterations")

	// Start the metrics server in a separate goroutine
	go p.StartMetricsServer("8080")

	// Continuously record metrics in an infinite loop
	for i := 1; ; i++ {
		// Start a new timer
		id := p.StartTimer()

		// Simulate a function execution by sleeping for a random duration
		sleepDuration := time.Duration(rand.Intn(1000)) * time.Millisecond
		time.Sleep(sleepDuration)

		// Record the function execution time
		p.RecordTime("myFunctionTime", id)

		// Increment the iteration counter
		p.Record("iterationCount", 1)

		fmt.Printf("Recorded iteration %d\n", i)
	}
}
