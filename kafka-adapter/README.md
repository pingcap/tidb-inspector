kafka-adapter 
------

**This tool is used to push alert messages to kafka**
### Build
- install Golang(1.8.3+)
- `make`

**The target executable binary file is bin/kafka-adapter**

### Usages
```
Usage of ./kafka-adapte:
  -kafka-address string
    	kafka adddress (default "10.0.3.4:9092,10.0.3.5:9092,10.0.3.6:9092")
  -kafka-topic string
    	kafka topic (default "test")
  -log-file string
    	log file (default "kafka-adapter.log")
  -log-level string
    	log level: debug, info, warn, error, fatal (default "info")
  -port int
    	http listen port (default 28082) 
```


### Examples:
1. Prometheus update alert.yml file and restart 
	- URL`https://github.com/pingcap/tidb-ansible/tree/master/roles/prometheus/files`
2. Alertmanager configure file add webhook scope 
	- `webhook` as first router
	- `url` URLPath is hardcode `/v1/alertmanager`
	
```
	route:
  receiver: "pingcap-dt"
  group_by: ['env','instance','type','group','job']
  group_wait:      30s
  group_interval:  1m
  repeat_interval: 3m
  routes:
  - match:
    receiver: webhook
  - match:
    receiver: pingcap-dt

  receivers:
  
  - name: 'webhook'
  webhook_configs:
  - send_resolved: false
    url: 'http://10.0.3.6:28082/v1/alertmanager'
    
```
	
3. run it 
	- `./kafka-adapter  -port 28082  -kafka-addres 10.0.3.4:9092,10.0.3.5:9092,10.0.3.6:9092 -kafka-topic test`

	
