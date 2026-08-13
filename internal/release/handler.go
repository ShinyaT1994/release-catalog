package release

import (
	"net/http"
	"strconv"

	"github.com/ShinyaT1994/release-catalog/internal/shared/middleware"
	"github.com/labstack/echo/v4"
)

// Handler handles HTTP requests for Release/Snapshot
type Handler struct {
	uc UseCase
}

// NewHandler creates a new release handler
func NewHandler(uc UseCase) *Handler {
	return &Handler{uc: uc}
}

// RegisterRoutes registers release routes
func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.POST("/branches/:branchId/snapshots", h.CreateSnapshot)
	g.POST("/branches/:branchId/releases", h.CreateRelease)
	g.GET("/branches/:branchId/releases", h.ListByBranchID)
	g.GET("/releases/:releaseId", h.GetByID)
}

func (h *Handler) CreateSnapshot(c echo.Context) error {
	branchID := c.Param("branchId")
	var input CreateSnapshotInput
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	snap, err := h.uc.CreateMainSnapshot(c.Request().Context(), branchID, input)
	if err != nil {
		return middleware.SendError(c, err)
	}
	return c.JSON(http.StatusCreated, snap)
}

func (h *Handler) CreateRelease(c echo.Context) error {
	branchID := c.Param("branchId")
	var input CreateReleaseInput
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	snap, err := h.uc.CreateRelease(c.Request().Context(), branchID, input)
	if err != nil {
		return middleware.SendError(c, err)
	}
	return c.JSON(http.StatusCreated, snap)
}

func (h *Handler) GetByID(c echo.Context) error {
	id := c.Param("releaseId")
	snap, err := h.uc.GetByID(c.Request().Context(), id)
	if err != nil {
		return middleware.SendError(c, err)
	}
	return c.JSON(http.StatusOK, snap)
}

func (h *Handler) ListByBranchID(c echo.Context) error {
	branchID := c.Param("branchId")
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
	releases, err := h.uc.ListByBranchID(c.Request().Context(), branchID, opts)
	if err != nil {
		return middleware.SendError(c, err)
	}
	if releases == nil {
		return c.JSON(http.StatusOK, []interface{}{})
	}
	return c.JSON(http.StatusOK, releases)
}
