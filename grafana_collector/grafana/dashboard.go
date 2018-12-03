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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/ngaut/log"
)

var (
	queryMetricRegexp = regexp.MustCompile(`^label_values\((.+),\s*([a-zA-Z_][a-zA-Z0-9_]+)\)$`)
	variableRegexp    = regexp.MustCompile(`\$.+`)
)

// ScopedVar represents template variable
type ScopedVar struct {
	Text  string
	Value string
}

// Panel represents a Grafana dashboard panel
type Panel struct {
	ID         int
	Type       string // Panel Type: Graph/Singlestat
	Title      string
	RowTitle   string
	ScopedVars map[string]ScopedVar
}

// Row represents a container for Panels
type Row struct {
	ID              int
	Showtitle       bool // Row is visible or hidden
	Title           string
	Repeat          string
	RepeatIteration int64
	RepeatRowID     int
	Panels          []Panel
}

// TemplatingVariable represents templating variable
type TemplatingVariable struct {
	Name       string
	Datasource string
	Query      string
}

// Dashboard represents a Grafana dashboard
// This is used to unmarshal the dashbaord JSON
type Dashboard struct {
	Title      string
	Templating map[string][]TemplatingVariable
	Rows       []Row
	Panels     []Panel
	url        string
	apiToken   string
	timeRange  TimeRange
	iteration  int64
}

type dashContainer struct {
	Dashboard Dashboard
	Meta      struct {
		Slug string
	}
}

// MetircResult represents templating variable metric result
type MetircResult struct {
	Status string
	Data   []map[string]interface{}
}

func handleMetircResult(mr MetircResult, label string) []string {
	result := make([]string, 0, len(mr.Data))
	filter := make(map[string]bool)

	for _, m := range mr.Data {
		if v, ok := m[label].(string); ok {
			if _, exist := filter[v]; !exist {
				filter[v] = true
				result = append(result, v)
			}
		}
	}

	sort.Strings(result)
	return result
}

func getMetricAndLabel(tv TemplatingVariable) (string, string, error) {
	matched := queryMetricRegexp.FindStringSubmatch(tv.Query)
	metric := matched[1]
	label := matched[2]

	if metric != "" && label != "" {
		return metric, label, nil
	}

	return "", "", fmt.Errorf("%s should be label_values(metric, label) format", tv.Query)
}

func (d *Dashboard) getTemplatingVariableValue(tv TemplatingVariable) ([]string, error) {
	metric, label, err := getMetricAndLabel(tv)
	if err != nil {
		return nil, err
	}

	metricURL := fmt.Sprintf("%s/api/datasources/proxy/1/api/v1/series?match[]=%s&start=%d&end=%d", d.url, metric, d.timeRange.FromToUnix(), d.timeRange.ToToUnix())

	log.Infof("request metric at %s\n", metricURL)

	clientTimeout := time.Duration(60) * time.Second
	client := &http.Client{Timeout: clientTimeout}
	req, err := http.NewRequest("GET", metricURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating metric request for %v error: %v", metricURL, err)
	}

	if d.apiToken != "" {
		req.Header.Add("Authorization", "Bearer "+d.apiToken)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing metric request for %v error: %v", metricURL, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading metric response body from %v error: %v", metricURL, err)
	}

	var result MetircResult
	err = json.Unmarshal(body, &result)

	if err != nil {
		return nil, fmt.Errorf("unmarshaling metric response from %v error: %v", metricURL, err)
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("metric request status from %v is not success: %v", metricURL, err)
	}

	return handleMetircResult(result, label), nil
}

func (d *Dashboard) process() {
	for i := 0; i < len(d.Rows); i++ {
		row := d.Rows[i]
		if row.Repeat != "" {
			d.repeatRow(row, i)
		} else if row.RepeatRowID != 0 && row.RepeatIteration != d.iteration {
			d.removeRow(i)
			i = i - 1
			continue
		} else {
			rowTitle := row.Title
			for j := range row.Panels {
				d.Rows[i].Panels[j].RowTitle = rowTitle
			}
		}
	}
}

func (d *Dashboard) removeRow(rowIndex int) {
	d.Rows = append(d.Rows[:rowIndex], d.Rows[rowIndex+1:]...)
}

func (d *Dashboard) repeatRow(row Row, rowIndex int) {
	var (
		err      error
		exist    bool
		selected []string
	)

	label := row.Repeat
	for _, tv := range d.Templating["list"] {
		if tv.Name == label {
			exist = true
			selected, err = d.getTemplatingVariableValue(tv)
			if err != nil {
				log.Errorf("getting templateing varaible value error: %v\n", err)
			}
			break
		}
	}

	if !exist {
		return
	}

	for index, option := range selected {
		d.getRowClone(row, index, rowIndex, label, option)
	}
}

