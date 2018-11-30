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
   Copyright 2018 Vastech SA (PTY) LTD

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
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ngaut/log"
	"github.com/pingcap/tidb-inspect-tools/grafana_collector/grafana"
	"github.com/pingcap/tidb-inspect-tools/grafana_collector/report"
)

// ServeReportHandler generates grafana dashboard pdf file and returns to client
type ServeReportHandler struct {
	newGrafanaClient func(url string, apiToken string, timeRange grafana.TimeRange) grafana.Client
	newReport        func(g grafana.Client, dashName string, timeRange grafana.TimeRange) report.Report
}

// RegisterHandlers registers all http.Handler with their associated routes to
// the router. Two different serve report handlers are used to provide support
// for both Grafana v4 (and older) and v5 APIs
func RegisterHandlers(router *mux.Router, reportServerV4, reportServerV5 ServeReportHandler) {
	router.Handle("/api/report/{dashId}", reportServerV4)
	router.Handle("/api/v5/report/{dashId}", reportServerV5)
}

func (h ServeReportHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Info("reporter called")
	grafanaClient := h.newGrafanaClient(*proto+*ip, apiToken(req), time(req))
	reporter := h.newReport(grafanaClient, dashID(req), time(req))

	file, err := reporter.Generate()
	if err != nil {
		log.Errorf("generating report error: %v", err)
		http.Error(w, err.Error(), 500)
		return
	}
	defer reporter.Clean()
	defer file.Close()

	_, err = io.Copy(w, file)
	if err != nil {
		log.Errorf("copying pdf data to response error: %v", err)
		http.Error(w, err.Error(), 500)
		return
	}
	log.Info("report generated correctly")
}

func dashID(r *http.Request) string {
	vars := mux.Vars(r)
	d := vars["dashId"]
	log.Infof("called with dashboard: %s", d)
	return d
}

func time(r *http.Request) grafana.TimeRange {
	params := r.URL.Query()
	t := grafana.NewTimeRange(params.Get("from"), params.Get("to"))
	log.Infof("called with time range: %v", t)
	return t
}

func apiToken(r *http.Request) string {
	apiToken := r.URL.Query().Get("apitoken")
	log.Infof("called with API Token: %s", apiToken)
	return apiToken
}
