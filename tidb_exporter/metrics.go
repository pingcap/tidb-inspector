package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	promTiKVType    = "tikv"
	promTiDBType    = "tidb"
	promPDType      = "pd"
	promPromType    = "prometheus"
	promGrafanaType = "grafana"

	checkedAllInstance = "all"
	checkedFailed      = "failed"
	checkedSuccess     = "success"
	checkedRestart     = "restart"
)

var (
	exporter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "exporter",
			Subsystem: "op",
			Name:      "check_instance",
			Help:      "check cluster status.",
		}, []string{"type", "checked", "status"})
)

func daemonProm(instance string) {
	for {
		err := push.AddFromGatherer(
			"tools",
			map[string]string{"instance": instance},
			*metrics,
			prometheus.DefaultGatherer,
		)
		if err != nil {
			log.Errorf("can not push metrics to prometheus with error %v", err)
		}
		time.Sleep(10 * time.Second)
	}
}
