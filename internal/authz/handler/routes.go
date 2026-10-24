package handler

import (
	"github.com/gin-gonic/gin"
)

// init installs the authorization route on the /v1 group.
func init() {
	Register(func(engine *gin.RouterGroup, h *Handler, mws ...gin.HandlerFunc) {
		authz := engine.Group("/authz", mws...)
		authz.POST("", h.Authorize)
	})
}