func (d *Dashboard) getRowClone(sourceRow Row, repeatIndex int, sourceRowIndex int, label string, option string) {
	rowTitle := sourceRow.Title
	matched := variableRegexp.FindString(rowTitle)
	if matched != "" {
		if strings.TrimPrefix(matched, "$") == label {
			rowTitle = variableRegexp.ReplaceAllString(rowTitle, option)
		}
	}

	if repeatIndex == 0 {
		d.Rows[sourceRowIndex].Title = rowTitle
		for i := range d.Rows[sourceRowIndex].Panels {
			d.Rows[sourceRowIndex].Panels[i].RowTitle = rowTitle
			d.Rows[sourceRowIndex].Panels[i].ScopedVars = map[string]ScopedVar{label: {Text: option, Value: option}}
		}
		return
	}

	sourceRowID := sourceRowIndex + 1

	for i := 0; i < len(d.Rows); i++ {
		row := d.Rows[i]
		if row.RepeatRowID == sourceRowID && row.RepeatIteration != d.iteration {
			d.Rows[i].Title = rowTitle
			d.Rows[i].Repeat = ""
			d.Rows[i].RepeatRowID = sourceRowID
			d.Rows[i].RepeatIteration = d.iteration
			return
		}
	}

	repeat := Row{}
	repeat.Repeat = ""
	repeat.RepeatRowID = sourceRowID
	repeat.RepeatIteration = d.iteration
	repeat.Panels = make([]Panel, len(sourceRow.Panels))
	copy(repeat.Panels, sourceRow.Panels)

	d.Rows = append(d.Rows, Row{})
	copy(d.Rows[sourceRowIndex+repeatIndex+1:], d.Rows[sourceRowIndex+repeatIndex:])
	d.Rows[sourceRowIndex+repeatIndex] = repeat
	d.Rows[sourceRowIndex+repeatIndex].Title = rowTitle

	for i := range d.Rows[sourceRowIndex+repeatIndex].Panels {
		d.Rows[sourceRowIndex+repeatIndex].Panels[i].ID = d.getNextPanelID()
		d.Rows[sourceRowIndex+repeatIndex].Panels[i].RowTitle = rowTitle
		d.Rows[sourceRowIndex+repeatIndex].Panels[i].ScopedVars = map[string]ScopedVar{label: {Text: option, Value: option}}
	}
}

func (d *Dashboard) getNextPanelID() int {
	max := 0
	for _, row := range d.Rows {
		for _, panel := range row.Panels {
			if panel.ID > max {
				max = panel.ID
			}
		}
	}
	return max + 1
}

// NewDashboard creates Dashboard from Grafana's internal JSON dashboard definition
func NewDashboard(dashJSON []byte, url string, apiToken string, timeRange TimeRange) Dashboard {
	var dash dashContainer
	err := json.Unmarshal(dashJSON, &dash)
	if err != nil {
		panic(err)
	}
	d := dash.NewDashboard(url, apiToken, timeRange)

	b, err := json.MarshalIndent(d, "", "    ")
	if err != nil {
		log.Errorf("marshaling populated dashboard error: %v\n", err)
	}
	log.Infof("populated dashboard datastructure: %s\n", string(b))
	return d
}

func (dc dashContainer) NewDashboard(url string, apiToken string, timeRange TimeRange) Dashboard {
	var dash Dashboard
	iteration := UnixSecond(time.Now())

	dash.Title = dc.Dashboard.Title
	dash.Templating = dc.Dashboard.Templating
	dash.url = url
	dash.apiToken = apiToken
	dash.timeRange = timeRange
	dash.iteration = iteration

	if len(dc.Dashboard.Rows) == 0 {
		return populatePanelsFromV5JSON(dash, dc)
	}
	return populatePanelsFromV4JSON(dash, dc)
}

func populatePanelsFromV4JSON(dash Dashboard, dc dashContainer) Dashboard {
	for _, row := range dc.Dashboard.Rows {
		dash.Rows = append(dash.Rows, row)
	}

	// handle Row repeats and RowTitle
	dash.process()

	for _, row := range dash.Rows {
		for _, p := range row.Panels {
			dash.Panels = append(dash.Panels, p)
		}
	}
	return dash
}

func populatePanelsFromV5JSON(dash Dashboard, dc dashContainer) Dashboard {
	for _, p := range dc.Dashboard.Panels {
		if p.Type == "row" {
			continue
		}
		dash.Panels = append(dash.Panels, p)
	}
	return dash
}

// IsSingleStat ... checks if Panel is singlestat
func (p Panel) IsSingleStat() bool {
	if p.Type == "singlestat" {
		return true
	}
	return false
}

// IsVisible ... checks if Row is visible
func (r Row) IsVisible() bool {
	return r.Showtitle
}
