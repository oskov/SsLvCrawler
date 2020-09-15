package config

import (
	"github.com/go-sql-driver/mysql"
	"os"
)

var DbConfig = mysql.Config{
	User:                 os.Getenv("dbUser"),
	Passwd:               os.Getenv("dbPass"),
	Net:                  "tcp",
	Addr:                 os.Getenv("dbAddr"), // localhost:6001
	DBName:               os.Getenv("dbName"),
	AllowNativePasswords: true,
	ParseTime:            true,
}
