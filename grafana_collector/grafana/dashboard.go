package grafana

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/ngaut/log"
)

// Panel represents a Grafana dashboard panel
type Panel struct {
	ID         int
	Type       string // Panel Type: Graph/Singlestat
	Title      string
	RowTitle   string
	ScopedVars map[string]ScopedVar
}

// ScopedVar represents template vars
type ScopedVar struct {
	Text  string
	Value string
}

// Row represents a container for Panels
type Row struct {
	ID              int
	Showtitle       bool // Row is visible or hidden
	Title           string
	Repeat          string
	RepeatIteration int
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
	Title          string
	Templating     map[string][]TemplatingVariable
	Rows           []Row
	Panels         []Panel
	VariableValues string
	url            string
	apiToken       string
	Iteration      int
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

func unique(input []string) []string {
	result := make([]string, 0, len(input))
	m := make(map[string]bool)

	for _, item := range input {
		if _, ok := m[item]; !ok {
			m[item] = true
			result = append(result, item)
		}
	}

	return result
}

func (d *Dashboard) getTemplatingVariable(tv TemplatingVariable) []string {
	re := regexp.MustCompile(`label_values\((\w+),\s*(\w+)\)$`)
	matched := re.FindStringSubmatch(tv.Query)
	metric := matched[1]
	label := matched[2]

	metricURL := fmt.Sprintf("%s/api/datasources/proxy/1/api/v1/series?match[]=%s&start=1543386918&end=1543390518", d.url, metric)

	log.Infof("request metric at %s", metricURL)

	clientTimeout := time.Duration(300) * time.Second
	client := &http.Client{Timeout: clientTimeout}
	req, _ := http.NewRequest("GET", metricURL, nil)

	if d.apiToken != "" {
		req.Header.Add("Authorization", "Bearer "+d.apiToken)
	}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var result MetircResult
	json.Unmarshal(body, &result)

	labelResult := make([]string, 0, len(result.Data))
	for _, m := range result.Data {
		s, ok := m[label].(string)
		if ok {
			labelResult = append(labelResult, s)
		}
	}

	uniqueLabelResult := unique(labelResult)
	sort.Strings(uniqueLabelResult)
	return uniqueLabelResult
}

func (d *Dashboard) process() {
	if len(d.Templating["list"]) == 0 {
		return
	}

	for i := 0; i < len(d.Rows); i++ {
		row := d.Rows[i]
		if row.Repeat != "" {
			d.repeatRow(row, i)
		}
	}
}

func (d *Dashboard) removeRow(rowIndex int) {
	d.Rows = append(d.Rows[:rowIndex], d.Rows[rowIndex+1:]...)
}

func (d *Dashboard) repeatRow(row Row, rowIndex int) {
	var (
		exist    bool
		selected []string
	)

	repeatValue := row.Repeat
	for _, tv := range d.Templating["list"] {
		if tv.Name == repeatValue {
			exist = true
			selected = d.getTemplatingVariable(tv)
		}
	}

	if !exist {
		return
	}

	for index, option := range selected {
		duplicate := d.getRowClone(row, index, rowIndex)

		for i := 0; i < len(duplicate.Panels); i++ {
			panel := duplicate.Panels[i]
			panel.ScopedVars = map[string]ScopedVar{repeatValue: {Text: option, Value: option}}
		}
	}
}

func (d *Dashboard) getRowClone(sourceRow Row, repeatIndex int, sourceRowIndex int) Row {
	if repeatIndex == 0 {
		return sourceRow
	}

	sourceRowID := sourceRowIndex + 1

	repeat := Row{}
	repeat.Repeat = ""
	repeat.RepeatRowID = sourceRowID
	repeat.RepeatIteration = d.Iteration
	repeat.Panels = make([]Panel, len(sourceRow.Panels))
	copy(repeat.Panels, sourceRow.Panels)

	d.Rows = append(d.Rows, Row{})
	copy(d.Rows[sourceRowIndex+repeatIndex+1:], d.Rows[sourceRowIndex+repeatIndex:])
	d.Rows[sourceRowIndex+repeatIndex] = repeat

	for i := range d.Rows[sourceRowIndex+repeatIndex].Panels {
		d.Rows[sourceRowIndex+repeatIndex].Panels[i].ID = d.getNextPanelID()
	}

	return repeat
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
func NewDashboard(dashJSON []byte, url string, apiToken string, variables url.Values) Dashboard {
	var dash dashContainer
	err := json.Unmarshal(dashJSON, &dash)
	if err != nil {
		panic(err)
	}
	d := dash.NewDashboard(url, apiToken, variables)
	log.Infof("Populated dashboard datastructure: %+v\n", d)
	return d
}

func (dc dashContainer) NewDashboard(url string, apiToken string, variables url.Values) Dashboard {
	var dash Dashboard
	iteration := int(time.Now().UnixNano() / int64(time.Millisecond))

	dash.Title = dc.Dashboard.Title
	dash.Templating = dc.Dashboard.Templating
	dash.VariableValues = getVariablesValues(variables)
	dash.url = url
	dash.apiToken = apiToken
	dash.Iteration = iteration

	if len(dc.Dashboard.Rows) == 0 {
		return populatePanelsFromV5JSON(dash, dc)
	}
	return populatePanelsFromV4JSON(dash, dc)
}

func populatePanelsFromV4JSON(dash Dashboard, dc dashContainer) Dashboard {
	for _, row := range dc.Dashboard.Rows {
		dash.Rows = append(dash.Rows, row)
	}

	dash.process()
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

// IsSingleStat ... checks if panel is singlestat
func (p Panel) IsSingleStat() bool {
	if p.Type == "singlestat" {
		return true
	}
	return false
}

// IsVisible ... checks if row is visible
func (r Row) IsVisible() bool {
	return r.Showtitle
}

func getVariablesValues(variables url.Values) string {
	values := []string{}
	for _, v := range variables {
		values = append(values, strings.Join(v, ", "))
	}
	return strings.Join(values, ", ")
}
