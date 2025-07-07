package tracer

import "context"

type contextKey int

const (
	RequestIDKey contextKey = iota
	CorrelationIDKey
)

var (
	// RequestIDHeader is the name of the HTTP Header which contains the request id.
	// Exported so that it can be changed by developers
	RequestIDHeader     = "X-Request-ID"
	CorrelationIDHeader = "X-Correlation-ID"
)

// GetRequestID returns a request ID from the given context if one is present.
// Returns the empty string if a request ID cannot be found.
func GetRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value(RequestIDKey).(string); ok {
		return reqID
	}
	return ""
}

// GetCorrelationID returns a correlation ID from the given context if one is present.
// Returns the empty string if a correlation ID cannot be found.
func GetCorrelationID(ctx context.Context) string {
	if corrID, ok := ctx.Value(CorrelationIDKey).(string); ok {
		return corrID
	}
	return ""
}
