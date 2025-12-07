package errenvelope

// Code is a stable, machine-readable error identifier.
type Code string

const (
	// Generic
	CodeInternal    Code = "INTERNAL"
	CodeBadRequest  Code = "BAD_REQUEST"
	CodeNotFound    Code = "NOT_FOUND"
	CodeConflict    Code = "CONFLICT"
	CodeRateLimited Code = "RATE_LIMITED"
	CodeUnavailable Code = "UNAVAILABLE"

	// Validation / auth
	CodeValidationFailed Code = "VALIDATION_FAILED"
	CodeUnauthorized     Code = "UNAUTHORIZED"
	CodeForbidden        Code = "FORBIDDEN"

	// Timeouts / cancellations
	CodeTimeout  Code = "TIMEOUT"
	CodeCanceled Code = "CANCELED"

	// Downstream
	CodeDownstream        Code = "DOWNSTREAM_ERROR"
	CodeDownstreamTimeout Code = "DOWNSTREAM_TIMEOUT"
)
