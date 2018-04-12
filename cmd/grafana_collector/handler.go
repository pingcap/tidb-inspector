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
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"github.com/ngaut/log"
	"github.com/pingcap/tidb-inspect-tools/grafana_collector/grafana"
	"github.com/pingcap/tidb-inspect-tools/grafana_collector/report"
)

// ServeReportHandler generates grafana dashboard pdf file and returns to client
type ServeReportHandler struct {
	newGrafanaClient func(url string, apiToken string, variables url.Values) grafana.Client
	newReport        func(g grafana.Client, dashName string, time grafana.TimeRange) report.Report
}

// RegisterHandlers registers all http.Handler with their associated routes to the router
// Two different serve report handlers are used to provide support for both Grafana v4 (and older) and v5 APIs
func RegisterHandlers(router *mux.Router, reportServerV4, reportServerV5 ServeReportHandler) {
	router.Handle("/api/report/{dashId}", reportServerV4)
	router.Handle("/api/v5/report/{dashId}", reportServerV5)
}

func (h ServeReportHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Info("Reporter called")
	g := h.newGrafanaClient(*proto+*ip, apiToken(req), dashVariables(req))
	rep := h.newReport(g, dashID(req), time(req))

	file, err := rep.Generate()
	if err != nil {
		log.Errorf("Error generating report: %v", err)
		http.Error(w, err.Error(), 500)
		return
	}
	defer rep.Clean()
	defer file.Close()

	_, err = io.Copy(w, file)
	if err != nil {
		log.Errorf("Error copying data to response: %v", err)
		http.Error(w, err.Error(), 500)
		return
	}
	log.Info("Report generated correctly")
}

func dashID(r *http.Request) string {
	vars := mux.Vars(r)
	d := vars["dashId"]
	log.Infof("Called with dashboard: %s", d)
	return d
}

func time(r *http.Request) grafana.TimeRange {
	params := r.URL.Query()
	t := grafana.NewTimeRange(params.Get("from"), params.Get("to"))
	log.Infof("Called with time range: %v", t)
	return t
}

func apiToken(r *http.Request) string {
	apiToken := r.URL.Query().Get("apitoken")
	log.Infof("Called with API Token: %s", apiToken)
	return apiToken
}

func dashVariables(r *http.Request) url.Values {
	output := url.Values{}
	for k, v := range r.URL.Query() {
		if strings.HasPrefix(k, "var-") {
			log.Infof("Called with variable: %s: %v", k, v)
			for _, singleV := range v {
				output.Add(k, singleV)
			}
		}
	}
	if len(output) == 0 {
		log.Info("Called without variable")
	}
	return output
}
