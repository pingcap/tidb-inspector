package config

import (
	"github.com/BurntSushi/toml"
	"github.com/juju/errors"
)

// variables for rendering pdf
var (
	Cfg = &Config{}
)

// SetConfig ... parses pdf template configure file
func SetConfig(configFile string) error {
	_, err := toml.DecodeFile(configFile, Cfg)
	return errors.Trace(err)
}

// Config ... pdf template config
type Config struct {
	Grafana  grafana
	Font     font
	Rect     map[string]rect
	Position position
}

type grafana struct {
	Theme         string
	ClientTimeout int `toml:"client-timeout"`
	ServerTimeout int `toml:"server-timeout"`
	RetryInterval int `toml:"retry-interval"`
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
