package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	requestCreate = "create"
	requestRead   = "read"
	requestUpdate = "update"
	requestDelete = "delete"

	requestSuccess = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "user_request_success_total",
		Help: "The total number of successful requests",
	}, []string{"request_type"})

	requestFailure = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "user_request_failure_total",
		Help: "The total number of failed requests",
	}, []string{"request_type", "error_code"})
)
