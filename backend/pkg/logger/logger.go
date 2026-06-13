package logger

import (
	"log"
	"os"
)

func New(appEnv string) *log.Logger {
	prefix := "[ticketstream][" + appEnv + "] "
	return log.New(os.Stdout, prefix, log.LstdFlags|log.Lshortfile)
}
