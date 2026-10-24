package handler

import (
	"encoding/base64"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/onexstack/onexstack/pkg/core"

	"github.com/onexdemo/iamv2/internal/pkg/errno"
	v1 "github.com/onexdemo/iamv2/pkg/api/apiserver/v1"
)

// CreateUser registers a new user account.
func (h *Handler) CreateUser(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.UserV1().Create, h.val.ValidateCreateUserRequest)
}

// GetUser returns a single user.
func (h *Handler) GetUser(c *gin.Context) {
	rq := &v1.GetUserRequest{Username: c.Param("username")}
	if err := h.val.ValidateGetUserRequest(c.Request.Context(), rq); err != nil {
		core.WriteResponse(c, nil, err)

		return
	}

	resp, err := h.biz.UserV1().Get(c.Request.Context(), rq)
	core.WriteResponse(c, resp, err)
}

// UpdateUser updates a user's profile.
func (h *Handler) UpdateUser(c *gin.Context) {
	var rq v1.UpdateUserRequest

	if err := c.ShouldBindJSON(&rq); err != nil {
		core.WriteResponse(c, nil, errno.ErrBind.WithMessage("%s", err.Error()))

		return
	}

	rq.Name = c.Param("username")

	if err := h.val.ValidateUpdateUserRequest(c.Request.Context(), &rq); err != nil {
		core.WriteResponse(c, nil, err)

		return
	}

	resp, err := h.biz.UserV1().Update(c.Request.Context(), &rq)
	core.WriteResponse(c, resp, err)
}

// DeleteUser removes a single user.
func (h *Handler) DeleteUser(c *gin.Context) {
	rq := &v1.DeleteUserRequest{Username: c.Param("username")}
	if err := h.val.ValidateDeleteUserRequest(c.Request.Context(), rq); err != nil {
		core.WriteResponse(c, nil, err)

		return
	}

	resp, err := h.biz.UserV1().Delete(c.Request.Context(), rq)
	core.WriteResponse(c, resp, err)
}

// DeleteUsers removes several users in one request.
func (h *Handler) DeleteUsers(c *gin.Context) {
	rq := &v1.DeleteUsersRequest{Usernames: c.QueryArray("name")}
	if err := h.val.ValidateDeleteUsersRequest(c.Request.Context(), rq); err != nil {
		core.WriteResponse(c, nil, err)

		return
	}

	resp, err := h.biz.UserV1().DeleteCollection(c.Request.Context(), rq)
	core.WriteResponse(c, resp, err)
}

// ListUsers returns a paginated user list.
func (h *Handler) ListUsers(c *gin.Context) {
	core.HandleQueryRequest(c, h.biz.UserV1().List, h.val.ValidateListUsersRequest)
}

// ChangePassword changes the authenticated user's password.
func (h *Handler) ChangePassword(c *gin.Context) {
	var rq v1.ChangePasswordRequest

	if err := c.ShouldBindJSON(&rq); err != nil {
		core.WriteResponse(c, nil, errno.ErrBind.WithMessage("%s", err.Error()))

		return
	}

	if err := h.val.ValidateChangePasswordRequest(c.Request.Context(), &rq); err != nil {
		core.WriteResponse(c, nil, err)

		return
	}

	resp, err := h.biz.UserV1().ChangePassword(c.Request.Context(), &rq)
	core.WriteResponse(c, resp, err)
}

// Login authenticates a user and returns a signed JWT. It accepts either a
// JSON body or an HTTP Basic Authorization header, matching the original iam
// API.
func (h *Handler) Login(c *gin.Context) {
	rq, err := parseLoginRequest(c)
	if err != nil {
		core.WriteResponse(c, nil, err)

		return
	}

	if err := h.val.ValidateLoginRequest(c.Request.Context(), rq); err != nil {
		core.WriteResponse(c, nil, err)

		return
	}

	resp, err := h.biz.UserV1().Login(c.Request.Context(), rq)
	if err != nil {
		core.WriteResponse(c, nil, err)

		return
	}

	setTokenCookie(c, resp)
	core.WriteResponse(c, resp, nil)
}

// Logout clears the legacy JWT cookie. Tokens are stateless, so there is
// nothing else to invalidate.
func (h *Handler) Logout(c *gin.Context) {
	c.SetCookie("jwt", "", -1, "/", "", false, true)
	core.WriteResponse(c, nil, nil)
}

// RefreshToken issues a fresh token for the authenticated caller.
func (h *Handler) RefreshToken(c *gin.Context) {
	resp, err := h.biz.UserV1().RefreshToken(c.Request.Context())
	if err != nil {
		core.WriteResponse(c, nil, err)

		return
	}

	setTokenCookie(c, resp)
	core.WriteResponse(c, resp, nil)
}

func parseLoginRequest(c *gin.Context) (*v1.LoginRequest, error) {
	header := c.Request.Header.Get("Authorization")
	if username, password, ok := parseBasicHeader(header); ok {
		return &v1.LoginRequest{Username: username, Password: password}, nil
	}

	var rq v1.LoginRequest

	if err := c.ShouldBindJSON(&rq); err != nil {
		return nil, errno.ErrBind.WithMessage("%s", err.Error())
	}

	return &rq, nil
}

func parseBasicHeader(header string) (string, string, bool) {
	scheme, credential, ok := strings.Cut(header, " ")
	if !ok || !strings.EqualFold(scheme, "Basic") {
		return "", "", false
	}

	payload, err := base64.StdEncoding.DecodeString(credential)
	if err != nil {
		return "", "", false
	}

	username, password, ok := strings.Cut(string(payload), ":")
	if !ok {
		return "", "", false
	}

	return username, password, true
}

func setTokenCookie(c *gin.Context, resp *v1.LoginResponse) {
	maxAge := int(time.Until(resp.ExpireAt).Seconds())
	if maxAge <= 0 {
		maxAge = int((24 * time.Hour).Seconds())
	}

	c.SetCookie("jwt", resp.Token, maxAge, "/", "", false, true)
}
