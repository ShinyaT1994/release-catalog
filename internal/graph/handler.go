package graph

import (
	"net/http"
	"strconv"

	"github.com/ShinyaT1994/release-catalog/internal/shared/middleware"
	"github.com/labstack/echo/v4"
)

// Handler handles HTTP requests for Graph
type Handler struct {
	uc UseCase
}

func NewHandler(uc UseCase) *Handler {
	return &Handler{uc: uc}
}

// RegisterRoutes registers graph routes
func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.GET("/branches/:branchId/current/graph", h.GetBranchCurrentGraph)
	g.GET("/releases/:releaseId/graph", h.GetReleaseGraph)
}

func (h *Handler) GetBranchCurrentGraph(c echo.Context) error {
	branchID := c.Param("branchId")
	opts := parseOptions(c)
	g, err := h.uc.GetBranchCurrentGraph(c.Request().Context(), branchID, opts)
	if err != nil {
		return middleware.SendError(c, err)
	}
	return c.JSON(http.StatusOK, g)
}

func (h *Handler) GetReleaseGraph(c echo.Context) error {
	releaseID := c.Param("releaseId")
	opts := parseOptions(c)
	g, err := h.uc.GetReleaseGraph(c.Request().Context(), releaseID, opts)
	if err != nil {
		return middleware.SendError(c, err)
	}
	return c.JSON(http.StatusOK, g)
}

func parseOptions(c echo.Context) Options {
	opts := DefaultOptions()
	if v := c.QueryParam("maxDepth"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			opts.MaxDepth = n
		}
	}
	if v := c.QueryParam("maxNodes"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			opts.MaxNodes = n
		}
	}
	return opts
}
