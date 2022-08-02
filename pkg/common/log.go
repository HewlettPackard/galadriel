package common

import (
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	debugLevel = "DEBUG"
	infoLevel  = "INFO"
	warnLevel  = "WARN"
	errLevel   = "ERROR"
)

type Logger struct {
	name   string
	logger *log.Logger
}

func NewLogger(name string) *Logger {
	return &Logger{
		name:   name,
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

func (l *Logger) Debug(args ...interface{}) {
	l.logWithLevel(debugLevel, args)
}

func (l *Logger) Info(args ...interface{}) {
	l.logWithLevel(infoLevel, args)
}

func (l *Logger) Warn(args ...interface{}) {
	l.logWithLevel(warnLevel, args)
}

func (l *Logger) Error(args ...interface{}) {
	l.logWithLevel(errLevel, args)
}

func (l *Logger) logWithLevel(level string, args []interface{}) {
	var params []string
	for _, v := range args {
		params = append(params, fmt.Sprint(v))
	}
	l.logger.Println(fmt.Sprintf("[%s] %s:", l.name, level), strings.Join(params, " "))
}
