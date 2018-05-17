package main

import (
	"github.com/BurntSushi/toml"
	"github.com/juju/errors"
)

var (
	probeConfig = &Config{}
)

// SetConfig ... parses tcp_prober template configure file
func SetConfig(configFile string) error {
	_, err := toml.DecodeFile(configFile, probeConfig)
	return errors.Trace(err)
}

// Config ... tcp_prober template config
type Config struct {
	Service map[string]service
}

type service struct {
	Addr      string
	Alertname string
	Summary   string
	Level     string
}
