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

package grafana

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGrafanaClientFetchesDashboard(t *testing.T) {
	Convey("When fetching a Dashboard", t, func() {
		requestURI := ""
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestURI = r.RequestURI
			fmt.Fprintln(w, `{"":""}`)
		}))
		defer ts.Close()

		timeRange := TimeRange{"now-1h", "now"}
		Convey("When using the Grafana v4 client", func() {
			grf := NewV4Client(ts.URL, "", timeRange)
			grf.GetDashboard("testDash")

			Convey("It should use the v4 dashboards endpoint", func() {
				So(requestURI, ShouldEqual, "/api/dashboards/db/testDash")
			})
		})

		Convey("When using the Grafana v5 client", func() {
			grf := NewV5Client(ts.URL, "", timeRange)
			grf.GetDashboard("rYy7Paekz")

			Convey("It should use the v5 dashboards endpoint", func() {
				So(requestURI, ShouldEqual, "/api/dashboards/uid/rYy7Paekz")
			})
		})

	})
}

func TestGrafanaClientFetchesPanelPNG(t *testing.T) {
	Convey("When fetching a panel PNG", t, func() {
		requestURI := ""
		requestHeaders := http.Header{}

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestURI = r.RequestURI
			requestHeaders = r.Header
		}))
		defer ts.Close()

		apiToken := "1234"
		timeRange := TimeRange{"now-1h", "now"}

		cases := map[string]struct {
			client      Client
			pngEndpoint string
		}{
			"v4": {NewV4Client(ts.URL, apiToken, timeRange), "/render/dashboard-solo/db/testDash"},
			"v5": {NewV5Client(ts.URL, apiToken, timeRange), "/render/d-solo/testDash/_"},
		}
		for clientDesc, cl := range cases {
			grf := cl.client
			grf.GetPanelPng(Panel{44, "singlestat", "title", "rowtitle", nil}, "testDash", TimeRange{"now-1h", "now"})

			Convey(fmt.Sprintf("The %s client should use the render endpoint with the dashboard name", clientDesc), func() {
				So(requestURI, ShouldStartWith, cl.pngEndpoint)
			})

			Convey(fmt.Sprintf("The %s client should request the panel ID", clientDesc), func() {
				So(requestURI, ShouldContainSubstring, "panelId=44")
			})

			Convey(fmt.Sprintf("The %s client should request the time", clientDesc), func() {
				So(requestURI, ShouldContainSubstring, "from=now-1h")
				So(requestURI, ShouldContainSubstring, "to=now")
			})

			Convey(fmt.Sprintf("The %s client should render singlestat panels should request a smaller size", clientDesc), func() {
				So(requestURI, ShouldContainSubstring, "width=480")
				So(requestURI, ShouldContainSubstring, "height=93")
			})

			Convey(fmt.Sprintf("The %s client should insert auth token should in request header", clientDesc), func() {
				So(requestHeaders.Get("Authorization"), ShouldContainSubstring, apiToken)
			})

			Convey(fmt.Sprintf("The %s client should request other panels in a larger size", clientDesc), func() {
				grf.GetPanelPng(Panel{44, "graph", "title", "rowtitle", nil}, "testDash", TimeRange{"now", "now-1h"})
				So(requestURI, ShouldContainSubstring, "width=1000")
				So(requestURI, ShouldContainSubstring, "height=500")
			})
		}

	})
}

func TestGrafanaClientFetchPanelPNGErrorHandling(t *testing.T) {
	Convey("When trying to fetching a panel from the server sometimes returns an error", t, func() {
		try := 0

		//create a server that will return error on the first call
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if try < 1 {
				w.WriteHeader(http.StatusInternalServerError)
				try++
			}
		}))
		defer ts.Close()

		grf := NewV4Client(ts.URL, "", TimeRange{"now-1h", "now"})

		_, err := grf.GetPanelPng(Panel{44, "singlestat", "title", "rowtitle", nil}, "testDash", TimeRange{"now-1h", "now"})

		Convey("It should retry a couple of times if it receives errors", func() {
			So(err, ShouldBeNil)
		})
	})

	Convey("When trying to fetching a panel from the server consistently returns an error", t, func() {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		grf := NewV4Client(ts.URL, "", TimeRange{"now-1h", "now"})

		_, err := grf.GetPanelPng(Panel{44, "singlestat", "title", "rowtitle", nil}, "testDash", TimeRange{"now-1h", "now"})

		Convey("The Grafana API should return an error", func() {
			So(err, ShouldNotBeNil)
		})
	})
}
