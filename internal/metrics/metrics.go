package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	HTTPRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "forum_http_requests_total",
		Help: "Total number of HTTP requests",
	},
		[]string{"method", "route", "status"})

	HTTPRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "forum_http_request_duration_seconds",
		Help:    "HTTP request latency in seconds",
		Buckets: prometheus.DefBuckets,
	},
		[]string{"method", "route"})

	PostsCreatedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "forum_posts_created_total",
		Help: "Total number of created posts",
	})

	CommentsCreatedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "forum_comments_created_total",
		Help: "Total number of created comments",
	})

	LikeToggleTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "forum_like_toggles_total",
		Help: "Total number of like toggle actions",
	},
		[]string{"target_type", "action"})
)

func MustRegister() {
	prometheus.MustRegister(
		HTTPRequestsTotal,
		HTTPRequestDuration,
		PostsCreatedTotal,
		CommentsCreatedTotal,
		LikeToggleTotal)
}
