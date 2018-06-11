TiDB exporter
------

TiDB exporter for Prometheus. The `tidb_exporter` will expose metrics via HTTP on the `/metrics` endpoint. Currently only one tidb server address is supported. It is recommended to set load balancing IP. We'll support multiple tidb servers in the near future.

#### Required Grants

```
CREATE USER 'tidb_exporter'@'%' IDENTIFIED BY 'XXXXXXXX';
GRANT SELECT ON mysql.tidb TO 'tidb_exporter'@'%';
```

#### Build

```
make tidb_exporter
```

#### Running

```
$ ./tidb_exporter -h
usage: tidb_exporter [<flags>]

Flags:
  -h, --help              Show context-sensitive help (also try --help-long and --help-man).
      --web.listen-address=":9200"
                          Address to listen on for web interface and telemetry
      --web.telemetry-path="/metrics"
                          Path under which to expose metrics.
      --log-file=""       Log file path.
      --log-level="info"  Log level: debug, info, warn, error, fatal.
      --log-rotate="day"  Log file rotate type: hour/day.
      --tidb.address=""   Address (host:port) of TiDB server.
      --tidb.username=""  TiDB user name.
      --tidb.password=""  TiDB user password.
```

```
nohup bin/tidb_exporter \
    --web.listen-address=":9200" \
    --web.telemetry-path="/metrics" \
    --tidb.address="172.16.10.71:4000" \
    --tidb.username="tidb_exporter" \
    --tidb.password="XXXXXXXX"
    --log-level="info" \
    --log-file="log/tidb_exporter.log" &
```

#### Metrics

| Name | Description |
| ---- | ----------- |
| `tidb_exporter_tidb_query_error` | Whether an error occurs while sending query to tidb server. Query: `SELECT count(*) FROM mysql.tidb` |

#### Prometheus Configuration

```
scrape_configs:
  - job_name: 'tidb_exporter'
    scrape_interval: 2m
    scrape_timeout: 1m
    honor_labels: true # don't overwrite job & instance labels
    static_configs:
    - targets:
      - '172.16.10.71:9200'
```
