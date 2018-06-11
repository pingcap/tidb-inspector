package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/juju/errors"
	"github.com/ngaut/log"
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

	return db, nil
}

func probeQuery(db *sql.DB) (label string, err error) {
	var count int
	rows, err := db.Query(probeSQL)
	if err != nil {
		log.Errorf("database query error, %v", err)
		return "query", err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			log.Errorf("scan result sets of query '%s' error, %v", probeSQL, err)
			return "scan", err
		}
		log.Infof("database query: %s, result sets: %d", probeSQL, count)
	}

	err = rows.Err()
	if err != nil {
		log.Errorf("retrieve result sets of query '%s' error, %v", probeSQL, err)
		return "retrieve", err
	}
	return "", nil
}
