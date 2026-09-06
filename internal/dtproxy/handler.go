package dtproxy

import (
	"net/http"

	"github.com/ShinyaT1994/release-catalog/internal/dtclient"
	"github.com/labstack/echo/v4"
)

// Handler proxies DT project search requests
type Handler struct {
	dt dtclient.Client
}

func NewHandler(dt dtclient.Client) *Handler {
	return &Handler{dt: dt}
}

// RegisterRoutes registers DT proxy routes
func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.GET("/dt/projects", h.SearchProjects)
}

func (h *Handler) SearchProjects(c echo.Context) error {
	name := c.QueryParam("name")

	var projects []*dtclient.Project
	var err error

	if name == "" {
		projects, err = h.dt.ListProjects(c.Request().Context())
	} else {
		projects, err = h.dt.SearchProjects(c.Request().Context(), name)
	}

	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{
			"error":   "DEPENDENCY_TRACK_UNAVAILABLE",
			"message": err.Error(),
		})
	}

	if projects == nil {
		projects = []*dtclient.Project{}
	}

	return c.JSON(http.StatusOK, projects)
}
