package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/src-d/code-annotation/server/serializer"

	"github.com/pressly/lg"
)

// Levels
const (
	Fatal int = iota
	ManualPanic
	Panic
	Error
	Warning
	Info
	Debug
	Unk
)

// Log testing handler
func Log(level int) RequestProcessFunc {
	return func(r *http.Request) (*serializer.Response, error) {
		logger := lg.RequestLog(r)
		switch level {
		case Fatal:
			logger.Fatal("Fatal logged. App will exit")
			logger.Error("This message should not appear after a fatal message")
		case ManualPanic:
			panic("Panic manually thrown")
		case Panic:
			logger.Panic("Panic logged")
			logger.Error("This message should not appear after a panic message")
		case Error:
			return nil, fmt.Errorf("Error sent")
		case Warning:
			logger.Warn("Warn logged")
		case Info:
			logger.Info("Info logged")
		case Debug:
			logger.Debug("Debug logged")
		default:
			logger.Error("unknown requested level")
		}

		logger.Info("Message sent")

		return serializer.NewVersionResponse(strconv.Itoa(level)), nil
	}
}

// LoggingTestMiddleware testing middleware
func LoggingTestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := lg.RequestLog(r)
		testValue := r.FormValue("test-middleware")

		switch testValue {
		case "fatal":
			logger.Fatal("LoggingTestMiddleware killed the app")
			logger.Error("This message should not appear after a fatal message")
		case "manual-panic":
			panic("LoggingTestMiddleware throwed a manual panic")
		case "panic":
			logger.Panic("LoggingTestMiddleware throwed a panic")
			logger.Error("This message should not appear after a panic message")
		case "error":
			logger.Error("LoggingTestMiddleware logged an error message")
			w.WriteHeader(http.StatusInternalServerError)
			return
		case "warn":
			logger.Info("LoggingTestMiddleware logged an warning message")
		case "info":
			logger.Info("LoggingTestMiddleware logged an info message")
		case "debug":
			logger.Info("LoggingTestMiddleware logged an debug message")
		}

		next.ServeHTTP(w, r)
	})
}
