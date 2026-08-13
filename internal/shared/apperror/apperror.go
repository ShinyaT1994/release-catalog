package apperror

import "fmt"

// Code represents application error codes
type Code string

const (
	CodeInvalidRequest      Code = "INVALID_REQUEST"
	CodeProductNotFound     Code = "PRODUCT_NOT_FOUND"
	CodeBranchNotFound      Code = "BRANCH_NOT_FOUND"
	CodeReleaseNotFound     Code = "RELEASE_NOT_FOUND"
	CodeRootProjectNotFound Code = "ROOT_PROJECT_NOT_FOUND"
	CodeRootBOMChanged      Code = "ROOT_BOM_CHANGED"
	CodeBOMLinkInvalid      Code = "BOM_LINK_INVALID"
	CodeBOMLinkUnresolved   Code = "BOM_LINK_UNRESOLVED"
	CodeDTUnavailable       Code = "DEPENDENCY_TRACK_UNAVAILABLE"
	CodeGraphLimitExceeded  Code = "GRAPH_LIMIT_EXCEEDED"
	CodeInternalError       Code = "INTERNAL_ERROR"
	CodeConflict            Code = "CONFLICT"
)

// Error is the application-level error
type Error struct {
	Code    Code
	Message string
	Details interface{}
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func New(code Code, message string) *Error {
	return &Error{Code: code, Message: message}
}

func WithDetails(code Code, message string, details interface{}) *Error {
	return &Error{Code: code, Message: message, Details: details}
}
