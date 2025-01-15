package logger

import "log"

type GooseLogger struct{}

func (GooseLogger) Fatalf(format string, v ...interface{}) { log.Fatalf(format, v...) }
func (GooseLogger) Printf(format string, v ...interface{}) { log.Printf(format, v...) }
