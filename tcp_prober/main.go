package main

import (
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ngaut/log"
)

var (
	kafkaAddress = flag.String("kafka-address", "", "kafka address, example: 10.0.3.4:9092,10.0.3.5:9092,10.0.3.6:9092")
	kafkaTopic   = flag.String("kafka-topic", "", "kafka topic")
	logFile      = flag.String("log-file", "", "log file path")
	logLevel     = flag.String("log-level", "info", "log level: debug, info, warn, error, fatal")
	logRotate    = flag.String("log-rotate", "day", "log file rotate type: hour/day")
	configFile   = flag.String("config", "", "path to configuration file")
	clusterName  = flag.String("cluster-name", "", "TiDB Cluster name")
)

func main() {
	flag.Parse()
	if *kafkaAddress == "" {
		log.Fatalf("missing parameter: -kafka-address")
	}
	addrs := strings.Split(*kafkaAddress, ",")

	if *kafkaTopic == "" {
		log.Fatalf("missing parameter: -kafka-topic")
	}

	if *clusterName == "" {
		log.Fatalf("missing parameter: -cluster-name")
	}

	err := SetConfig(*configFile)
	if err != nil {
		log.Fatalf("parsing configure file error: %v", err)
	}

	log.SetLevelByString(*logLevel)
	if *logFile != "" {
		log.SetOutputByName(*logFile)
		if *logRotate == "hour" {
			log.SetRotateByHour()
		} else {
			log.SetRotateByDay()
		}
	}

	r := &Run{}

	if err := r.CreateKafkaProducer(addrs); err != nil {
		log.Fatalf("Failed to create kafka producer with error: %v", err)
	}

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
		r.KafkaClient.Close()
		os.Exit(0)
	}()

	r.Scheduler()
}
