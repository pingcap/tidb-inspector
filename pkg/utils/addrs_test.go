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

package utils

import (
	"testing"

	. "github.com/pingcap/check"
)

func TestClient(t *testing.T) {
	TestingT(t)
}

var _ = Suite(&testAddrs{})

type testAddrs struct{}

func (*testAddrs) TestParseHostPortAddr(c *C) {
	_, err := ParseHostPortAddr(" 172.16.10.71:4000, 172.16.10.72:4000 ")
	c.Assert(err, IsNil)

	_, err = ParseHostPortAddr("172.16.10.71,172.16.10.72:4000")
	c.Assert(err, NotNil)
}
