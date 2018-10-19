// Copyright 2018 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/ngaut/log"
	"github.com/pingcap/kvproto/pkg/debugpb"
	"github.com/pingcap/tidb-inspect-tools/pkg/utils"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	getMetricsTimeout = time.Duration(15) * time.Second
)

// Exporter exposes Prometheus metrics of TiKV server
type Exporter struct {
	client *rpcClient
	stores []string
}

type tikvOpts struct {
	addrs    string
	security Security
}

func checkFlags(opts tikvOpts) {
	if opts.addrs == "" {
		log.Fatal("missing startup flag: --tikv.addrs")
	}
}

// NewExporter returns an initialized Exporter.
func NewExporter(opts tikvOpts) (*Exporter, error) {
	stores, err := utils.ParseHostPortAddr(opts.addrs)
	if err != nil {
		return nil, errors.Trace(err)
	}

	client := newRPCClient(opts.security)
	for _, store := range stores {
		_, err := client.getConn(store)
		if err != nil {
			return nil, errors.Trace(err)
		}
	}

	return &Exporter{client: client, stores: stores}, nil
}

func sanitizeLabels(
	metricFamilies map[string]*dto.MetricFamily,
	groupingLabels map[string]string,
) {
	for _, mf := range metricFamilies {
		for _, m := range mf.GetMetric() {
			for key, value := range groupingLabels {
				l := &dto.LabelPair{
					Name:  proto.String(key),
					Value: proto.String(value),
				}
				m.Label = append(m.Label, l)
			}
			sort.Sort(prometheus.LabelPairSorter(m.Label))
		}
	}
}

func (e *Exporter) getMetricFamilies() []*dto.MetricFamily {
	wg := sync.WaitGroup{}
	mutex := &sync.Mutex{}
	allMetrics := make([]*dto.MetricFamily, 0, 1024)

	getStoreMetrics := func(store string) {
		defer wg.Done()

		tikvConn, err := e.client.getConn(store)
		if err != nil {
			return
		}

		tikvClient := debugpb.NewDebugClient(tikvConn)
		ctx, cancel := context.WithTimeout(context.Background(), getMetricsTimeout)
		defer cancel()
		metrics, err := tikvClient.GetMetrics(ctx, &debugpb.GetMetricsRequest{})
		if err != nil {
			log.Errorf("tikv store '%s', get metrics error, %v", store, err)
			return
		}

		mData := metrics.GetPrometheus()
		storeID := metrics.GetStoreId()

		labels := map[string]string{
			"job":      fmt.Sprintf("tikv_%d", storeID),
			"instance": store,
		}

		var parser expfmt.TextParser
		metricFamilies, err := parser.TextToMetricFamilies(bytes.NewBufferString(mData))
		if err != nil {
			log.Errorf("tikv store '%s', TextToMetricFamilies error, %v", store, err)
			return
		}

		sanitizeLabels(metricFamilies, labels)

		mutex.Lock()
		for _, m := range metricFamilies {
			allMetrics = append(allMetrics, m)
		}
		mutex.Unlock()
	}

	for _, store := range e.stores {
		wg.Add(1)
		go getStoreMetrics(store)
	}

	wg.Wait()

	return allMetrics
}

func main() {
	var (
		listenAddress = kingpin.Flag("web.listen-address", "Address on which to expose metrics and web interface.").Default(":9600").String()
		metricsPath   = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
		logFile       = kingpin.Flag("log-file", "Log file path.").Default("").String()
		logLevel      = kingpin.Flag("log-level", "Log level: debug, info, warn, error, fatal.").Default("info").String()
		logRotate     = kingpin.Flag("log-rotate", "Log file rotate type: hour/day.").Default("day").String()

		opts = tikvOpts{}
	)

	kingpin.Flag("tikv.addrs", "Addresses (host:port) of TiKV instances, comma separated.").Default("").StringVar(&opts.addrs)
	kingpin.Flag("tls.ca-file", "Path of file that contains list of trusted SSL CAs for connection with tikv servers.").Default("").StringVar(&opts.security.ClusterSSLCA)
	kingpin.Flag("tls.cert-file", "Path of file that contains X509 certificate in PEM format for connection with tikv servers.").Default("").StringVar(&opts.security.ClusterSSLCert)
	kingpin.Flag("tls.key-file", "Path of file that contains X509 key in PEM format for connection with with tikv servers.").Default("").StringVar(&opts.security.ClusterSSLKey)
	kingpin.Version(utils.GetRawInfo("tikv_metrics_proxy"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	checkFlags(opts)

	log.SetLevelByString(*logLevel)
	if *logFile != "" {
		log.SetOutputByName(*logFile)
		if *logRotate == "hour" {
			log.SetRotateByHour()
		} else {
			log.SetRotateByDay()
		}
	}

	log.Info("starting tikv_metrics_proxy")

	exporter, err := NewExporter(opts)
	if err != nil {
		log.Fatalf("initialize tikv_metrics_proxy error, %v", errors.ErrorStack(err))
	}

	prometheus.DefaultGatherer = prometheus.Gatherers{
		prometheus.GathererFunc(func() ([]*dto.MetricFamily, error) { return exporter.getMetricFamilies(), nil }),
	}

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
	        <head><title>TiKV metrics proxy</title></head>
	        <body>
	        <h1>TiKV metrics proxy</h1>
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

		exporter.client.closeConns()
		os.Exit(0)
	}()

	log.Info("listening on", *listenAddress)
	err = http.ListenAndServe(*listenAddress, nil)
	exporter.client.closeConns()
	if err != nil {
		log.Fatal(err)
	}
}
