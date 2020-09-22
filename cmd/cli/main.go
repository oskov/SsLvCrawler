package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/retailerTool/config"
	"github.com/retailerTool/crawlerDb"
	"github.com/retailerTool/crawlerPackage"
	"github.com/retailerTool/util"
	"os"
	"strings"
)

type applicationArgs struct {
	debug      bool
	logActions bool
	jobType    crawlerPackage.JobType
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
	var job crawlerPackage.JobType
	switch firstArg {
	case "sell":
		job = crawlerPackage.SellJob
	case "rent":
		job = crawlerPackage.RentJob
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

func initDb() *sqlx.DB {
	db, err := sql.Open("mysql", config.DbConfig.FormatDSN())
	if err != nil {
		fmt.Println("Unable to open mysql connection")
		os.Exit(-1)
	}
	err = crawlerDb.RunMigrations(db)
	if err != nil {
		fmt.Println("Unable to make migrations")
		fmt.Println(err)
		os.Exit(-1)
	}
	sqlxDb := sqlx.NewDb(db, "mysql")
	return sqlxDb
}

func runWithAppArgs(args *applicationArgs) {
	var db *sqlx.DB
	if !args.debug {
		db = initDb()
	}

	var logger util.Logger
	if args.logActions {
		logger = util.ConsoleLogger{}
	} else {
		logger = util.StubLogger{}
	}

	crawler := crawlerPackage.Crawler{
		Logger: logger,
	}
	command := crawlerPackage.Command{
		UserAgent: crawlerPackage.Firefox,
		JobType:   args.jobType,
		Lang:      crawlerPackage.JobLang(args.lang),
		City:      crawlerPackage.City(args.city),
		Interval:  crawlerPackage.Interval(args.interval),
	}

	flatStorage := crawler.RunCommand(command)

	if !args.debug {
		flatStorage.Save(db)
		util.LogSuccess(db, args.jobType.DbType)
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
