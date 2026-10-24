package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/onexstack/onexstack/pkg/core"
	"github.com/ory/ladon"

	"github.com/onexdemo/iamv2/internal/pkg/contextx"
	"github.com/onexdemo/iamv2/internal/pkg/errno"
)

// Authorize evaluates whether the authenticated user may perform the requested
// action on the requested resource.
func (h *Handler) Authorize(c *gin.Context) {
	var rq ladon.Request

	if err := c.ShouldBindJSON(&rq); err != nil {
		core.WriteResponse(c, nil, errno.ErrBind.WithMessage("%s", err.Error()))

		return
	}

	if rq.Context == nil {
		rq.Context = ladon.Context{}
	}

	rq.Context["username"] = contextx.Username(c.Request.Context())

	resp, err := h.biz.Authorize(c.Request.Context(), &rq)
	core.WriteResponse(c, resp, err)
}
