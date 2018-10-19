// Copyright 2018 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/ngaut/log"
	"github.com/pkg/errors"
)

var (
	dbname           = "mysql"
	probeSQL         = "SELECT count(*) FROM tidb"
	tidbMaxOpenConns = 5
	tidbMaxIdleConns = 5
	tidbDialTimeout  = "30s"
	tidbReadTimeout  = "30s"
	tidbWriteTimeout = "30s"
)

func accessDatabase(username, password, address, dbname string) (*sql.DB, error) {
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s)/%s?timeout=%s&readTimeout=%s&writeTimeout=%s", username, password, address, dbname, tidbDialTimeout, tidbReadTimeout, tidbWriteTimeout)
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return nil, errors.Annotatef(err, "create database handle '%s'", dataSourceName)
	}

	db.SetMaxOpenConns(tidbMaxOpenConns)
	db.SetMaxIdleConns(tidbMaxIdleConns)

	err = db.Ping()
	if err != nil {
		return nil, errors.Annotatef(err, "ping database '%s'", dataSourceName)
	}

	log.Infof("ping database '%s'", address)
	return db, nil
}

func probeQuery(addr string, db *sql.DB) (label string, err error) {
	var count int
	rows, err := db.Query(probeSQL)
	if err != nil {
		log.Errorf("<%s> database query error, %v", addr, err)
		return "query", err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			log.Errorf("<%s> scan result sets of query '%s' error, %v", addr, probeSQL, err)
			return "scan", err
		}
		log.Infof("<%s> database query: %s, result sets: %d", addr, probeSQL, count)
	}

	err = rows.Err()
	if err != nil {
		log.Errorf("<%s> retrieve result sets of query '%s' error, %v", addr, probeSQL, err)
		return "retrieve", err
	}

	return "", nil
}
