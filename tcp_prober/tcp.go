package main

import (
	"net"
	"time"

	"github.com/ngaut/log"
)

const (
	probeDialTimeout  = 5 * time.Second
	maxDialRetry      = 3
	retryDialInterval = 5 * time.Second
)

func dialTCP(target string) error {
	conn, err := net.DialTimeout("tcp", target, probeDialTimeout)
	if err != nil {
		return err
	}
	defer conn.Close()

	return nil
}

func probeTCP(target string) bool {
	for i := 0; i < maxDialRetry; i++ {
		err := dialTCP(target)
		if err != nil {
			log.Errorf("Failed to dial %s, %v", target, err)
			time.Sleep(retryDialInterval)
			continue
		}
		log.Info("Successfully dialed")
		return true
	}
	return false
}
