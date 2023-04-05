package logger

import (
	"net/http"
	"time"

	"github.com/ernesto-jimenez/httplogger"
)

// HTTPLogger wraps a L and satisfys the interface required by
// https://godoc.org/github.com/ernesto-jimenez/httplogger#HTTPLogger
type HTTPLogger struct {
	fn func(msg any, keyvals ...any) error
}

var _ httplogger.HTTPLogger = (*HTTPLogger)(nil)

// NewHTTPLogger creates a new HTTPLogger. level is the log level at which you want
// your HTTP requests logged at.
func NewHTTPLogger(l *L, level string) *HTTPLogger {
	hl := &HTTPLogger{}
	switch ParseLevel(level) {
	case Err:
		hl.fn = l.Err
	case Warn:
		hl.fn = l.Warn
	case Info:
		hl.fn = l.Info
	default:
		hl.fn = l.Debug
	}
	return hl
}

// LogRequest doesn't do anything since we'll be logging replies only
func (h *HTTPLogger) LogRequest(*http.Request) {}

// LogResponse logs path, host, status code and duration in milliseconds
func (h *HTTPLogger) LogResponse(req *http.Request, res *http.Response, err error, duration time.Duration) {
	duration /= time.Millisecond
	if err != nil {
		h.fn("HTTP Request Error",
			"method", req.Method,
			"host", req.Host,
			"path", req.URL.Path,
			"status", "error",
			"durationMs", duration,
			"error", err,
		)
	} else {
		h.fn("HTTP Request",
			"method", req.Method,
			"host", req.Host,
			"path", req.URL.Path,
			"status", res.StatusCode,
			"durationMs", duration,
		)
	}
}
