package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Counters for the Santa API routes
	PreflightRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "moroz_santa_preflight_requests_total",
			Help: "Total number of preflight requests.",
		},
		[]string{"method"},
	)

	RuleDownloadRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "moroz_santa_ruledownload_requests_total",
			Help: "Total number of ruledownload requests.",
		},
		[]string{"method"},
	)

	EventUploadRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "moroz_santa_eventupload_requests_total",
			Help: "Total number of eventupload requests.",
		},
		[]string{"method"},
	)

	PostflightRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "moroz_santa_postflight_requests_total",
			Help: "Total number of postflight requests.",
		},
		[]string{"method"},
	)

	// Histogram for preflight request durations
	PreflightRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "moroz_santa_preflight_request_duration_seconds",
			Help:    "Duration of preflight requests in seconds.",
			Buckets: prometheus.DefBuckets, // Default buckets for timing metrics
		},
		[]string{"status"}, // Labels: "success" or "error"
	)

	// Histogram for rule download request duration
	RuleDownloadRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "moroz_santa_ruledownload_request_duration_seconds",
			Help:    "Duration of ruledownload requests in seconds.",
			Buckets: prometheus.DefBuckets, // Default buckets for timing metrics
		},
		[]string{"status"}, // Labels: "success" or "error"
	)

	EventUploadRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "moroz_santa_eventupload_request_duration_seconds",
			Help:    "Duration of eventupload requests in seconds.",
			Buckets: prometheus.DefBuckets, // Default buckets for timing metrics
		},
		[]string{"status"}, // Labels: success or error
	)
)

func Init() {
	// Register the counters with Prometheus
	prometheus.MustRegister(PreflightRequests)
	prometheus.MustRegister(RuleDownloadRequests)
	prometheus.MustRegister(EventUploadRequests)
	prometheus.MustRegister(PostflightRequests)
	prometheus.MustRegister(RuleDownloadRequestDuration)
	prometheus.MustRegister(EventUploadRequestDuration)
}
