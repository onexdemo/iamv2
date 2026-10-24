package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/onexstack/onexstack/pkg/core"

	"github.com/onexdemo/iamv2/internal/pkg/errno"
	v1 "github.com/onexdemo/iamv2/pkg/api/apiserver/v1"
)

// CreatePolicy stores a new ladon policy.
func (h *Handler) CreatePolicy(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.PolicyV1().Create, h.val.ValidateCreatePolicyRequest)
}

// GetPolicy returns one of the caller's policies.
func (h *Handler) GetPolicy(c *gin.Context) {
	rq := &v1.GetPolicyRequest{Name: c.Param("name")}
	if err := h.val.ValidateGetPolicyRequest(c.Request.Context(), rq); err != nil {
		core.WriteResponse(c, nil, err)

		return
	}

	resp, err := h.biz.PolicyV1().Get(c.Request.Context(), rq)
	core.WriteResponse(c, resp, err)
}

// UpdatePolicy replaces the ladon rules of a policy.
func (h *Handler) UpdatePolicy(c *gin.Context) {
	var rq v1.UpdatePolicyRequest

	if err := c.ShouldBindJSON(&rq); err != nil {
		core.WriteResponse(c, nil, errno.ErrBind.WithMessage("%s", err.Error()))

		return
	}

	rq.Name = c.Param("name")

	if err := h.val.ValidateUpdatePolicyRequest(c.Request.Context(), &rq); err != nil {
		core.WriteResponse(c, nil, err)

		return
	}

	resp, err := h.biz.PolicyV1().Update(c.Request.Context(), &rq)
	core.WriteResponse(c, resp, err)
}

// DeletePolicy removes one of the caller's policies.
func (h *Handler) DeletePolicy(c *gin.Context) {
	rq := &v1.DeletePolicyRequest{Name: c.Param("name")}
	if err := h.val.ValidateDeletePolicyRequest(c.Request.Context(), rq); err != nil {
		core.WriteResponse(c, nil, err)

		return
	}

	resp, err := h.biz.PolicyV1().Delete(c.Request.Context(), rq)
	core.WriteResponse(c, resp, err)
}

// DeletePolicies removes several of the caller's policies.
func (h *Handler) DeletePolicies(c *gin.Context) {
	rq := &v1.DeletePoliciesRequest{Names: c.QueryArray("name")}
	if err := h.val.ValidateDeletePoliciesRequest(c.Request.Context(), rq); err != nil {
		core.WriteResponse(c, nil, err)

		return
	}

	resp, err := h.biz.PolicyV1().DeleteCollection(c.Request.Context(), rq)
	core.WriteResponse(c, resp, err)
}

// ListPolicies returns the caller's policies.
func (h *Handler) ListPolicies(c *gin.Context) {
	core.HandleQueryRequest(c, h.biz.PolicyV1().List, h.val.ValidateListPoliciesRequest)
}
