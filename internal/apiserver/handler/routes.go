package handler

import (
	"github.com/gin-gonic/gin"
)

// init installs the iam REST routes on the /v1 group.
//
// User registration and login are public. Every other route requires either a
// Bearer token or HTTP Basic credentials; the middleware is passed in by the
// httpserver so that the handler package stays transport agnostic.
func init() {
	Register(func(engine *gin.RouterGroup, h *Handler, mws ...gin.HandlerFunc) {
		users := engine.Group("/users")
		// Public registration, matching the original iam API.
		users.POST("", h.CreateUser)
		users.Use(mws...)
		users.GET("", h.ListUsers)
		users.DELETE("", h.DeleteUsers)
		users.GET("/:username", h.GetUser)
		users.PUT("/:username", h.UpdateUser)
		users.DELETE("/:username", h.DeleteUser)
		users.PUT("/:username/change-password", h.ChangePassword)

		secrets := engine.Group("/secrets", mws...)
		secrets.POST("", h.CreateSecret)
		secrets.GET("", h.ListSecrets)
		secrets.DELETE("", h.DeleteSecrets)
		secrets.GET("/:name", h.GetSecret)
		secrets.PUT("/:name", h.UpdateSecret)
		secrets.DELETE("/:name", h.DeleteSecret)

		policies := engine.Group("/policies", mws...)
		policies.POST("", h.CreatePolicy)
		policies.GET("", h.ListPolicies)
		policies.DELETE("", h.DeletePolicies)
		policies.GET("/:name", h.GetPolicy)
		policies.PUT("/:name", h.UpdatePolicy)
		policies.DELETE("/:name", h.DeletePolicy)

		// Internal snapshot endpoints consumed by iam-authz-server. They are
		// additionally restricted to administrators in the handler.
		cache := engine.Group("/cache", mws...)
		cache.GET("/secrets", h.ListSecretInfo)
		cache.GET("/policies", h.ListPolicyInfo)
	})
}
