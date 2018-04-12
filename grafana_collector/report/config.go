package report

import (
	"github.com/BurntSushi/toml"
	"github.com/juju/errors"
)

// variables for rendering pdf
var (
	ReportConfig = &Config{}
	FontDir      = ""
)

// SetConfig ... parses pdf template configure file
func SetConfig(configFile string) error {
	_, err := toml.DecodeFile(configFile, ReportConfig)
	return errors.Trace(err)
}

// SetFontDir ... sets up ttf font directory
func SetFontDir(fontDir string) {
	FontDir = fontDir
}

// Config ... pdf template config
type Config struct {
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
