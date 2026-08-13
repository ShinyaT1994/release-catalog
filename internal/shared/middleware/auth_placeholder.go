package middleware

import "github.com/labstack/echo/v4"

// AuthPlaceholder is a placeholder middleware for future authentication.
// Currently allows all requests through.
func AuthPlaceholder() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// TODO: Implement JWT / API Key authentication
			return next(c)
		}
	}
}
