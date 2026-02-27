package logging

import (
	"log"
	"os"
)

var Logger *log.Logger

func Init() {
	Logger = log.New(os.Stdout, "", log.LstdFlags|log.LUTC)
}
