package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"os"
	"strings"
)
import _ "github.com/go-sql-driver/mysql"

var dbConfig = mysql.Config{
	User:                 os.Getenv("dbUser"),
	Passwd:               os.Getenv("dbPass"),
	Net:                  "tcp",
	Addr:                 os.Getenv("dbAddr"), // localhost:6001
	DBName:               os.Getenv("dbName"),
	AllowNativePasswords: true,
}

func logSuccess(db *sql.DB, mode string) {
	sqlQuery := "INSERT INTO logs (type, log_dt, error) VALUES (?,?,?)"
	statement, _ := db.Prepare(sqlQuery)
	_, _ = statement.Exec(mode, CurrentDateTime(), nil)
}

type applicationArgs struct {
	debug      bool
	logActions bool
	jobType    JobType
	city       string
	interval   string
	lang       string
}

func printHelp() {
	fmt.Println("How to use:")
	fmt.Println("retailerTool jobType additionalArgs")
	fmt.Println("Available job types: sell / rent")
	fmt.Println("Available args:")
	fmt.Println("--city=riga")
	fmt.Println("--interval=all (all/today/today-2/today-5)")
	fmt.Println("--lang=lv")
	fmt.Println("--logOff")
	fmt.Println("--debug")
	fmt.Println("Example call:")
	fmt.Println("retailerTool rent --city=riga --interval=today")
}

func createApplicationArgs(args []string) (*applicationArgs, error) {
	if len(args) == 0 {
		return nil, errors.New("Try run with --help")
	}

	firstArg := args[0]
	var job JobType
	switch firstArg {
	case "sell":
		job = SellJob
	case "rent":
		job = RentJob
	default:
		printHelp()
		os.Exit(0)
	}

	appArgs := &applicationArgs{
		debug:      false,
		logActions: true,
		jobType:    job,
		city:       "riga",
		interval:   "all",
		lang:       "ru",
	}

	for _, v := range args[1:] {
		command := strings.Split(v, "=")
		if len(command) == 2 {
			switch command[0] {
			case "--city":
				appArgs.city = command[1]
			case "--interval":
				appArgs.interval = command[1]
			case "--lang":
				appArgs.lang = command[1]
			}
		}

		if len(command) == 1 {
			switch command[0] {
			case "--logOff":
				appArgs.logActions = false
			case "--debug":
				appArgs.debug = true
			}
		}

	}

	return appArgs, nil
}

func initDb() *sql.DB {
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
	return db
}

func runWithAppArgs(args *applicationArgs) {
	var db *sql.DB
	if !args.debug {
		db = initDb()
	}

	var logger Logger
	if args.logActions {
		logger = ConsoleLogger{}
	} else {
		logger = StubLogger{}
	}

	crawler := Crawler{
		logger: logger,
	}
	command := Command{
		UserAgent: Firefox,
		JobType:   args.jobType,
		Lang:      JobLang(args.lang),
		City:      City(args.city),
		Interval:  Interval(args.interval),
	}

	flatStorage := crawler.RunCommand(command)

	if !args.debug {
		flatStorage.Save(db)
		logSuccess(db, string(args.jobType))
		db.Close()
	} else {
		fmt.Println(flatStorage.GetAll())
	}
}

func main() {
	args := os.Args[1:]
	appArgs, err := createApplicationArgs(args)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	runWithAppArgs(appArgs)
}
