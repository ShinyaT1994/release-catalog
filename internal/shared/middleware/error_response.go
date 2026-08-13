package middleware

import (
	"net/http"

	"github.com/ShinyaT1994/release-catalog/internal/shared/apperror"
	"github.com/labstack/echo/v4"
)

// APIError is the standard error response
type APIError struct {
	Error     string      `json:"error"`
	Message   string      `json:"message"`
	Details   interface{} `json:"details,omitempty"`
	RequestID string      `json:"requestId"`
}

// HTTPStatusForCode maps error codes to HTTP status codes
func HTTPStatusForCode(code apperror.Code) int {
	switch code {
	case apperror.CodeInvalidRequest:
		return http.StatusBadRequest
	case apperror.CodeProductNotFound, apperror.CodeBranchNotFound,
		apperror.CodeReleaseNotFound, apperror.CodeRootProjectNotFound:
		return http.StatusNotFound
	case apperror.CodeRootBOMChanged, apperror.CodeConflict:
		return http.StatusConflict
	case apperror.CodeBOMLinkInvalid, apperror.CodeBOMLinkUnresolved:
		return http.StatusUnprocessableEntity
	case apperror.CodeDTUnavailable:
		return http.StatusBadGateway
	case apperror.CodeGraphLimitExceeded:
		return http.StatusRequestEntityTooLarge
	default:
		return http.StatusInternalServerError
	}
}

// SendError sends a standardized error response
func SendError(c echo.Context, err error) error {
	reqID, _ := c.Get(RequestIDKey).(string)

	if appErr, ok := err.(*apperror.Error); ok {
		return c.JSON(HTTPStatusForCode(appErr.Code), APIError{
			Error:     string(appErr.Code),
			Message:   appErr.Message,
			Details:   appErr.Details,
			RequestID: reqID,
		})
	}

	return c.JSON(http.StatusInternalServerError, APIError{
		Error:     string(apperror.CodeInternalError),
		Message:   "internal server error",
		RequestID: reqID,
	})
}
