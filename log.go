package main

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

type StubLogger struct{}

func (StubLogger) Log(string string) {
}

func (StubLogger) LogData(data interface{}) {
}
