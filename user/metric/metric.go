package metric

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestCreate = "create"
	RequestRead   = "read"
	RequestUpdate = "update"
	RequestDelete = "delete"

	RequestSuccess = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "user_request_success_total",
		Help: "The total number of successful requests",
	}, []string{"request_type"})

	RequestFailure = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "user_request_failure_total",
		Help: "The total number of failed requests",
	}, []string{"request_type", "error_code"})
)
