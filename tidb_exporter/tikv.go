package main

import (
	"fmt"
	"github.com/juju/errors"
	"github.com/pingcap/kvproto/pkg/metapb"
	"github.com/pingcap/pd/pkg/typeutil"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

const (
	//pdStoresAPI get tikv stores status
	pdStoresAPI       = "/pd/api/v1/stores"
	tikvSuccessStatus = "Up"
)

// all struct is lower case, can not import
//https://github.com/pingcap/pd/blob/master/server/api/store.go
type metaStore struct {
	*metapb.Store
	StateName string `json:"state_name"`
}
type storeStatus struct {
	Capacity           typeutil.ByteSize  `json:"capacity,omitempty"`
	Available          typeutil.ByteSize  `json:"available,omitempty"`
	LeaderCount        int                `json:"leader_count,omitempty"`
	LeaderWeight       float64            `json:"leader_weight,omitempty"`
	LeaderScore        float64            `json:"leader_score,omitempty"`
	LeaderSize         int64              `json:"leader_size,omitempty"`
	RegionCount        int                `json:"region_count,omitempty"`
	RegionWeight       float64            `json:"region_weight,omitempty"`
	RegionScore        float64            `json:"region_score,omitempty"`
	RegionSize         int64              `json:"region_size,omitempty"`
	SendingSnapCount   uint32             `json:"sending_snap_count,omitempty"`
	ReceivingSnapCount uint32             `json:"receiving_snap_count,omitempty"`
	ApplyingSnapCount  uint32             `json:"applying_snap_count,omitempty"`
	IsBusy             bool               `json:"is_busy,omitempty"`
	StartTS            *time.Time         `json:"start_ts,omitempty"`
	LastHeartbeatTS    *time.Time         `json:"last_heartbeat_ts,omitempty"`
	Uptime             *typeutil.Duration `json:"uptime,omitempty"`
}
type storeInfo struct {
	Store  *metaStore   `json:"store"`
	Status *storeStatus `json:"status"`
}
type storesInfo struct {
	Count  int          `json:"count"`
	Stores []*storeInfo `json:"stores"`
}

func checkTiKV(pdURLs []string) (*storesInfo, error) {
	var err error
	stores := &storesInfo{}
	for _, pdURL := range pdURLs {
		if err = xGet(fmt.Sprintf("%s%s", pdURL, pdStoresAPI), stores, true); err == nil {
			return stores, nil
		}
	}
	return stores, errors.Trace(err)
}

func goroutineTiKV(pdURLs []string, checkInterval time.Duration) {
	for {
		if tikvhs, err := checkTiKV(pdURLs); err != nil {
			log.Debugf("check TiKV with error %v", err)
			exporter.WithLabelValues(promTiKVType, checkedAllInstance, checkedFailed).Inc()
		} else {
			for _, tikvh := range tikvhs.Stores {
				checked := checkedSuccess
				if strings.ToLower(tikvh.Store.StateName) != strings.ToLower(tikvSuccessStatus) {
					checked = checkedFailed
				}
				if tikvh.Status.Uptime.Nanoseconds() < checkInterval.Nanoseconds() {
					checked = checkedRestart
				}

				if checked != checkedSuccess {
					xAlert("", promTiKVType, tikvh.Store.Address, checked)
				}
				exporter.WithLabelValues(promTiKVType, tikvh.Store.Address, checked).Inc()
			}
		}
		time.Sleep(checkInterval)
	}
}
