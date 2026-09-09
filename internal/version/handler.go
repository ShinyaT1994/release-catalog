package version

import (
	"net/http"
	"strconv"

	"github.com/ShinyaT1994/release-catalog/internal/shared/middleware"
	"github.com/labstack/echo/v4"
)

// Handler handles HTTP requests for versions.
type Handler struct {
	uc UseCase
}

func NewHandler(uc UseCase) *Handler {
	return &Handler{uc: uc}
}

// RegisterRoutes registers version routes.
func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.POST("/branches/:branchId/versions", h.Create)
	g.GET("/branches/:branchId/versions", h.ListByBranchID)
	g.GET("/versions/:versionId", h.GetByID)
	g.PATCH("/versions/:versionId", h.Update)
	g.DELETE("/versions/:versionId", h.Delete)

	g.PUT("/versions/:versionId/projects/root", h.SetRoot)
	g.POST("/versions/:versionId/projects", h.AddProject)
	g.DELETE("/versions/:versionId/projects/:bindingId", h.RemoveProject)
}

func (h *Handler) Create(c echo.Context) error {
	branchID := c.Param("branchId")
	var input CreateInput
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	v, err := h.uc.Create(c.Request().Context(), branchID, input)
	if err != nil {
		return middleware.SendError(c, err)
	}
	return c.JSON(http.StatusCreated, v)
}

func (h *Handler) GetByID(c echo.Context) error {
	v, err := h.uc.GetByID(c.Request().Context(), c.Param("versionId"))
	if err != nil {
		return middleware.SendError(c, err)
	}
	return c.JSON(http.StatusOK, v)
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
	versions, err := h.uc.ListByBranchID(c.Request().Context(), branchID, opts)
	if err != nil {
		return middleware.SendError(c, err)
	}
	if versions == nil {
		return c.JSON(http.StatusOK, []interface{}{})
	}
	return c.JSON(http.StatusOK, versions)
}

func (h *Handler) Update(c echo.Context) error {
	var input UpdateInput
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	v, err := h.uc.Update(c.Request().Context(), c.Param("versionId"), input)
	if err != nil {
		return middleware.SendError(c, err)
	}
	return c.JSON(http.StatusOK, v)
}

func (h *Handler) Delete(c echo.Context) error {
	if err := h.uc.Delete(c.Request().Context(), c.Param("versionId")); err != nil {
		return middleware.SendError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) SetRoot(c echo.Context) error {
	var input SetProjectInput
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	p, err := h.uc.SetRoot(c.Request().Context(), c.Param("versionId"), input)
	if err != nil {
		return middleware.SendError(c, err)
	}
	return c.JSON(http.StatusOK, p)
}

func (h *Handler) AddProject(c echo.Context) error {
	var input AddProjectInput
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	p, err := h.uc.AddProject(c.Request().Context(), c.Param("versionId"), input)
	if err != nil {
		return middleware.SendError(c, err)
	}
	return c.JSON(http.StatusCreated, p)
}

func (h *Handler) RemoveProject(c echo.Context) error {
	bindingID, err := strconv.ParseInt(c.Param("bindingId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid binding id"})
	}
	if err := h.uc.RemoveProject(c.Request().Context(), c.Param("versionId"), bindingID); err != nil {
		return middleware.SendError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}
