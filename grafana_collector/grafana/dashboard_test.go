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
		dash, err := NewDashboard([]byte(v4DashJSON), "", "", TimeRange{"now-1h", "now"})

		Convey("Panel IsSingelStat should work for all panels", func() {
			So(err, ShouldBeNil)
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
		dash, err := NewDashboard([]byte(v5DashJSON), "", "", TimeRange{"now-1h", "now"})

		Convey("Panel IsSingelStat should work for all panels", func() {
			So(err, ShouldBeNil)
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

func TestGetMetricAndLabel(t *testing.T) {
	Convey("When analysing a correct TemplatingVariable ", t, func() {
		variable := TemplatingVariable{"db", "test-cluster", "label_values(tikv_engine_block_cache_size_bytes, db)"}
		metric, label, err := getMetricAndLabel(variable)

		Convey("metric and label should not be empty and correct", func() {
			So(err, ShouldBeNil)
			So(metric, ShouldEqual, "tikv_engine_block_cache_size_bytes")
			So(label, ShouldEqual, "db")
		})
	})
}

func TestGetMetricAndLabelErrorHandling(t *testing.T) {
	Convey("When analysing a wrong TemplatingVariable", t, func() {
		v1 := TemplatingVariable{"db", "test-cluster", "label_values(tikv_engine_block_cache_size_bytes, 2db)"}
		v2 := TemplatingVariable{"db", "test-cluster", "db, db"}

		metric1, label1, err1 := getMetricAndLabel(v1)
		metric2, label2, err2 := getMetricAndLabel(v2)

		Convey("should return error", func() {
			So(err1, ShouldNotBeNil)
			So(metric1, ShouldBeEmpty)
			So(label1, ShouldBeEmpty)
			So(err2, ShouldNotBeNil)
			So(metric2, ShouldBeEmpty)
			So(label2, ShouldBeEmpty)
		})
	})
}
