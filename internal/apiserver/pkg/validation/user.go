package validation

import (
	"context"
	"net/mail"
	"regexp"
	"strings"

	genericvalidation "github.com/onexstack/onexstack/pkg/validation"

	"github.com/onexdemo/iamv2/internal/pkg/errno"
	v1 "github.com/onexdemo/iamv2/pkg/api/apiserver/v1"
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)

// ValidateCreateUserRequest validates a user creation request.
func (v *Validator) ValidateCreateUserRequest(_ context.Context, rq *v1.CreateUserRequest) error {
	if err := validateUsername(rq.Name); err != nil {
		return err
	}

	if err := validatePassword(rq.Password); err != nil {
		return err
	}

	return validateUserProfile(rq.Nickname, rq.Email, rq.Phone)
}

// ValidateUpdateUserRequest validates a user update request.
func (v *Validator) ValidateUpdateUserRequest(_ context.Context, rq *v1.UpdateUserRequest) error {
	if err := validateUsername(rq.Name); err != nil {
		return err
	}

	return validateUserProfile(rq.Nickname, rq.Email, rq.Phone)
}

// ValidateChangePasswordRequest validates a password change request.
func (v *Validator) ValidateChangePasswordRequest(_ context.Context, rq *v1.ChangePasswordRequest) error {
	if err := validatePassword(rq.NewPassword); err != nil {
		return err
	}

	if rq.OldPassword == "" {
		return errno.ErrInvalidArgument.WithMessage("oldPassword cannot be empty")
	}

	return nil
}

// ValidateLoginRequest validates a login request.
func (v *Validator) ValidateLoginRequest(_ context.Context, rq *v1.LoginRequest) error {
	return genericvalidation.ValidateAllFields(rq, genericvalidation.Rules{
		"Username": func(value any) error {
			return validateUsername(value.(string))
		},
		"Password": func(value any) error {
			if value.(string) == "" {
				return errno.ErrInvalidArgument.WithMessage("password cannot be empty")
			}

			return nil
		},
	})
}

// ValidateListUsersRequest validates list query parameters.
func (v *Validator) ValidateListUsersRequest(_ context.Context, rq *v1.ListUsersRequest) error {
	return validateListOptions(rq.ListOptions)
}

// ValidateGetUserRequest validates a single-user lookup.
func (v *Validator) ValidateGetUserRequest(_ context.Context, rq *v1.GetUserRequest) error {
	return validateUsername(rq.Username)
}

// ValidateDeleteUserRequest validates a single-user deletion.
func (v *Validator) ValidateDeleteUserRequest(_ context.Context, rq *v1.DeleteUserRequest) error {
	return validateUsername(rq.Username)
}

// ValidateDeleteUsersRequest validates a batch deletion.
func (v *Validator) ValidateDeleteUsersRequest(_ context.Context, rq *v1.DeleteUsersRequest) error {
	if len(rq.Usernames) == 0 {
		return errno.ErrInvalidArgument.WithMessage("at least one username is required")
	}

	for _, username := range rq.Usernames {
		if err := validateUsername(username); err != nil {
			return err
		}
	}

	return nil
}

func validateUsername(username string) error {
	if !usernamePattern.MatchString(strings.TrimSpace(username)) {
		return errno.ErrUsernameInvalid
	}

	return nil
}

func validatePassword(password string) error {
	if len(password) < 6 {
		return errno.ErrPasswordInvalid
	}

	return nil
}

func validateUserProfile(nickname, email, phone string) error {
	if strings.TrimSpace(nickname) == "" || len(nickname) > 30 {
		return errno.ErrInvalidArgument.WithMessage("nickname must be between 1 and 30 characters")
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return errno.ErrInvalidArgument.WithMessage("email is invalid")
	}

	if phone != "" && len(phone) > 20 {
		return errno.ErrInvalidArgument.WithMessage("phone must be at most 20 characters")
	}

	return nil
}

func validateListOptions(opts v1.ListOptions) error {
	if opts.Limit < -1 {
		return errno.ErrInvalidArgument.WithMessage("limit must be greater than or equal to -1")
	}

	if opts.Offset < 0 {
		return errno.ErrInvalidArgument.WithMessage("offset must be greater than or equal to 0")
	}

	return nil
}
