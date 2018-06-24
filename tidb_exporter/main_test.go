package main

import (
	"testing"

	. "github.com/pingcap/check"
)

func TestClient(t *testing.T) {
	TestingT(t)
}

var _ = Suite(&testExporter{})

type testExporter struct{}

func (*testExporter) TestParseHostPortAddr(c *C) {
	_, err := ParseHostPortAddr(" 172.16.10.71:4000, 172.16.10.72:4000 ")
	c.Assert(err, IsNil)

	_, err = ParseHostPortAddr("172.16.10.71,172.16.10.72:4000")
	c.Assert(err, NotNil)
}
