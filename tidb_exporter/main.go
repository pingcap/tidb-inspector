package main

import (
	"database/sql"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/juju/errors"
	"github.com/ngaut/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Exporter collects database query stats and exports them using
// the prometheus metrics package.
type Exporter struct {
	nodes []tidbNode
}

type tidbNode struct {
	addr string
	db   *sql.DB
}

type tidbOpts struct {
	addrs    string
	username string
	password string
}

// ParseHostPortAddr returns a host:port list
func ParseHostPortAddr(s string) ([]string, error) {
	strs := strings.Split(s, ",")
	addrs := make([]string, 0, len(strs))

	for _, str := range strs {
		str = strings.TrimSpace(str)

		_, _, err := net.SplitHostPort(str)
		if err != nil {
			return nil, errors.Annotatef(err, `tidb.addrs does not have the form "host:port": %s`, str)
		}

		addrs = append(addrs, str)
	}

	return addrs, nil
}

// NewExporter returns an initialized Exporter.
func NewExporter(opts tidbOpts) (*Exporter, error) {
	addrs, err := ParseHostPortAddr(opts.addrs)
	if err != nil {
		return nil, errors.Trace(err)
	}

	nodes := make([]tidbNode, 0, len(addrs))
	for _, addr := range addrs {
		db, err := accessDatabase(opts.username, opts.password, addr, dbname)
		if err != nil {
			return nil, errors.Trace(err)
		}

		node := tidbNode{addr: addr, db: db}
		nodes = append(nodes, node)
	}

	log.Infof("monitoring tidb servers: %s", opts.addrs)
	return &Exporter{nodes: nodes}, nil
}

// Describe describes all the metrics ever exported by tidb_exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- queryErrorDesc
}

// Collect fetches database query stats and delivers them
// as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	var wg = sync.WaitGroup{}

	getNodeMetrics := func(node tidbNode) {
		defer wg.Done()

		var queryError float64
		label, err := probeQuery(node.addr, node.db)
		if err != nil {
			queryError = 1
		}
		ch <- prometheus.MustNewConstMetric(
			queryErrorDesc, prometheus.GaugeValue, queryError, node.addr, label,
		)
	}

	for _, node := range e.nodes {
		wg.Add(1)
		go getNodeMetrics(node)
	}

	wg.Wait()
}

func checkParameters(opts tidbOpts) {
	if opts.addrs == "" {
		log.Fatalf("missing startup parameter: --tidb.addrs")
	}

	if opts.username == "" {
		log.Fatalf("missing startup parameter: --tidb.username")
	}
}

func main() {
	var (
		listenAddress = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry").Default(":9200").String()
		metricsPath   = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
		logFile       = kingpin.Flag("log-file", "Log file path.").Default("").String()
		logLevel      = kingpin.Flag("log-level", "Log level: debug, info, warn, error, fatal.").Default("info").String()
		logRotate     = kingpin.Flag("log-rotate", "Log file rotate type: hour/day.").Default("day").String()

		opts = tidbOpts{}
	)

	kingpin.Flag("tidb.addrs", "Addresses (host:port) of TiDB server nodes, comma separated.").Default("").StringVar(&opts.addrs)
	kingpin.Flag("tidb.username", "TiDB user name.").Default("").StringVar(&opts.username)
	kingpin.Flag("tidb.password", "TiDB user password.").Default("").StringVar(&opts.password)

	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	checkParameters(opts)

	log.SetLevelByString(*logLevel)
	if *logFile != "" {
		log.SetOutputByName(*logFile)
		if *logRotate == "hour" {
			log.SetRotateByHour()
		} else {
			log.SetRotateByDay()
		}
	}

	log.Info("Starting tidb_exporter")
	exporter, err := NewExporter(opts)
	if err != nil {
		log.Fatalf("initialize tidb_exporter error, %v", errors.ErrorStack(err))
	}

	prometheus.MustRegister(exporter)

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
	        <head><title>TiDB Exporter</title></head>
	        <body>
	        <h1>TiDB Exporter</h1>
	        <p><a href='` + *metricsPath + `'>Metrics</a></p>
	        </body>
	        </html>`))
	})

	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		syscall.SIGKILL,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		sig := <-sc
		log.Infof("got signal [%d] to exit", sig)

		for _, node := range exporter.nodes {
			if node.db != nil {
				node.db.Close()
			}
		}

		os.Exit(0)
	}()

	log.Info("Listening on", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
