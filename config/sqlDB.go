package config

import (
	"database/sql"
	"fmt"
	"log"
	_ "github.com/go-sql-driver/mysql"
)

func EstablishConnection() {
	db, err := sql.Open("mysql", fmt.Sprintf("%v:%v@(localhost)/%v", DBUser, DBPass, DBName))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Pre DB ping")
	err = db.Ping()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Finished DB ping")
	defer db.Close()
}