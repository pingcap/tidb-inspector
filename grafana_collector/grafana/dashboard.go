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
	"net/url"
	// "regexp"
	"strings"
	// "time"

	"github.com/ngaut/log"
)

var (
	templating = map[string][]string{"db": []string{"kv", "raft"}, "command": []string{"batch_get", "commit", "gc", "get", "prewrite", "scan", "scan_lock"}}
)

// Panel represents a Grafana dashboard panel
type Panel struct {
	ID       int
	Type     string // Panel Type: Graph/Singlestat
	Title    string
	RowTitle string
	// ScopedVars map[string]ScopedVar
}

// ScopedVar represents template vars
type ScopedVar struct {
	Selected bool
	Text     string
	Value    string
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

// Dashboard represents a Grafana dashboard
// This is used to unmarshal the dashbaord JSON
type Dashboard struct {
	Title          string
	VariableValues string //Not present in the Grafana JSON structure
	Rows           []Row
	Panels         []Panel
}

type dashContainer struct {
	Dashboard Dashboard
	Meta      struct {
		Slug string
	}
}

func (d *Dashboard) process() {
	if len(templating) == 0 {
		return
	}

	// iteration := int(time.Now().UnixNano() / int64(time.Millisecond))

	for i := 0; i < len(d.Rows); i++ {
		row := d.Rows[i]
		if row.Repeat != "" {
			d.repeatRow(row, i)
		}
	}
}

func (d *Dashboard) repeatRow(row Row, rowIndex int) {
	selected, ok := templating[row.Repeat]
	if !ok {
		return
	}

	for index, option := range selected {
		d.getRowClone(row, index, rowIndex)
	}
}

func (d *Dashboard) getRowClone(sourceRow Row, repeatIndex int, sourceRowIndex int) {
	if repeatIndex == 0 {
		return
	}

	sourceRowID := sourceRowIndex + 1

	repeat := Row{}
	repeat.Panels = make([]Panel, len(sourceRow.Panels))
	copy(repeat.Panels, sourceRow.Panels)

	repeat.RepeatRowID = sourceRowID
	// repeat.RepeatIteration = iteration

	d.Rows = append(d.Rows, Row{Title: "temp"})
	copy(d.Rows[sourceRowIndex+repeatIndex+1:], d.Rows[sourceRowIndex+repeatIndex:])
	d.Rows[sourceRowIndex+repeatIndex] = repeat

	for i := range d.Rows[sourceRowIndex+repeatIndex].Panels {
		d.Rows[sourceRowIndex+repeatIndex].Panels[i].ID = d.getNextPanelID()
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
func NewDashboard(dashJSON []byte, variables url.Values) Dashboard {
	var dash dashContainer
	err := json.Unmarshal(dashJSON, &dash)
	if err != nil {
		panic(err)
	}
	d := dash.NewDashboard(variables)
	log.Infof("Populated dashboard datastructure: %+v\n", d)
	return d
}

func (dc dashContainer) NewDashboard(variables url.Values) Dashboard {
	var dash Dashboard
	dash.Title = dc.Dashboard.Title
	dash.VariableValues = getVariablesValues(variables)

	if len(dc.Dashboard.Rows) == 0 {
		return populatePanelsFromV5JSON(dash, dc)
	}
	return populatePanelsFromV4JSON(dash, dc)
}

func populatePanelsFromV4JSON(dash Dashboard, dc dashContainer) Dashboard {
	// re := regexp.MustCompile(`\$\w+`)

	for _, row := range dc.Dashboard.Rows {
		// rowTitle := row.Title
		// matched := re.FindString(rowTitle)

		// for i, p := range row.Panels {
		// if matched != "" {
		//     for k, v := range p.ScopedVars {
		//         if strings.TrimPrefix(matched, "$") == k {
		//             rowTitle = re.ReplaceAllString(rowTitle, v.Value)
		//         }
		//     }
		// }
		// p.RowTitle = rowTitle
		// row.Panels[i] = p
		// dash.Panels = append(dash.Panels, p)
		// }
		dash.Rows = append(dash.Rows, row)
	}

	log.Errorf("888888888888: cap, %v\n", cap(dash.Rows))
	log.Errorf("888888888888: len, %v\n", len(dash.Rows))

	for _, row := range dash.Rows {
		log.Errorf("999999999999999, %v\n", row.Title)
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
