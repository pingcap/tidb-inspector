package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
	"github.com/ngaut/log"
	"github.com/unrolled/render"
)

var (
	port         = flag.Int("port", 28082, "port to listen on for the web interface")
	kafkaAddress = flag.String("kafka-address", "", "kafka address, example: 10.0.3.4:9092,10.0.3.5:9092,10.0.3.6:9092")
	kafkaTopic   = flag.String("kafka-topic", "", "kafka topic")
	logFile      = flag.String("log-file", "", "log file path")
	logLevel     = flag.String("log-level", "info", "log level: debug, info, warn, error, fatal")
	logRotate    = flag.String("log-rotate", "day", "log file rotate type: hour/day")
)

//KafkaMsg represents kafka message
type KafkaMsg struct {
	Title       string `json:"title"`
	Source      string `json:"source"`
	Node        string `json:"node"`
	Expr        string `json:"expr"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Level       string `json:"level"`
	Note        string `json:"note"`
	Value       string `json:"value"`
	Time        string `json:"time"`
}

//Run represents runtime information
type Run struct {
	Rdr         *render.Render
	AlertMsgs   chan *AlertData
	KafkaClient sarama.SyncProducer
}

//Scheduler for monitoring chan data
func (r *Run) Scheduler() {
	for {
		lenAlertMsgs := len(r.AlertMsgs)
		if lenAlertMsgs > 0 {
			for i := 0; i < lenAlertMsgs; i++ {
				r.TransferData(<-r.AlertMsgs)
			}
		}
		time.Sleep(3 * time.Second)
	}
}

func main() {
	flag.Parse()
	if *kafkaAddress == "" {
		log.Fatalf("missing parameter: -kafka-address")
	}
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
	if err := r.CreateKafkaProducer(); err != nil {
		log.Errorf("create kafka producer with error %v", err)
		return
	}
	go r.Scheduler()

	log.Infof("create a http server serving at %s", *port)
	r.CreateRender()
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), r.CreateRouter()))

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
		log.Errorf("Got signal [%d] to exit.", sig)
		r.KafkaClient.Close()
		wg.Done()
	}()

	wg.Wait()

}
