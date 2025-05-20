package handler

import (
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/mccune1224/strider/pages/pages"
)

type Handler struct{}

func (h *Handler) Render(ctx echo.Context, statusCode int, comp templ.Component) error {
	buf := templ.GetBuffer()
	if err := comp.Render(ctx.Request().Context(), buf); err != nil {
		return err
	}
	return ctx.HTML(statusCode, buf.String())
}

func (h *Handler) AttachRoutes(e *echo.Echo) {
	e.GET("/", h.Home)
}

func (h *Handler) Home(ctx echo.Context) error {
	// Render the home page
	return h.Render(ctx, 200, pages.PageHome())
}
