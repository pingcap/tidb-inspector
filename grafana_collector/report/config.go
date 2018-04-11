package report

import (
	"github.com/BurntSushi/toml"
	"github.com/juju/errors"
)

// variables for rendering pdf
var (
	ReportConfig = &tomlConfig{}
	FontDir      = ""
)

// NewConfig ... parses pdf template configure file
func NewConfig(configFile string) error {
	if _, err := toml.DecodeFile(configFile, ReportConfig); err != nil {
		return errors.Trace(err)
	}
	return nil
}

// NewFontDir ... sets up ttf font directory
func NewFontDir(fontDir string) {
	FontDir = fontDir
}

type tomlConfig struct {
	Font     font
	Rect     map[string]rect
	Position position
}

type font struct {
	Family string
	Ttf    string
	Size   int
}

type rect struct {
	Width  float64
	Height float64
}

type position struct {
	X  float64
	Y1 float64
	Y2 float64
	Br float64
}
