package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ngaut/log"
)

var (
	port         = flag.Int("port", 28082, "port to listen on for the web interface")
	kafkaAddress = flag.String("kafka-address", "", "kafka address, example: 10.0.3.4:9092,10.0.3.5:9092,10.0.3.6:9092")
	kafkaTopic   = flag.String("kafka-topic", "", "kafka topic")
	logFile      = flag.String("log-file", "", "log file path")
	logLevel     = flag.String("log-level", "info", "log level: debug, info, warn, error, fatal")
	logRotate    = flag.String("log-rotate", "day", "log file rotate type: hour/day")
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

	log.SetLevelByString(*logLevel)
	if *logFile != "" {
		log.SetOutputByName(*logFile)
		if *logRotate == "hour" {
			log.SetRotateByHour()
		} else {
			log.SetRotateByDay()
		}
	}

	r := &Run{
		AlertMsgs: make(chan *AlertData, 1000),
	}

	if err := r.CreateKafkaProducer(addrs); err != nil {
		log.Fatalf("Failed to create kafka producer with error: %v", err)
	}

	go r.Scheduler()

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

	log.Infof("create a http server serving at %d", *port)
	r.CreateRender()
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), r.CreateRouter()))

}
