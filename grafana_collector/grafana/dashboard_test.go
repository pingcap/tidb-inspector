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
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestV4Dashboard(t *testing.T) {
	Convey("When creating a new dashboard from Grafana v4 dashboard JSON", t, func() {
		const v4DashJSON = `
{"Dashboard":
	{
		"Rows":
			[{
				"Panels":
					[{"Type":"singlestat", "ID":1},
					{"Type":"graph", "ID":2}],
				"Title": "RowTitle #"
			},
			{"Panels":
				[{"Type":"singlestat", "ID":3, "Title": "Panel3Title #"}]
			}],
		"title":"DashTitle #"
	},
"Meta":
	{"Slug":"testDash"}
}`
		dash := NewDashboard([]byte(v4DashJSON), url.Values{})

		Convey("Panel IsSingelStat should work for all panels", func() {
			So(dash.Panels[0].IsSingleStat(), ShouldBeTrue)
			So(dash.Panels[1].IsSingleStat(), ShouldBeFalse)
			So(dash.Panels[2].IsSingleStat(), ShouldBeTrue)
		})

		Convey("Panels should contain all panels from all rows", func() {
			So(dash.Panels, ShouldHaveLength, 3)
		})
	})
}

func TestV5Dashboard(t *testing.T) {
	Convey("When creating a new dashboard from Grafana v5 dashboard JSON", t, func() {
		const v5DashJSON = `
{"Dashboard":
	{
		"Panels":
			[{"Type":"singlestat", "ID":0},
			{"Type":"graph", "ID":1},
			{"Type":"singlestat", "ID":2, "Title":"Panel3Title #"},
			{"Type":"row", "ID":3}],
		"Title":"DashTitle #"
	},

"Meta":
	{"Slug":"testDash"}
}`
		dash := NewDashboard([]byte(v5DashJSON), url.Values{})

		Convey("Panel IsSingelStat should work for all panels", func() {
			So(dash.Panels[0].IsSingleStat(), ShouldBeTrue)
			So(dash.Panels[1].IsSingleStat(), ShouldBeFalse)
			So(dash.Panels[2].IsSingleStat(), ShouldBeTrue)
		})

		Convey("Panels should contain all panels that have type != row", func() {
			So(dash.Panels, ShouldHaveLength, 3)
			So(dash.Panels[0].ID, ShouldEqual, 0)
			So(dash.Panels[1].ID, ShouldEqual, 1)
			So(dash.Panels[2].ID, ShouldEqual, 2)
		})
	})
}

func TestVariableValues(t *testing.T) {
	Convey("When creating a dashboard and passing url varialbes in", t, func() {
		const v5DashJSON = `
{
	"Dashboard":
		{
		}
}`
		vars := url.Values{}
		vars.Add("var-one", "oneval")
		vars.Add("var-two", "twoval")
		dash := NewDashboard([]byte(v5DashJSON), vars)

		Convey("The dashboard should contain the variable values in a random order", func() {
			So(dash.VariableValues, ShouldContainSubstring, "oneval")
			So(dash.VariableValues, ShouldContainSubstring, "twoval")
		})
	})
}
