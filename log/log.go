package log

import "fmt"

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
