package v1

import "time"

// User is the public representation of an iam user.
type User struct {
	ObjectMeta `json:",inline"`

	// Nickname is the display name of the user.
	Nickname string `json:"nickname"`
	// Email is the user's email address.
	Email string `json:"email"`
	// Phone is the user's phone number.
	Phone string `json:"phone"`
	// Status is 1 for enabled and 0 for disabled accounts.
	Status int `json:"status"`
	// IsAdmin reports whether the user has administrative privileges.
	IsAdmin int `json:"isAdmin"`
	// TotalPolicy is the number of authorization policies owned by the user.
	TotalPolicy int64 `json:"totalPolicy"`
	// LoginedAt is the last successful login time.
	LoginedAt time.Time `json:"loginedAt,omitempty"`
}

// CreateUserRequest is the payload accepted by POST /v1/users.
type CreateUserRequest struct {
	ObjectMeta `json:",inline"`

	// Password is the plaintext password. It is encrypted before persisting.
	Password string `json:"password" binding:"required"`
	// Nickname is the display name of the user.
	Nickname string `json:"nickname"`
	// Email is the user's email address.
	Email string `json:"email"`
	// Phone is the user's phone number.
	Phone string `json:"phone"`
}

// UpdateUserRequest is the payload accepted by PUT /v1/users/:username.
type UpdateUserRequest struct {
	ObjectMeta `json:",inline"`

	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

// ChangePasswordRequest is the payload accepted by
// PUT /v1/users/:username/change-password.
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

// GetUserRequest identifies a single user.
type GetUserRequest struct {
	Username string `json:"username" binding:"required" uri:"username"`
}

// DeleteUserRequest identifies a single user to delete.
type DeleteUserRequest struct {
	Username string `json:"username" binding:"required" uri:"username"`
}

// DeleteUsersRequest deletes several users in one request.
type DeleteUsersRequest struct {
	Usernames []string `json:"usernames" binding:"required" form:"name"`
}

// ListUsersRequest is the query payload accepted by GET /v1/users.
type ListUsersRequest struct {
	ListOptions
}

// ListUsersResponse is returned by GET /v1/users.
type ListUsersResponse struct {
	ListMeta `json:",inline"`

	Items []*User `json:"items"`
}

// LoginRequest authenticates a user.
type LoginRequest struct {
	Username string `json:"username" binding:"required" form:"username"`
	Password string `json:"password" binding:"required" form:"password"`
}

// CreateUserResponse is returned by POST /v1/users.
type CreateUserResponse struct {
	User *User `json:"user"`
}

// UpdateUserResponse is returned by PUT /v1/users/:username.
type UpdateUserResponse struct {
	User *User `json:"user"`
}

// GetUserResponse is returned by GET /v1/users/:username.
type GetUserResponse struct {
	User *User `json:"user"`
}

// DeleteUserResponse is returned by DELETE /v1/users/:username.
type DeleteUserResponse struct{}

// DeleteUsersResponse is returned by DELETE /v1/users.
type DeleteUsersResponse struct{}

// ChangePasswordResponse is returned by PUT /v1/users/:username/change-password.
type ChangePasswordResponse struct{}

// LoginResponse is returned after a successful login.
type LoginResponse struct {
	Token    string    `json:"token"`
	ExpireAt time.Time `json:"expire"`
	User     *User     `json:"user,omitempty"`
}

// RefreshTokenRequest is accepted by POST /v1/refresh.
type RefreshTokenRequest struct {
	// Token is optional. When empty the token is read from the Authorization header.
	Token string `json:"token"`
}
