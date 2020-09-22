package util

import (
	"fmt"
	"github.com/jmoiron/sqlx"
)

type Logger interface {
	Log(string string)
	LogData(data interface{})
}

type ConsoleLogger struct{}

func (ConsoleLogger) Log(string string) {
	fmt.Println(string)
}

func (ConsoleLogger) LogData(data interface{}) {
	fmt.Println(data)
}

type StubLogger struct{}

func (StubLogger) Log(string string) {
}

func (StubLogger) LogData(data interface{}) {
}

type DbLogger struct {
	Db *sqlx.DB
}

func LogSuccess(db *sqlx.DB, mode string) {
	sqlQuery := "INSERT INTO logs (type, log_dt, error) VALUES (?,?,?)"
	statement, _ := db.Prepare(sqlQuery)
	_, err := statement.Exec(mode, CurrentDateTime(), nil)
	if err != nil {
		fmt.Println(err)
	}
}
