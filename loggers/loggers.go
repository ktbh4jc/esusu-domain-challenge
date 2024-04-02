package loggers

import (
	"io/ioutil"
	"log"
	"os"
)

/*
  A simple collection of loggers to be used by the rest of the project
	Copied from https://rollbar.com/blog/golang-error-logging-guide/
*/

var (
	WarningLog *log.Logger
	InfoLog    *log.Logger
	ErrorLog   *log.Logger
)

func Init() {
	InfoLog = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLog = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLog = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// Silent Initializer for unit tests
func SilentInit() {
	InfoLog = log.New(ioutil.Discard, "", 0)
	WarningLog = log.New(ioutil.Discard, "", 0)
	// ErrorLog = log.New(ioutil.Discard, "", 0)
	ErrorLog = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}
