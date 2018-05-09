package main

import (
	"bytes"
	"encoding/json"
	gonetstat "github.com/drael/GOnetstat"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	checkTCPTimeout = 5 * time.Second
	contentTypeJSON = "application/json"
)

func getHostName() string {
	instance, err := os.Hostname()
	if err != nil {
		instance = "Unknownhost"
	}
	return instance
}

func runCommand(command string) (int, string, error) {
	log.Debugf("run command %s", command)
	cmd := exec.Command("/bin/bash", "-c", command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("exec command %s, output %v, and error %v", command, string(out), err)
		return -1, string(out), errors.Trace(err)
	}
	waitStatus := cmd.ProcessState.Sys().(syscall.WaitStatus)
	return int(waitStatus), string(out), errors.Trace(err)
}

func isTCPPortAvailable(hostPort string) bool {
	conn, err := net.DialTimeout("tcp", hostPort, checkTCPTimeout)
	if err != nil {
		log.Errorf("conn host  port %s with error %v", hostPort, err)
		return false
	}
	defer conn.Close()
	return true
}

func getPidFromPort(port int64) int64 {
	var pid int64
	if port == 0 {
		return pid
	}

	netstat := gonetstat.Tcp()
	netstat = append(netstat, gonetstat.Tcp6()...)
	for _, p := range netstat {
		if p.Port == port && !strings.Contains(p.Pid, "-") {
			if tPid, err := strconv.ParseInt(p.Pid, 10, 64); err != nil {
				log.Errorf("can not tranform pid %s with error %v", p.Pid, errors.Trace(err))
			} else {
				pid = tPid
			}
			break
		}
	}
	return pid
}

func xGet(url string, data interface{}, getData bool) error {
	resp, err := http.Get(url)
	if err != nil {
		return errors.Trace(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("http get url %s returncode %d", url, resp.StatusCode)
	}

	if getData {
		return readJSON(resp.Body, data)
	}
	return nil
}

func readJSON(r io.ReadCloser, data interface{}) error {
	d, err := ioutil.ReadAll(r)
	if err != nil {
		return errors.Trace(err)
	}
	err = json.Unmarshal(d, data)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

func xPost(url string, data interface{}) error {
	b, errJ := json.Marshal(data)
	if errJ != nil {
		return errJ
	}
	var err error
	for i := 0; i < 3; i++ {
		resp, errR := http.Post(url, contentTypeJSON, bytes.NewBuffer(b))
		if err == nil || resp.StatusCode/100 == 2 {
			break
		}
		err = errR
	}
	return err
}
