package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/ngaut/log"
)

var (
	dbname           = "mysql"
	probeSQL         = "SELECT count(*) FROM tidb"
	tidbMaxOpenConns = 5
	tidbMaxIdleConns = 5
)

func accessDatabase(username, password, address, dbname string) (*sql.DB, error) {
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s)/%s", username, password, address, dbname)
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Errorf("create database handle '%s' error, %v", dataSourceName, err)
		return nil, err
	}

	db.SetMaxOpenConns(tidbMaxOpenConns)
	db.SetMaxIdleConns(tidbMaxIdleConns)

	err = db.Ping()
	if err != nil {
		log.Errorf("ping database '%s' error, %v", dataSourceName, err)
		return nil, err
	}

	return db, nil
}

func probeQuery(db *sql.DB) error {
	var count int
	rows, err := db.Query(probeSQL)
	if err != nil {
		log.Errorf("database query error, %v", err)
		return err
	}

	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			log.Errorf("scan result sets of query '%s' error, %v", probeSQL, err)
		}
		log.Infof("database query: %s, result sets: %d", probeSQL, count)
	}

	err = rows.Err()
	if err != nil {
		log.Errorf("retrieve result sets of query '%s' error, %v", probeSQL, err)
		return err
	}
	return nil
}
