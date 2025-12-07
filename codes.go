package errenvelope

// Code is a stable, machine-readable error identifier.
type Code string

const (
	// Generic
	CodeInternal         Code = "INTERNAL"
	CodeBadRequest       Code = "BAD_REQUEST"
	CodeNotFound         Code = "NOT_FOUND"
	CodeMethodNotAllowed Code = "METHOD_NOT_ALLOWED"
	CodeGone             Code = "GONE"
	CodeConflict         Code = "CONFLICT"
	CodePayloadTooLarge  Code = "PAYLOAD_TOO_LARGE"
	CodeRequestTimeout   Code = "REQUEST_TIMEOUT"
	CodeRateLimited      Code = "RATE_LIMITED"
	CodeUnavailable      Code = "UNAVAILABLE"

	// Validation / auth
	CodeValidationFailed     Code = "VALIDATION_FAILED"
	CodeUnauthorized         Code = "UNAUTHORIZED"
	CodeForbidden            Code = "FORBIDDEN"
	CodeUnprocessableEntity  Code = "UNPROCESSABLE_ENTITY"

	// Timeouts / cancellations
	CodeTimeout  Code = "TIMEOUT"
	CodeCanceled Code = "CANCELED"

	// Downstream
	CodeDownstream        Code = "DOWNSTREAM_ERROR"
	CodeDownstreamTimeout Code = "DOWNSTREAM_TIMEOUT"
)
