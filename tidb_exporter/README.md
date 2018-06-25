TiDB exporter
------

TiDB exporter (for Prometheus) is a tool for checking the TiDB server's health by sending SQL queries to it. The `tidb_exporter` exposes metrics via HTTP on the `/metrics` endpoint. 

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
  -h, --help               Show context-sensitive help (also try --help-long and --help-man).
      --web.listen-address=":9500"
                           Address to listen on for web interface and telemetry
      --web.telemetry-path="/metrics"
                           Path under which to expose metrics.
      --log-file=""        Log file path.
      --log-level="info"   Log level: debug, info, warn, error, fatal.
      --log-rotate="day"   Log file rotate type: hour/day.
      --tidb.addrs=""  Addresses (host:port) of TiDB server nodes, comma separated.
      --tidb.username=""   TiDB user name.
      --tidb.password=""   TiDB user password.
```

```
nohup bin/tidb_exporter \
    --web.listen-address=":9500" \
    --web.telemetry-path="/metrics" \
    --tidb.addrs="172.16.10.71:4000,172.16.10.72:4000" \
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
