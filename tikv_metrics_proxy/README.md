TiKV metrics proxy
------

TiKV metrics proxy is a tool for exposing Prometheus metrics of TiKV server. The `tikv_metrics_proxy` exposes metrics via HTTP on the `/metrics` endpoint. 

#### Build

```
make tikv_metrics_proxy
```

#### Running

```
$ ./tikv_metrics_proxy -h
usage: tikv_metrics_proxy [<flags>]

Flags:
  -h, --help              Show context-sensitive help (also try --help-long and --help-man).
      --web.listen-address=":9600"
                          Address on which to expose metrics and web interface.
      --web.telemetry-path="/metrics"
                          Path under which to expose metrics.
      --log-file=""       Log file path.
      --log-level="info"  Log level: debug, info, warn, error, fatal.
      --log-rotate="day"  Log file rotate type: hour/day.
      --tikv.addrs=""     Addresses (host:port) of TiKV instances, comma separated.
      --tls.ca-file=""    Path of file that contains list of trusted SSL CAs for connection with tikv servers.
      --tls.cert-file=""  Path of file that contains X509 certificate in PEM format for connection with tikv servers.
      --tls.key-file=""   Path of file that contains X509 key in PEM format for connection with with tikv servers.
      --version           Show application version.
```

```
nohup bin/tikv_metrics_proxy \
    --web.listen-address=":9600" \
    --tikv.addrs="172.16.10.72:20160,172.16.10.73:20160,172.16.10.74:20160" \
    --log-level="info" \
    --log-file="tikv_metrics_proxy.log" &
```

#### Prometheus Configuration

```
scrape_configs:
  - job_name: "tikv"
    honor_labels: true # don't overwrite job & instance labels
    static_configs:
    - targets:
      - '172.16.10.71:9600'
```
