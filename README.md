# serversage

## promc

`promc` is a CLI utility for generating Prometheus metrics code from JSON configuration files.

### Installation

`go install github.com/remiges-tech/serversage/cmd/promc`

### Usage

`promc generate -c config.json -o metrics.go -p dbmetrics`


- `-c`, `--config`: Path to the JSON configuration file (required).
- `-o`, `--output`: Path to the output file for the generated code (required).
- `-p`, `--package`: Package name for the generated code (required).

### Configuration File Format

```json
{
  "metrics": [
    {
      "name": "metric_name",
      "type": "counter | gauge | histogram | summary",
      "description": "A brief description of the metric.",
      "labels": [
        "label1",
        "label2"
      ],
      "buckets": [0.1, 0.5, 1, 2.5, 5, 10]  # For histogram only
    }
  ]
}
```

The JSON configuration consists of a top-level metrics field, which is an array of metric definitions. Each metric definition has the following fields:
- name (required): The name of the metric.
- type (required): The type of the metric. Valid values are counter, gauge, histogram, and summary.
- description (optional): A brief description of the metric.
- labels (optional): An array of label names associated with the metric.
- buckets (optional, histogram only): An array of bucket values for histogram metrics.

### Examples

```json
{
  "metrics": [
    {
      "name": "request_duration_seconds",
      "type": "histogram",
      "description": "Request duration in seconds.",
      "labels": [
        "method",
        "path"
      ],
      "buckets": [0.1, 0.5, 1, 2.5, 5, 10]
    },
    {
      "name": "active_users",
      "type": "gauge",
      "description": "Number of active users."
    }
  ]
}
```

This configuration defines two metrics: request_duration_seconds (a histogram) and active_users (a gauge).
