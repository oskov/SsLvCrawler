package config

import (
	"github.com/go-sql-driver/mysql"
	"os"
)

var DbConfig = mysql.Config{
	User:                 os.Getenv("MYSQL_USER"),
	Passwd:               os.Getenv("MYSQL_PASSWORD"),
	Net:                  "tcp",
	Addr:                 os.Getenv("MYSQL_ADDR"), // localhost:6001
	DBName:               os.Getenv("MYSQL_DATABASE"),
	AllowNativePasswords: true,
	ParseTime:            true,
	CheckConnLiveness:    true,
}
