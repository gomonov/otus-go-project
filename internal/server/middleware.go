package server

import (
	"fmt"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(logger Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrappedWriter := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrappedWriter, r)

		latency := time.Since(start)
		userAgent := r.UserAgent()
		if userAgent == "" {
			userAgent = "-"
		}

		timestamp := start.Format("02/Jan/2006:15:04:05 -0700")

		logEntry := fmt.Sprintf("%s [%s] %s %s %s %d %d \"%s\"",
			getClientIP(r),
			timestamp,
			r.Method,
			r.URL.RequestURI(),
			r.Proto,
			wrappedWriter.statusCode,
			latency.Milliseconds(),
			userAgent,
		)

		logger.Info(logEntry)
	})
}

func getClientIP(r *http.Request) string {
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		return forwardedFor
	}
	return r.RemoteAddr
}
