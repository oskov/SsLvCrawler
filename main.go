package main

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/retailerTool/log"
	"github.com/retailerTool/storage"
	"github.com/retailerTool/utils"
	"os"
)
import _ "github.com/go-sql-driver/mysql"

const dbUser = "golang"
const dbPass = "password"
const dbName = "retail_db"

var dbConfig = mysql.Config{
	User:                 dbUser,
	Passwd:               dbPass,
	Net:                  "tcp",
	Addr:                 "localhost:6001",
	DBName:               dbName,
	AllowNativePasswords: true,
}

func logSuccess(db *sql.DB, mode string) {
	sqlQuery := "INSERT INTO logs (type, log_dt, error) VALUES (?,?,?)"
	statement, _ := db.Prepare(sqlQuery)
	_, _ = statement.Exec(mode, utils.CurrentDateTime(), nil)
}

func main() {
	crawler := Crawler{
		logger:      log.ConsoleLogger{},
		flatStorage: storage.NewFlatStorage(),
		userAgent:   UserAgents[0],
	}
	db, err := sql.Open("mysql", dbConfig.FormatDSN())
	if err != nil {
		fmt.Println("Unable to open mysql connection")
		os.Exit(-1)
	}
	err = utils.RunMigrations(db)
	if err != nil {
		fmt.Println("Unable to make migrations")
		fmt.Println(err)
		os.Exit(-1)
	}

	args := os.Args[1:]
	var mode string
	if len(args) == 0 {
		fmt.Println("Send sell or rent to args")
		os.Exit(-1)
	} else {
		mode = args[0]
	}

	var job Job

	switch mode {
	case "sell":
		job = Sell
	case "rent":
		job = Rent
	}

	crawler.Run(job)
	//crawler.flatStorage.Save(db)
	//logSuccess(db, mode)
	db.Close()
}
