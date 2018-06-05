package main

import (
	"fmt"
	//"github.com/prometheus/prometheus/notifier"
	//"github.com/prometheus/prometheus/pkg/labels"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

const (
	alertPushEndpoint = "/api/v1/alerts"
)

var (
	baseLabelsMap = map[string]string{
		"level":   "emergency",
		"monitor": "tidb_exporter",
	}
)

/* alert message json
[
    {
        "annotations": {
            "description": "alert: instance: 10.1.0.5:9100 values: 40.76086956521739",
            "summary": "disk inode more than 80%",
            "value": "40.76086956521739"
        },
        "endsAt": "0001-01-01T00:00:00Z",
        "generatorURL": "http://10.1.0.4:9090/graph?g0.expr=node_filesystem_files_free%7Bfstype%3D~%22%28ext.%7Cxfs%29%22%7D+%2F+node_filesystem_files%7Bfstype%3D~%22%28ext.%7Cxfs%29%22%7D+%2A+100+%3C%3D+80u0026g0.tab=1",
        "labels": {
            "alertname": "NODE_disk_inode_more_than_80%",
            "cluster": "test-cluster",
            "device": "/dev/sda1",
            "env": "test-cluster",
            "expr": "node_filesystem_files_free{fstype=~\"(ext.|xfs)\"} / node_filesystem_files{fstype=~\"(ext.|xfs)\"}  * 100 < 20",
            "fstype": "xfs",
            "instance": "10.1.0.5:9100",
            "job": "overwritten-nodes",
            "level": "critical",
            "monitor": "prometheus",
            "mountpoint": "/boot"
        },
        "startsAt": "2018-03-08T15:46:58.503042643+08:00"
    }
]
*/

type alert *Alert

// Alert ...
// import error and copy it from prometheus https://github.com/prometheus/prometheus/blob/master/notifier/notifier.go
// Alert is a generic representation of an alert in the Prometheus eco-system.
type Alert struct {
	// Label value pairs for purpose of aggregation, matching, and disposition
	// dispatching. This must minimally include an "alertname" label.
	Labels map[string]string `json:"labels"`

	// Extra key/value information which does not define alert identity.
	Annotations map[string]string `json:"annotations"`

	// The known time range for this alert. Both ends are optional.
	StartsAt     time.Time `json:"startsAt,omitempty"`
	EndsAt       time.Time `json:"endsAt,omitempty"`
	GeneratorURL string    `json:"generatorURL,omitempty"`
}

func genAlert(alertLables, alertAnnotations map[string]string) alert {
	ls := make(map[string]string)
	//generate labels
	for k, v := range baseLabelsMap {
		ls[k] = v
	}
	for k, v := range alertLables {
		ls[k] = v
	}
	//generate annotations
	//for k, v := range alertAnnotations {
	//	as = append(as, genLabel(k, v))
	//}

	return &Alert{
		StartsAt:    time.Now(),
		Labels:      ls,
		Annotations: alertAnnotations,
	}
}

func sendAlert(alertHosts []string, alertLabels, alertAnnotations map[string]string) bool {
	alerts := []alert{
		genAlert(alertLabels, alertAnnotations),
	}
	log.Debugf("alerts message %v", *alerts[0])

	var sendStatus bool
	for _, host := range alertHosts {
		alertURL := fmt.Sprintf("http://%s%s", host, alertPushEndpoint)
		if err := xPost(alertURL, alerts); err == nil {
			sendStatus = true
		} else {
			log.Errorf("send alert message error %v", err)
		}
	}
	return sendStatus
}

func xAlert(env, promType, instance, checked string) bool {
	alertHosts := strings.Split(*alertmangers, ",")
	if len(alertHosts) == 0 {
		return false
	}
	ls := map[string]string{
		"job":       "probe",
		"env":       env,
		"alertname": fmt.Sprintf("%s failed", promType),
		"instance":  instance,
		"type":      promType,
	}
	as := map[string]string{
		"description": fmt.Sprintf("probe failed %s %s %s", promType, instance, checked),
		"summary":     "probe_failed tidb_exporter",
		"value":       "",
		"checked":     checked,
	}

	return sendAlert(alertHosts, ls, as)
}
