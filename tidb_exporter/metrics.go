package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "tidb_exporter"
)

var (
	queryErrorDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "tidb", "query_error"),
		"Whether an error occurs while sending query to tidb server.",
		[]string{"target", "label"}, nil,
	)
)
