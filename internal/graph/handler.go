package graph

import (
	"net/http"
	"strconv"

	"github.com/ShinyaT1994/release-catalog/internal/shared/middleware"
	"github.com/labstack/echo/v4"
)

// Handler handles HTTP requests for graphs and timelines.
type Handler struct {
	sbom    SBOMUseCase
	lineage LineageUseCase
}

func NewHandler(sbom SBOMUseCase, lineage LineageUseCase) *Handler {
	return &Handler{sbom: sbom, lineage: lineage}
}

// RegisterRoutes registers graph routes.
func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.GET("/versions/:versionId/projects/:role/graph", h.GetProjectGraph)
	g.GET("/products/:productId/lineage-graph", h.GetProductLineage)
	g.GET("/products/:productId/timeline", h.GetProductTimeline)
	g.GET("/branches/:branchId/timeline", h.GetBranchTimeline)
}

func (h *Handler) GetProjectGraph(c echo.Context) error {
	versionID := c.Param("versionId")
	role := c.Param("role")
	opts := parseOptions(c)
	g, err := h.sbom.GetProjectGraph(c.Request().Context(), versionID, role, opts)
	if err != nil {
		return middleware.SendError(c, err)
	}
	return c.JSON(http.StatusOK, g)
}

func (h *Handler) GetProductLineage(c echo.Context) error {
	productID := c.Param("productId")
	axis := LineageAxis(c.QueryParam("axis"))
	g, err := h.lineage.GetProductLineage(c.Request().Context(), productID, axis)
	if err != nil {
		return middleware.SendError(c, err)
	}
	return c.JSON(http.StatusOK, g)
}

func (h *Handler) GetProductTimeline(c echo.Context) error {
	productID := c.Param("productId")
	t, err := h.lineage.GetProductTimeline(c.Request().Context(), productID)
	if err != nil {
		return middleware.SendError(c, err)
	}
	return c.JSON(http.StatusOK, t)
}

func (h *Handler) GetBranchTimeline(c echo.Context) error {
	branchID := c.Param("branchId")
	t, err := h.lineage.GetBranchTimeline(c.Request().Context(), branchID)
	if err != nil {
		return middleware.SendError(c, err)
	}
	return c.JSON(http.StatusOK, t)
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
