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
	"log"
	"net/url"
	"strings"
)

// Panel represents a Grafana dashboard panel
type Panel struct {
	ID    int
	Type  string
	Title string
}

// Row represents a container for Panels
type Row struct {
	ID        int
	Showtitle bool
	Title     string
	Panels    []Panel
}

// Dashboard represents a Grafana dashboard
// This is both used to unmarshal the dashbaord JSON into
// and then enriched (sanitize fields for TeX consumption and add VarialbeValues)
type Dashboard struct {
	Title          string
	Description    string
	VariableValues string //Not present in the Grafana JSON structure. Enriched data passed used by the Tex templating
	Rows           []Row
	Panels         []Panel
}

type dashContainer struct {
	Dashboard Dashboard
	Meta      struct {
		Slug string
	}
}

// NewDashboard creates Dashboard from Grafana's internal JSON dashboard definition
func NewDashboard(dashJSON []byte, variables url.Values) Dashboard {
	var dash dashContainer
	err := json.Unmarshal(dashJSON, &dash)
	if err != nil {
		panic(err)
	}
	d := dash.NewDashboard(variables)
	log.Printf("Populated dashboard datastructure: %+v\n", d)
	return d
}

func (dc dashContainer) NewDashboard(variables url.Values) Dashboard {
	var dash Dashboard
	dash.Title = dc.Dashboard.Title
	dash.Description = dc.Dashboard.Description
	dash.VariableValues = getVariablesValues(variables)

	if len(dc.Dashboard.Rows) == 0 {
		return populatePanelsFromV5JSON(dash, dc)
	}
	return populatePanelsFromV4JSON(dash, dc)
}

func populatePanelsFromV4JSON(dash Dashboard, dc dashContainer) Dashboard {
	for _, row := range dc.Dashboard.Rows {
		for i, p := range row.Panels {
			row.Panels[i] = p
			dash.Panels = append(dash.Panels, p)
		}
		dash.Rows = append(dash.Rows, row)
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

// IsSingleStat ... check if singlestat
func (p Panel) IsSingleStat() bool {
	if p.Type == "singlestat" {
		return true
	}
	return false
}

// IsVisible ... check if visible
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
