// Package seed creates the bootstrap administrator account.
package seed

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/onexstack/onexstack/pkg/authn"
	"github.com/onexstack/onexstack/pkg/store/where"
	"gorm.io/gorm"

	"github.com/onexdemo/iamv2/internal/apiserver/model"
	"github.com/onexdemo/iamv2/internal/apiserver/store"
)

// AdminOptions describes the bootstrap administrator.
type AdminOptions struct {
	Username string
	Password string
	Email    string
}

// EnsureAdmin creates the bootstrap administrator when no administrator exists.
// An existing account with the configured username is promoted instead of
// duplicated.
func EnsureAdmin(ctx context.Context, st store.IStore, opts AdminOptions) error {
	if opts.Username == "" || opts.Password == "" {
		return nil
	}

	count, _, err := st.User().List(ctx, where.F("is_admin", 1))
	if err != nil {
		return fmt.Errorf("count administrators: %w", err)
	}

	if count > 0 {
		return nil
	}

	existing, err := st.User().Get(ctx, where.F("username", opts.Username))
	if err == nil {
		existing.IsAdmin = 1
		existing.Status = 1
		if err := st.User().Update(ctx, existing); err != nil {
			return fmt.Errorf("promote bootstrap administrator: %w", err)
		}

		slog.InfoContext(ctx, "promoted existing user to administrator", "username", opts.Username)

		return nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("look up bootstrap administrator: %w", err)
	}

	password, err := authn.Encrypt(opts.Password)
	if err != nil {
		return fmt.Errorf("encrypt bootstrap administrator password: %w", err)
	}

	email := opts.Email
	if email == "" {
		email = opts.Username + "@localhost"
	}

	admin := &model.UserM{
		Username: opts.Username,
		Password: password,
		Nickname: opts.Username,
		Email:    email,
		Status:   1,
		IsAdmin:  1,
		Extend:   map[string]string{},
	}
	if err := st.User().Create(ctx, admin); err != nil {
		return fmt.Errorf("create bootstrap administrator: %w", err)
	}

	slog.InfoContext(ctx, "created bootstrap administrator", "username", opts.Username)

	return nil
}
