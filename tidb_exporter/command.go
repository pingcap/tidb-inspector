package main

import (
	"fmt"
	gh "github.com/dustin/go-humanize"
	"strings"
)

func commandMode() {
	/*
	   PD:
	       Name: pd1 ID: 10574917096089111283 ClientUrl:http://127.0.0.1:2379    Health: true
	       Name: pd1 ID: 10574917096089111283 ClientUrl:http://127.0.0.1:2379    Health: true
	       Name: pd1 ID: 10574917096089111283 ClientUrl:http://127.0.0.1:2379    Health: true
	*/
	//pd information
	if *pds != "" {
		fmt.Printf("PD:\n")
		if pdhs, err := checkPD(strings.Split(*pds, ",")); err == nil {
			for _, pdh := range pdhs {
				fmt.Printf("\tName: %s\tID: %d\tClientURLs: %s\tHealth: %t\n",
					pdh.Name, pdh.MemberID, strings.Join(pdh.ClientUrls, ","), pdh.Health)
			}
		} else {
			fmt.Printf("Error: %v\n", err)
		}
	}

	/*
	   TiKV:
	       ID: 1 Address: 10.0.3.5:20160 State: Up Capacity: 55 GiB Available: 41 GiB LastHeartbeatTS: 2018-02-06T14:58:17.499983761+08:00
	       ID: 1 Address: 10.0.3.5:20160 State: Up Capacity: 55 GiB Available: 41 GiB LastHeartbeatTS: 2018-02-06T14:58:17.499983761+08:00
	       ID: 1 Address: 10.0.3.5:20160 State: Up Capacity: 55 GiB Available: 41 GiB LastHeartbeatTS: 2018-02-06T14:58:17.499983761+08:00

	*/
	//tikv information
	if *pds != "" {
		fmt.Printf("TiKV:\n")
		if tikvInfos, err := checkTiKV(strings.Split(*pds, ",")); err == nil {
			for _, tikvInfo := range tikvInfos.Stores {
				store := tikvInfo.Store
				stat := tikvInfo.Status
				fmt.Printf("\tID: %d\tAddress: %s\tState: %s\tCapacity: %s\tAvailable: %s\tLastHeartbeatTS: %s\n",
					store.Id, store.Address, store.StateName, gh.IBytes(uint64(stat.Capacity)), gh.IBytes(uint64(stat.Available)), stat.LastHeartbeatTS)
			}
		} else {
			fmt.Printf("Error: %v\n", err)
		}
	}
	/*
	   TiDB:
	       Address: 10.0.3.5:4000 Health: true Message:
	       Address: 10.0.3.5:4000 Health: false Message: can not connect db
	       Address: 10.0.3.5:4000 Health: true
	*/
	//tidb information
	if *tidbs != "" {
		fmt.Printf("TiDB:\n")
		if tidbhs, err := checkTiDB(strings.Split(*tidbs, ",")); err == nil {
			for _, tidbh := range tidbhs {
				fmt.Printf("\tAddress: %s\tHealth: %t\tMessage: %s\n", tidbh.Address, tidbh.Health, tidbh.Message)
			}
		} else {
			fmt.Printf("Error: %v\n", err)
		}
	}
}
