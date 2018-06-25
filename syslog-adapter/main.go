package main

import (
	"flag"
	"fmt"
	"github.com/juju/errors"
	"github.com/ngaut/log"
	"github.com/unrolled/render"
	"log/syslog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	port          = flag.Int("port", 28082, "The port to listen on for HTTP requests from alertmanager.")
	syslogAddress = flag.String("syslog-address", "", "syslog server address, example: 172.16.10.50:514")
	network       = flag.String("network", "", "(tcp or udp): send messages to the syslog server using UDP or TCP. If not set, connect to the local syslog server.")
	logFile       = flag.String("log-file", "", "log file path")
	logLevel      = flag.String("log-level", "info", "log level: debug, info, warn, error, fatal")
	alertLevel    = flag.String("alert-level", "7", "log level")
	alertIP       = flag.String("alert-ip", "", "alert IP")
	alertHostname = flag.String("alert-hostname", "", "alert hostname")
)

func initLog() error {
	log.SetLevelByString(*logLevel)
	if *logFile != "" {
		return log.SetOutputByName(*logFile)
	}
	return nil
}

func checkParams() error {
	if *syslogAddress != "" && *network == "" {
		return errors.New("please add network parameter: tcp or udp")
	}

	if *alertIP == "" || *alertHostname == "" {
		return errors.New("please add default alert-ip and alert-hostname parameters")
	}
	return nil
}

//Run represent runtime information
type Run struct {
	Rdr       *render.Render
	AlertMsgs chan *AlertData
	sysLog    *syslog.Writer
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
	if err := checkParams(); err != nil {
		fmt.Printf("Params error: %v\n", err)
		return
	}

	if err := initLog(); err != nil {
		fmt.Printf("Init log file with error: %v", err)
		return
	}

	sysLog, err := syslog.Dial(*network, *syslogAddress,
		syslog.LOG_CRIT|syslog.LOG_USER, "syslog-adapter")
	if err != nil {
		log.Fatalf("Failed to connect to syslog service: %v", err)
	}

	r := &Run{
		AlertMsgs: make(chan *AlertData, 1000),
		sysLog:    sysLog,
	}

	go r.Scheduler()
	go func() {
		log.Infof("Run http server, listen at [%d].", *port)
		r.CreateRender()
		http.ListenAndServe(fmt.Sprintf(":%d", *port), r.CreateRouter())
	}()
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
		sysLog.Close()
		log.Error("Close connection to the syslog daemon.")
		wg.Done()
	}()
	wg.Wait()
}
