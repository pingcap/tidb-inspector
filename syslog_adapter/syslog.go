package main

import (
	"fmt"
	"github.com/ngaut/log"
	"strings"
)

const (
	timeFormat = "2006-01-02 15:04:05"
)

// syslogMessage syslog base struct
type syslogMessage struct {
	AlertName   string
	Cluster     string
	Instance    string
	Value       string
	Summary     string
	Level       string
	Description string
	Status      string
	Time        string
}

// Send message to syslog service
func (r *Run) sendAlert(msg string) error {
	_, err := fmt.Fprint(r.sysLog, msg)
	if err != nil {
		log.Errorf("Failed to push message to syslog service: %v", err)
		return err
	}
	return nil
}

// TransferData transfer alert to syslog string
// The following is customer's syslog format:
// ip: alert-ip from startup parameter or instace
// hostname: alert-hostname from startup parameter or instance
// "Database"
// "TiDB"
// "Log"
// syslogMsg.Instance
// syslogMsg.Value
// syslogMsg.Summary
// syslogMsg.Summary
// syslogMsg.Level
// syslogMsg.Description
// "1": send_resolved: true
// syslogMsg.Time
// syslogMsg.Status: alertstate: firing: 0, resolved: 1
func (r *Run) TransferData(ad *AlertData) {
	for _, at := range ad.Alerts {
		syslogMsg := &syslogMessage{
			AlertName:   getValue(at.Labels, "alertname"),
			Cluster:     getValue(at.Labels, "env"),
			Instance:    getValue(at.Labels, "instance"),
			Value:       getValue(at.Annotations, "value"),
			Summary:     getValue(at.Annotations, "summary"),
			Level:       getValue(at.Labels, "level"),
			Description: getValue(at.Annotations, "description"),
			Status:      at.Status,
			Time:        at.StartsAt.Format(timeFormat),
		}

		syslogMsg.Level = *alertLevel
		ip := *alertIP
		hostname := *alertHostname

		switch syslogMsg.Status {
		case "firing":
			syslogMsg.Status = "0"
		case "resolved":
			syslogMsg.Status = "1"
		default:
			syslogMsg.Status = "1"
		}

		switch syslogMsg.AlertName {
		case "TiDB_is_Down", "TiKV_is_Down", "PD_is_Down":
			ip = strings.Split(syslogMsg.Instance, ":")[0]
			hostname = ip
		}

		var msg []string
		msg = append(msg, ip, hostname, "Database", "TiDB", "Log", syslogMsg.Instance, syslogMsg.Value, syslogMsg.Summary, syslogMsg.Summary, syslogMsg.Level, syslogMsg.Description, "1", syslogMsg.Time, syslogMsg.Status)

		log.Debugf("Push syslogMsg: %v", msg)

		if err := r.sendAlert(strings.Join(msg, "||")); err != nil {
			log.Errorf("Push message to syslog server with error: %v", err)
		}
	}
}

func getValue(kv KV, key string) string {
	if val, ok := kv[key]; ok {
		return val
	}
	return ""
}
