package log

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

// Microservice Logger
var Logger *log.Logger

func Init(serviceName string) {
	Logger = log.New(
		os.Stdout,
		fmt.Sprintf("[%s]", serviceName),
		log.LstdFlags|log.Ldate|log.Ltime|log.Lshortfile)
}

func Info(msg string) {
	Logger.Printf("[Info] %s", msg)
}

func Error(msg string) {
	Logger.Printf("[Error] %s", msg)
}

func Fatal(msg string) {
	Logger.Fatalf("[Fatal] %s", msg)
}

// Request logger
func LogRequest(r *http.Request) {
	Logger.Printf("[Request] %s %s %s", r.RemoteAddr, r.Method, r.URL)
}

// Response logger
func LogResponse(r *http.Request, status int) {
	Logger.Printf("[Response] %s %s %s %d", r.RemoteAddr, r.Method, r.URL, status)
}
