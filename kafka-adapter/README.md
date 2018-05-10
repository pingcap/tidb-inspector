kafka-adapter 
------

### Usages

```
Usage of ./kafka-adapter:
  -kafka-address string
    	kafka address, example: 10.0.3.4:9092,10.0.3.5:9092,10.0.3.6:9092
  -kafka-topic string
    	kafka topic
  -log-file string
    	log file path
  -log-level string
    	log level: debug, info, warn, error, fatal (default "info")
  -log-rotate string
    	log file rotate type: hour/day (default "day")
  -port int
    	port to listen on for the web interface (default 28082)
```

### Example:

```
#!/bin/bash
nohup ./kafka-adapter \
    --log-level="info" \
    --log-file="kafka-adapter.log" \
    --kafka-address="172.16.10.50:9092,172.16.10.61:9092,172.16.10.62:9092" \
    --kafka-topic="test" &
```

```
route:
  # A default receiver
  receiver: "db-alert-email"

  # The labels by which incoming alerts are grouped together. For example,
  # multiple alerts coming in for cluster=A and alertname=LatencyHigh would
  # be batched into a single group.
  group_by: ['env','instance','alertname','type','group','job']

  # When a new group of alerts is created by an incoming alert, wait at
  # least 'group_wait' to send the initial notification.
  # This way ensures that you get multiple alerts for the same group that start
  # firing shortly after another are batched together on the first
  # notification.
  group_wait:      30s

  # When the first notification was sent, wait 'group_interval' to send a batch
  # of new alerts that started firing for that group.
  group_interval:  3m

  # If an alert has successfully been sent, wait 'repeat_interval' to
  # resend them.
  repeat_interval: 3m

  routes:
  - match:
    receiver: kafka-adapter
    continue: true
  - match:
      env: test-cluster
    receiver: db-alert-slack
  - match:
      env: test-cluster
    receiver: db-alert-email

receivers:
- name: 'kafka-adapter'
  webhook_configs:
  - send_resolved: true
    url: 'http://172.16.10.49:28082/v1/alertmanager'
```
