package main

import (
	log "github.com/sirupsen/logrus"
	"time"
)

func probeProm(promAddr string) error {
	return xGet(promAddr, nil, false)
}

func goroutineProm(promAddr string, checkInterval time.Duration) {
	for {
		if err := probeProm(promAddr); err != nil {
			log.Debugf("probe prometheus with error %v", err)
			exporter.WithLabelValues(promPromType, checkedAllInstance, checkedFailed).Inc()
		} else {
			exporter.WithLabelValues(promPromType, checkedAllInstance, checkedSuccess).Inc()
		}
		time.Sleep(checkInterval)
	}

}

func probeGrafana(grafanaAddr string) error {
	return xGet(grafanaAddr, nil, false)
}

func goroutineGrafana(grafanaAddr string, checkInterval time.Duration) {
	for {
		if err := probeGrafana(grafanaAddr); err != nil {
			log.Debugf("probe grafan with error %v", err)
			exporter.WithLabelValues(promGrafanaType, checkedAllInstance, checkedFailed).Inc()
		} else {
			exporter.WithLabelValues(promGrafanaType, checkedAllInstance, checkedSuccess).Inc()
		}
		time.Sleep(checkInterval)
	}
}
