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
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/ngaut/log"
	"github.com/pingcap/tidb-inspect-tools/grafana_collector/config"
	"github.com/pkg/errors"
)

var (
	cfg = config.GetGlobalConfig()
)

// Client is a Grafana API client
type Client interface {
	GetDashboard(dashName string) (Dashboard, error)
	GetPanelPng(p Panel, dashName string, t TimeRange) (io.ReadCloser, error)
}

type client struct {
	url              string
	getDashEndpoint  func(dashName string) string
	getPanelEndpoint func(dashName string, vals url.Values) string
	apiToken         string
	timeRange        TimeRange
}

// NewV4Client creates a new Grafana 4 Client. If apiToken is the empty string,
// authorization headers will be omitted from requests.
// variables are Grafana template variable url values of the form
// var-{name}={value}, e.g. var-host=dev
func NewV4Client(grafanaURL string, apiToken string, timeRange TimeRange) Client {
	getDashEndpoint := func(dashName string) string {
		dashURL := grafanaURL + "/api/dashboards/db/" + dashName
		return dashURL
	}

	getPanelEndpoint := func(dashName string, vals url.Values) string {
		return fmt.Sprintf("%s/render/dashboard-solo/db/%s?%s", grafanaURL, dashName, vals.Encode())
	}
	return client{grafanaURL, getDashEndpoint, getPanelEndpoint, apiToken, timeRange}
}

// NewV5Client creates a new Grafana 5 Client. If apiToken is the empty string,
// authorization headers will be omitted from requests.
// variables are Grafana template variable url values of the form
// var-{name}={value}, e.g. var-host=dev
func NewV5Client(grafanaURL string, apiToken string, timeRange TimeRange) Client {
	getDashEndpoint := func(dashName string) string {
		dashURL := grafanaURL + "/api/dashboards/uid/" + dashName
		return dashURL
	}

	getPanelEndpoint := func(dashName string, vals url.Values) string {
		return fmt.Sprintf("%s/render/d-solo/%s/_?%s", grafanaURL, dashName, vals.Encode())
	}
	return client{grafanaURL, getDashEndpoint, getPanelEndpoint, apiToken, timeRange}
}

func (g client) GetDashboard(dashName string) (Dashboard, error) {
	dashURL := g.getDashEndpoint(dashName)
	log.Infof("connecting to dashboard at %s", dashURL)

	clientTimeout := time.Duration(cfg.Grafana.ClientTimeout) * time.Second
	client := &http.Client{Timeout: clientTimeout}
	req, err := http.NewRequest("GET", dashURL, nil)
	if err != nil {
		return Dashboard{}, errors.Errorf("creating getDashboard request for %s error: %v", dashURL, err)
	}

	if g.apiToken != "" {
		req.Header.Add("Authorization", "Bearer "+g.apiToken)
	}
	resp, err := client.Do(req)
	if err != nil {
		return Dashboard{}, errors.Errorf("executing getDashboard request for %s error: %v", dashURL, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Dashboard{}, errors.Errorf("reading getDashboard response body from %s: %v error", dashURL, err)
	}

	if resp.StatusCode != 200 {
		return Dashboard{}, errors.Errorf("obtaining dashboard from %s error, got status %s, message: %s", dashURL, resp.Status, string(body))
	}

	return NewDashboard(body, g.url, g.apiToken, g.timeRange)
}

func (g client) GetPanelPng(p Panel, dashName string, t TimeRange) (io.ReadCloser, error) {
	panelURL := g.getPanelURL(p, dashName, t)

	clientTimeout := time.Duration(cfg.Grafana.ClientTimeout) * time.Second
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return errors.New("getting panel png error. Redirected to login")
		},
		Timeout: clientTimeout,
	}
	req, err := http.NewRequest("GET", panelURL, nil)
	if err != nil {
		return nil, errors.Errorf("creating getPanelPng request for %s error: %v", panelURL, err)
	}
	if g.apiToken != "" {
		req.Header.Add("Authorization", "Bearer "+g.apiToken)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Errorf("executing getPanelPng request for %s error: %v", panelURL, err)
	}

	for retries := 1; retries < 3 && resp.StatusCode != 200; retries++ {
		getPanelRetryInterval := time.Duration(cfg.Grafana.RetryInterval) * time.Second
		delay := getPanelRetryInterval * time.Duration(retries)
		log.Errorf("obtaining render for panel %+v error, status: %d, retrying after %v...", p, resp.StatusCode, delay)
		time.Sleep(delay)
		resp, err = client.Do(req)
		if err != nil {
			return nil, errors.Errorf("executing getPanelPng retry request for %s error: %v", panelURL, err)
		}
	}

	if resp.StatusCode != 200 {
		log.Errorf("obtaining panel image request from %s is not successful, status: %s", panelURL, resp.Status)
		return nil, errors.Errorf("obtaining panel image request from %s is not successful, status: %s", panelURL, resp.Status)
	}

	return resp.Body, nil
}

func (g client) getPanelURL(p Panel, dashName string, t TimeRange) string {
	values := url.Values{}
	values.Add("theme", cfg.Grafana.Theme)
	values.Add("panelId", strconv.Itoa(p.ID))
	values.Add("from", t.From)
	values.Add("to", t.To)
	if p.IsSingleStat() {
		values.Add("width", "480")
		values.Add("height", "93")
	} else {
		values.Add("width", "1000")
		values.Add("height", "500")
	}
	values.Add("timeout", strconv.Itoa(cfg.Grafana.ServerTimeout))

	url := g.getPanelEndpoint(dashName, values)
	log.Infof("downloading image: %d %s", p.ID, url)
	return url
}
