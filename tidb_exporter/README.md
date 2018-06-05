tidb_exporter
------

**This tool is used to check TiDB's cluster status and report to prometheus, send critical message to alertmanger**
### Build
- install Golang(1.10+)
- `make`

**The target executable binary file is bin/tidb_exporter**

### Usages
```
Usage of ./bin/tidb_exporter:
  -alertmanger-list string
    	alertmanger list,example:'10.0.3.5:9093,10.0.3.6:9093'
  -daemon
    	run as daemon
  -grafana-address string
    	grafan address,example:'http://10.0.3.6:3000'
  -interval int
    	check alive interval (default 180)
  -log-file string
    	log filename
  -log-level string
    	log level:panic,fatal,error,warning,info,debug (default "info")
  -metrics string
    	metrics address
  -password string
    	tidb password
  -pd-list string
    	pd list, example:'http://10.0.3.5:2379,http://10.0.3.6:2379'
  -prometheus-address string
    	prometheus address,example:'http://10.0.3.6:9090'
  -query-timeout int
    	tidb execute query timeout (default 20)
  -tidb-list string
    	tidb list, example:'10.0.3.5:4000,10.0.3.6:4000'
  -tikv-list string
    	tikv list, example:'10.0.3.5:20160,10.0.3.6:20160'
  -user string
    	tidb user (default "root")
```


### Examples:
- check tidb's cluster
```
 ./bin/tidb_exporter \
     -pd-list "http://10.1.0.4:2379,http://10.1.0.5:2379,http://10.1.0.6:2379" \
     -tidb-list "10.1.0.4:4000,10.1.0.5:4000,10.1.0.6:4000" \
     -metrics 10.1.0.4:9091  \
     -daemon \
     -alertmanger-list 10.1.0.4:9093
```

	
