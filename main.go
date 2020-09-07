package main

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
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
	_, _ = statement.Exec(mode, CurrentDateTime(), nil)
}

func main() {
	crawler := Crawler{
		logger: ConsoleLogger{},
	}
	db, err := sql.Open("mysql", dbConfig.FormatDSN())
	if err != nil {
		fmt.Println("Unable to open mysql connection")
		os.Exit(-1)
	}
	err = RunMigrations(db)
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

	job := SellJob

	switch mode {
	case "sell":
		job = SellJob
	case "rent":
		job = RentJob
	}

	command := Command{
		UserAgent: Firefox,
		JobType:   job,
		Lang:      Ru,
		City:      City("riga"),
		Interval:  All,
	}

	crawler.logger.Log("Start crawling")
	flatStorage := crawler.RunCommand(command)

	crawler.logger.Log("Save to db")
	flatStorage.Save(db)

	logSuccess(db, mode)
	db.Close()
}
