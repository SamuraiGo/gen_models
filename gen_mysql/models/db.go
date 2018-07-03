package models

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

var (
	MainDB *sql.DB
	err    error
)

func InitMainDB(conn_string string) {
	MainDB, err = sql.Open("mysql", conn_string)
	if err != nil {
		log.Fatalf("【init_maindb.NewEngine】ex:%s\n", err.Error())
		return
	}
	err = MainDB.Ping()
	if err != nil {
		log.Fatalf("【init_maindb.Ping】ex:%s\n", err.Error())
		return
	}
	MainDB.SetMaxIdleConns(2)
	MainDB.SetMaxOpenConns(50)
}
