package main

import (
	"flag"
	"fmt"
	"github.com/juju/errors"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	//KillCMD kill tidb process
	KillCMD = "kill -9"
)

var (
	user              = flag.String("user", "root", "tidb user")
	password          = flag.String("password", "", "tidb password")
	metrics           = flag.String("metrics", "", "metrics address")
	querytimeout      = flag.Int("query-timeout", 20, "tidb execute query timeout")
	interval          = flag.Int64("interval", 180, "check alive interval")
	tidbs             = flag.String("tidb-list", "", "tidb list, example:'10.0.3.5:4000,10.0.3.6:4000'")
	tikvs             = flag.String("tikv-list", "", "tikv list, example:'10.0.3.5:20160,10.0.3.6:20160'")
	pds               = flag.String("pd-list", "", "pd list, example:'http://10.0.3.5:2379,http://10.0.3.6:2379'")
	daemon            = flag.Bool("daemon", false, "run as daemon")
	alertmangers      = flag.String("alertmanger-list", "", "alertmanger list,example:'10.0.3.5:9093,10.0.3.6:9093'")
	prometheusAddress = flag.String("prometheus-address", "", "prometheus address,example:'http://10.0.3.6:9090'")
	grafanaAddress    = flag.String("grafana-address", "", "grafan address,example:'http://10.0.3.6:3000'")

	logFile  = flag.String("log-file", "", "log filename")
	logLevel = flag.String("log-level", "info", "log level:panic,fatal,error,warning,info,debug")

	//suffixCommand = flag.String("suffix-command", "", "run shell command when check tidb failed, work with option kill-trigger")
	//killTrigger   = flag.Bool("kill-trigger", false, "only this mon)itor running with tidb process one host, kill -9 tidb's process that listen port;work with suffix-command")
	//self          = flag.Bool("self", true, "only this monitor and watched tidb's process are running one host, work with option kill-trigger and self")
)

func checkParams() error {
	if !*daemon && (*tidbs == "" && *pds == "") {
		return errors.Errorf("please input tidb-list or pd-list when not run as daemon")
	}
	return nil
}

func daemonMode() {
	instance := getHostName()
	if *metrics != "" {
		prometheus.MustRegister(exporter)
		go daemonProm(instance)
	}

	cinterval := time.Duration(*interval) * time.Second
	if *tidbs != "" {
		go goroutineTiDB(strings.Split(*tidbs, ","), cinterval)
	}

	if *pds != "" {
		go goroutinePD(strings.Split(*pds, ","), cinterval)
		go goroutineTiKV(strings.Split(*pds, ","), cinterval)
	}

	if *prometheusAddress != "" {
		go goroutineProm(*prometheusAddress, cinterval)
	}

	if *grafanaAddress != "" {
		go goroutineGrafana(*grafanaAddress, cinterval)
	}

}

func main() {
	flag.Parse()
	if err := checkParams(); err != nil {
		fmt.Printf("parms error : %v", err)
		return
	}
	if !*daemon {
		commandMode()
		return
	}
	if *logFile != "" {
		lf, err := os.OpenFile(*logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0660)
		if err != nil {
			fmt.Printf("can not open log file %s error %v", *logFile, err)
			return
		}
		formatter := &log.TextFormatter{
			FullTimestamp: true,
		}
		log.SetFormatter(formatter)
		log.SetOutput(lf)
		defer lf.Close()
	}
	log.SetLevel(log.DebugLevel)

	go daemonMode()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		sig := <-sc
		log.Errorf("get signal [%d] and exit", sig)
		wg.Done()
	}()
	wg.Wait()
}
