package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	//TestQuery check tidb normal
	TestQuery = "SELECT count(*) FROM mysql.tidb"
)

type tidbHealth struct {
	Address string
	Health  bool
	Message string
}
type tidbHealths []tidbHealth

func tidbTest(hostAndPort string) error {
	var dsn string
	if *password == "" {
		dsn = fmt.Sprintf("%s@tcp(%s)/mysql?charset=utf8&timeout=%ds&writeTimeout=%ds&readTimeout=%ds",
			*user, hostAndPort, *querytimeout, *querytimeout, *querytimeout)
	} else {
		dsn = fmt.Sprintf("%s:%s@tcp(%s)/mysql?charset=utf8&timeout=%ds&writeTimeout=%ds&readTimeout=%ds",
			*user, *password, hostAndPort, *querytimeout, *querytimeout, *querytimeout)
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	rs, err := db.Query(TestQuery)
	if err != nil {
		return err
	}
	defer rs.Close()
	var tc int64
	for rs.Next() {
		err := rs.Scan(&tc)
		if err != nil {
			return err
		}
		break
	}
	return nil
}

func checkTiDB(tidbs []string) (tidbHealths, error) {
	var tidbhs tidbHealths
	for _, tidb := range tidbs {
		h := tidbHealth{
			Address: tidb,
			Health:  true,
		}
		if err := tidbTest(tidb); err != nil {
			h.Health = false
			h.Message = err.Error()
		}
		tidbhs = append(tidbhs, h)
	}
	return tidbhs, nil
}

func goroutineTiDB(tidbs []string, checkInterval time.Duration) {
	for {
		if tidbhs, err := checkTiDB(tidbs); err != nil {
			log.Debugf("check tidb failed with error %v", err)
			exporter.WithLabelValues(promTiDBType, checkedAllInstance, checkedFailed).Inc()
		} else {
			for _, h := range tidbhs {
				checked := checkedSuccess
				if !h.Health {
					checked = checkedFailed
					xAlert("", promTiDBType, h.Address, checked)
				}
				exporter.WithLabelValues(promTiDBType, h.Address, checked).Inc()
			}
		}
		time.Sleep(checkInterval)
	}
}
