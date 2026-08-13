package product

import (
	"net/http"
	"strconv"

	"github.com/ShinyaT1994/release-catalog/internal/shared/middleware"
	"github.com/labstack/echo/v4"
)

// Handler handles HTTP requests for Product
type Handler struct {
	uc UseCase
}

// NewHandler creates a new product handler
func NewHandler(uc UseCase) *Handler {
	return &Handler{uc: uc}
}

// RegisterRoutes registers product routes
func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.POST("/products", h.Create)
	g.GET("/products", h.List)
	g.GET("/products/:productId", h.GetByID)
	g.PATCH("/products/:productId", h.Update)
	g.DELETE("/products/:productId", h.Delete)
}

func (h *Handler) Create(c echo.Context) error {
	var input CreateInput
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	p, err := h.uc.Create(c.Request().Context(), input)
	if err != nil {
		return middleware.SendError(c, err)
	}
	return c.JSON(http.StatusCreated, p)
}

func (h *Handler) GetByID(c echo.Context) error {
	id := c.Param("productId")
	p, err := h.uc.GetByID(c.Request().Context(), id)
	if err != nil {
		return middleware.SendError(c, err)
	}
	return c.JSON(http.StatusOK, p)
}

func (h *Handler) List(c echo.Context) error {
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
	products, err := h.uc.List(c.Request().Context(), opts)
	if err != nil {
		return middleware.SendError(c, err)
	}
	if products == nil {
		return c.JSON(http.StatusOK, []interface{}{})
	}
	return c.JSON(http.StatusOK, products)
}

func (h *Handler) Update(c echo.Context) error {
	id := c.Param("productId")
	var input UpdateInput
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	p, err := h.uc.Update(c.Request().Context(), id, input)
	if err != nil {
		return middleware.SendError(c, err)
	}
	return c.JSON(http.StatusOK, p)
}

func (h *Handler) Delete(c echo.Context) error {
	id := c.Param("productId")
	if err := h.uc.Delete(c.Request().Context(), id); err != nil {
		return middleware.SendError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}
