tidb-inspect-tools
------

tidb-inspect-tools are some useful inspection tool collections for TiDB.

## Tool list

- [grafana_collector](https://github.com/pingcap/tidb-inspect-tools/tree/master/grafana_collector)

  A tool for generating PDF reports for Grafana dashboards.

- [kafka_adapter](https://github.com/pingcap/tidb-inspect-tools/tree/master/kafka_adapter)

  Alertmanager webhook receiver for the Kafka service.

- [syslog_adapter](https://github.com/pingcap/tidb-inspect-tools/tree/master/syslog_adapter)

  Alertmanager webhook receiver for the syslog service.

- [tcp_prober](https://github.com/pingcap/tidb-inspect-tools/tree/master/tcp_prober)

  A tool for checking monitoring modules's health via TCP port probe.

- [tidb_exporter](https://github.com/pingcap/tidb-inspect-tools/tree/master/tidb_exporter)

  A tool for checking the TiDB server's health by sending SQL queries to it. The `tidb_exporter` exposes metrics via HTTP on the `/metrics` endpoint for Prometheus.

## License
Apache 2.0 license. See the [LICENSE](https://github.com/pingcap/tidb-inspect-tools/blob/master/LICENSE) file for details.
