package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/onexstack/onexstack/pkg/core"

	"github.com/onexdemo/iamv2/internal/pkg/errno"
	v1 "github.com/onexdemo/iamv2/pkg/api/apiserver/v1"
)

// CreateSecret creates a new access key pair.
func (h *Handler) CreateSecret(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.SecretV1().Create, h.val.ValidateCreateSecretRequest)
}

// GetSecret returns one of the caller's secrets.
func (h *Handler) GetSecret(c *gin.Context) {
	rq := &v1.GetSecretRequest{Name: c.Param("name")}
	if err := h.val.ValidateGetSecretRequest(c.Request.Context(), rq); err != nil {
		core.WriteResponse(c, nil, err)

		return
	}

	resp, err := h.biz.SecretV1().Get(c.Request.Context(), rq)
	core.WriteResponse(c, resp, err)
}

// UpdateSecret updates the expiry and description of a secret.
func (h *Handler) UpdateSecret(c *gin.Context) {
	var rq v1.UpdateSecretRequest

	if err := c.ShouldBindJSON(&rq); err != nil {
		core.WriteResponse(c, nil, errno.ErrBind.WithMessage("%s", err.Error()))

		return
	}

	rq.Name = c.Param("name")

	if err := h.val.ValidateUpdateSecretRequest(c.Request.Context(), &rq); err != nil {
		core.WriteResponse(c, nil, err)

		return
	}

	resp, err := h.biz.SecretV1().Update(c.Request.Context(), &rq)
	core.WriteResponse(c, resp, err)
}

// DeleteSecret removes one of the caller's secrets.
func (h *Handler) DeleteSecret(c *gin.Context) {
	rq := &v1.DeleteSecretRequest{Name: c.Param("name")}
	if err := h.val.ValidateDeleteSecretRequest(c.Request.Context(), rq); err != nil {
		core.WriteResponse(c, nil, err)

		return
	}

	resp, err := h.biz.SecretV1().Delete(c.Request.Context(), rq)
	core.WriteResponse(c, resp, err)
}

// DeleteSecrets removes several of the caller's secrets.
func (h *Handler) DeleteSecrets(c *gin.Context) {
	rq := &v1.DeleteSecretsRequest{Names: c.QueryArray("name")}
	if err := h.val.ValidateDeleteSecretsRequest(c.Request.Context(), rq); err != nil {
		core.WriteResponse(c, nil, err)

		return
	}

	resp, err := h.biz.SecretV1().DeleteCollection(c.Request.Context(), rq)
	core.WriteResponse(c, resp, err)
}

// ListSecrets returns the caller's secrets.
func (h *Handler) ListSecrets(c *gin.Context) {
	core.HandleQueryRequest(c, h.biz.SecretV1().List, h.val.ValidateListSecretsRequest)
}
