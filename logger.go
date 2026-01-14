package session

import (
	"fmt"
	"log"
)

type Logger interface {
	Infof(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
}

type defaultLogger struct{}

func (*defaultLogger) Infof(msg string, args ...interface{}) {
	log.Printf("[INFO] %s", fmt.Sprintf(msg, args...))
}

func (*defaultLogger) Debugf(msg string, args ...interface{}) {
	log.Printf("[DEBUG] %s", fmt.Sprintf(msg, args...))
}

func (*defaultLogger) Errorf(msg string, args ...interface{}) {
	log.Printf("[ERROR] %s", fmt.Sprintf(msg, args...))
}

func (*defaultLogger) Warnf(msg string, args ...interface{}) {
	log.Printf("[WARN] %s", fmt.Sprintf(msg, args...))
}
