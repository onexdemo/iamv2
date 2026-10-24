package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/onexdemo/iamv2/internal/authz/biz"
)

// Handler dispatches HTTP requests to the business layer.
type Handler struct {
	biz biz.IBiz
}

// Registrar registers a group of HTTP routes.
type Registrar func(engine *gin.RouterGroup, h *Handler, mws ...gin.HandlerFunc)

var registrars []Registrar

// NewHandler creates a Handler.
func NewHandler(biz biz.IBiz) *Handler {
	return &Handler{biz: biz}
}

// Register adds a route registrar to the global registry.
func Register(r Registrar) {
	registrars = append(registrars, r)
}

// ApplyTo installs every registered route group on engine.
func (h *Handler) ApplyTo(engine *gin.RouterGroup, mws ...gin.HandlerFunc) {
	for _, r := range registrars {
		r(engine, h, mws...)
	}

	slog.Info("rest api routes installed", "count", len(registrars))
}
