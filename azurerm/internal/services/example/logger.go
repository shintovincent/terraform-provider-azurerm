package example

import (
	"fmt"
	"log"
)

type ExampleLogger struct {
}

func (ExampleLogger) Info(message string) {
	log.Print(message)
}

func (ExampleLogger) InfoF(format string, args ...interface{}) {
	log.Print(fmt.Sprintf(format, args...))
}

func (ExampleLogger) Warn(message string) {
	log.Print(message)
}

func (ExampleLogger) WarnF(format string, args ...interface{}) {
	log.Print(fmt.Sprintf(format, args...))
}
