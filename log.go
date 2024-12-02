package main

import (
	"log"
	"os"
)

// log.go
// This file contains all logging functionality.

// errorLogger logs errors to the error.log file.
// It is used to log errors that occur during the execution of the program.
var errorLogger = func() *log.Logger {
	file, err := os.OpenFile("errors.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	return log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}()

// infoLogger is a logger that writes to a file called info.log.
// It is used to log non-error information, such as successful operations.
var infoLogger = func() *log.Logger {
	file, err := os.OpenFile("info.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	return log.New(file, "INFO: ", log.Ldate|log.Ltime)
}()
