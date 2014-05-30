package plog

import (
	"log"
	"os"
)

const (
	DEBUG = iota
	INFO  = iota
	WARN  = iota
	ERROR = iota
	FATAL = iota
)

var LEVEL = INFO

func Debug(msg string, v ...interface{}) {
	if LEVEL <= DEBUG {
		log.Printf(msg, v...)
	}
}

func Info(msg string, v ...interface{}) {
	if LEVEL <= INFO {
		log.Printf(msg, v...)
	}
}

func Warn(msg string, v ...interface{}) {
	if LEVEL <= WARN {
		log.Printf(msg, v...)
	}
}

func Error(msg string, v ...interface{}) {
	if LEVEL <= ERROR {
		log.Printf(msg, v...)
	}
}

func Exc(err error) {
	if LEVEL <= ERROR {
		log.Print(err)
	}
}

func Fatal(msg string, v ...interface{}) {
	if LEVEL <= FATAL {
		log.Printf(msg, v...)
	}
	os.Exit(1)
}

func SetLevel(level int) {
	LEVEL = level
}
