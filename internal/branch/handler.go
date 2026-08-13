package branch

import (
	"net/http"
	"strconv"

	"github.com/ShinyaT1994/release-catalog/internal/shared/middleware"
	"github.com/labstack/echo/v4"
)

// Handler handles HTTP requests for Branch
type Handler struct {
	uc UseCase
}

// NewHandler creates a new branch handler
func NewHandler(uc UseCase) *Handler {
	return &Handler{uc: uc}
}

// RegisterRoutes registers branch routes
func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.GET("/products/:productId/branches", h.ListByProductID)
	g.GET("/branches/:branchId", h.GetByID)
	g.POST("/products/:productId/release-lines", h.CreateReleaseLine)
	g.PATCH("/branches/:branchId", h.Update)
	g.GET("/branches/:branchId/current", h.GetCurrentState)
	g.PUT("/branches/:branchId/current", h.UpdateCurrentState)
}

func (h *Handler) CreateReleaseLine(c echo.Context) error {
	productID := c.Param("productId")
	var input CreateReleaseLineInput
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	b, err := h.uc.CreateReleaseLine(c.Request().Context(), productID, input)
	if err != nil {
		return middleware.SendError(c, err)
	}
	return c.JSON(http.StatusCreated, b)
}

func (h *Handler) GetByID(c echo.Context) error {
	id := c.Param("branchId")
	b, err := h.uc.GetByID(c.Request().Context(), id)
	if err != nil {
		return middleware.SendError(c, err)
	}
	return c.JSON(http.StatusOK, b)
}

func (h *Handler) ListByProductID(c echo.Context) error {
	productID := c.Param("productId")
	opts := DefaultListOptions()
	if v := c.QueryParam("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			opts.Limit = n
		}
	}
	if v := c.QueryParam("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			opts.Offset = n
		}
	}
	branches, err := h.uc.ListByProductID(c.Request().Context(), productID, opts)
	if err != nil {
		return middleware.SendError(c, err)
	}
	if branches == nil {
		return c.JSON(http.StatusOK, []interface{}{})
	}
	return c.JSON(http.StatusOK, branches)
}

func (h *Handler) Update(c echo.Context) error {
	id := c.Param("branchId")
	var input UpdateBranchInput
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	b, err := h.uc.Update(c.Request().Context(), id, input)
	if err != nil {
		return middleware.SendError(c, err)
	}
	return c.JSON(http.StatusOK, b)
}

func (h *Handler) GetCurrentState(c echo.Context) error {
	branchID := c.Param("branchId")
	cs, err := h.uc.GetCurrentState(c.Request().Context(), branchID)
	if err != nil {
		return middleware.SendError(c, err)
	}
	return c.JSON(http.StatusOK, cs)
}

func (h *Handler) UpdateCurrentState(c echo.Context) error {
	branchID := c.Param("branchId")
	var input UpdateCurrentStateInput
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	cs, err := h.uc.UpdateCurrentState(c.Request().Context(), branchID, input)
	if err != nil {
		return middleware.SendError(c, err)
	}
	return c.JSON(http.StatusOK, cs)
}
