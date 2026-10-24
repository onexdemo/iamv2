package errno

import (
	"net/http"

	"github.com/onexstack/onexstack/pkg/errorsx"
)

var (
	// ErrUserNotFound indicates that the requested user does not exist.
	ErrUserNotFound = errorsx.New(http.StatusNotFound, "NotFound.UserNotFound", "User not found.")

	// ErrUserAlreadyExists indicates that the username is already taken.
	ErrUserAlreadyExists = errorsx.New(http.StatusBadRequest, "AlreadyExist.UserAlreadyExists", "User already exists.")

	// ErrPasswordIncorrect indicates that the supplied password is wrong.
	ErrPasswordIncorrect = errorsx.New(http.StatusUnauthorized, "Unauthenticated.PasswordIncorrect", "Password is incorrect.")

	// ErrUsernameInvalid indicates that the username does not satisfy the
	// platform naming rules.
	ErrUsernameInvalid = errorsx.New(
		http.StatusBadRequest,
		"InvalidArgument.UsernameInvalid",
		"Username must consist of letters, digits, and underscores only, and its length must be between 3 and 32 characters.",
	)

	// ErrPasswordInvalid indicates that the password is too weak.
	ErrPasswordInvalid = errorsx.New(
		http.StatusBadRequest,
		"InvalidArgument.PasswordInvalid",
		"Password must contain at least 6 characters.",
	)

	// ErrUserForbidden indicates that the user is not allowed to perform the
	// requested operation.
	ErrUserForbidden = errorsx.New(http.StatusForbidden, "PermissionDenied.UserForbidden", "You are not allowed to perform this operation.")
)
