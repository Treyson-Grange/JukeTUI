package main

import (
	"log"
	"os"
)

// log.go
// This file contains all logging functionality.

var errorLogger = func() *log.Logger {
	file, err := os.OpenFile("errors.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	return log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}()

var infoLogger = func() *log.Logger {
	file, err := os.OpenFile("info.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	return log.New(file, "INFO: ", log.Ldate|log.Ltime)
}()
