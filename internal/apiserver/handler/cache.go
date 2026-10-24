package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/onexstack/onexstack/pkg/core"

	"github.com/onexdemo/iamv2/internal/pkg/contextx"
	"github.com/onexdemo/iamv2/internal/pkg/errno"
	v1 "github.com/onexdemo/iamv2/pkg/api/apiserver/v1"
)

// ListSecretInfo returns a snapshot of every access key. It is restricted to
// administrators because the payload contains secret keys.
func (h *Handler) ListSecretInfo(c *gin.Context) {
	if !contextx.IsAdmin(c.Request.Context()) {
		core.WriteResponse(c, nil, errno.ErrPermissionDenied)

		return
	}

	var rq v1.ListSecretInfoRequest

	if err := c.ShouldBindQuery(&rq); err != nil {
		core.WriteResponse(c, nil, errno.ErrBind.WithMessage("%s", err.Error()))

		return
	}

	resp, err := h.biz.CacheV1().ListSecrets(c.Request.Context(), &rq)
	core.WriteResponse(c, resp, err)
}

// ListPolicyInfo returns a snapshot of every authorization policy. It is
// restricted to administrators because policies are security-sensitive.
func (h *Handler) ListPolicyInfo(c *gin.Context) {
	if !contextx.IsAdmin(c.Request.Context()) {
		core.WriteResponse(c, nil, errno.ErrPermissionDenied)

		return
	}

	var rq v1.ListPolicyInfoRequest

	if err := c.ShouldBindQuery(&rq); err != nil {
		core.WriteResponse(c, nil, errno.ErrBind.WithMessage("%s", err.Error()))

		return
	}

	resp, err := h.biz.CacheV1().ListPolicies(c.Request.Context(), &rq)
	core.WriteResponse(c, resp, err)
}
