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

/*
   Copyright 2016 Vastech SA (PTY) LTD

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/ngaut/log"
	"github.com/pingcap/tidb-inspect-tools/grafana_collector/config"
	"github.com/pingcap/tidb-inspect-tools/grafana_collector/grafana"
	"github.com/pingcap/tidb-inspect-tools/grafana_collector/report"
	"github.com/pingcap/tidb-inspect-tools/pkg/utils"
)

var (
	proto        = flag.String("proto", "http://", "Grafana Protocol")
	ip           = flag.String("ip", "localhost:3000", "Grafana IP and port")
	port         = flag.String("port", ":8686", "Port to serve on")
	logFile      = flag.String("log-file", "", "log file path")
	logLevel     = flag.String("log-level", "info", "log level: debug, info, warn, error, fatal")
	logRotate    = flag.String("log-rotate", "day", "log file rotate type: hour/day")
	configFile   = flag.String("config", "", "path to configuration file")
	fontDir      = flag.String("font-dir", "", "ttf fonts directory")
	printVersion = flag.Bool("V", false, "prints version and exit")
)

func main() {
	flag.Parse()

	if *printVersion {
		fmt.Println(utils.GetRawInfo("grafana_collector"))
		return
	}

	err := config.SetConfig(*configFile)
	if err != nil {
		log.Fatalf("parsing configure file error: %v", err)
	}

	if *fontDir == "" {
		log.Fatalf("missing parameter: -font-dir")
	}
	report.SetFontDir(*fontDir)

	log.SetLevelByString(*logLevel)
	if *logFile != "" {
		log.SetOutputByName(*logFile)
		if *logRotate == "hour" {
			log.SetRotateByHour()
		} else {
			log.SetRotateByDay()
		}
	}

	log.Infof("serving at '%s' and using grafana at '%s%s'", *port, *proto, *ip)

	router := mux.NewRouter()
	RegisterHandlers(
		router,
		ServeReportHandler{grafana.NewV4Client, report.New},
		ServeReportHandler{grafana.NewV5Client, report.New},
	)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		syscall.SIGKILL,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		sig := <-sc
		log.Infof("got signal [%d] to exit.", sig)
		os.Exit(0)
	}()

	log.Fatal(http.ListenAndServe(*port, router))

}
