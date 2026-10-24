// Package user implements the user business logic.
package user

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/onexstack/onexstack/pkg/authn"
	"github.com/onexstack/onexstack/pkg/store/where"
	"github.com/onexstack/onexstack/pkg/token"
	"gorm.io/gorm"

	bizutil "github.com/onexdemo/iamv2/internal/apiserver/biz/v1/bizutil"
	model "github.com/onexdemo/iamv2/internal/apiserver/model"
	"github.com/onexdemo/iamv2/internal/apiserver/pkg/conversion"
	"github.com/onexdemo/iamv2/internal/apiserver/store"
	"github.com/onexdemo/iamv2/internal/pkg/contextx"
	"github.com/onexdemo/iamv2/internal/pkg/errno"
	"github.com/onexdemo/iamv2/internal/pkg/notifier"
	v1 "github.com/onexdemo/iamv2/pkg/api/apiserver/v1"
)

// UserBiz defines the user business operations.
type UserBiz interface {
	Create(ctx context.Context, rq *v1.CreateUserRequest) (*v1.CreateUserResponse, error)
	Update(ctx context.Context, rq *v1.UpdateUserRequest) (*v1.UpdateUserResponse, error)
	Delete(ctx context.Context, rq *v1.DeleteUserRequest) (*v1.DeleteUserResponse, error)
	DeleteCollection(ctx context.Context, rq *v1.DeleteUsersRequest) (*v1.DeleteUsersResponse, error)
	ChangePassword(ctx context.Context, rq *v1.ChangePasswordRequest) (*v1.ChangePasswordResponse, error)
	Get(ctx context.Context, rq *v1.GetUserRequest) (*v1.GetUserResponse, error)
	List(ctx context.Context, rq *v1.ListUsersRequest) (*v1.ListUsersResponse, error)
	Login(ctx context.Context, rq *v1.LoginRequest) (*v1.LoginResponse, error)
	RefreshToken(ctx context.Context) (*v1.LoginResponse, error)

	// Authenticate verifies a username/password pair.
	Authenticate(ctx context.Context, username, password string) (*model.UserM, error)
	// GetByUserID returns a user by its stable identifier.
	GetByUserID(ctx context.Context, userID string) (*model.UserM, error)
}

type userBiz struct {
	store    store.IStore
	notifier notifier.Notifier
}

var _ UserBiz = (*userBiz)(nil)

// New creates a user business module.
func New(store store.IStore, notifier notifier.Notifier) *userBiz {
	return &userBiz{store: store, notifier: notifier}
}

// Create registers a new user account.
func (b *userBiz) Create(ctx context.Context, rq *v1.CreateUserRequest) (*v1.CreateUserResponse, error) {
	username := strings.TrimSpace(rq.Name)

	if _, err := b.store.User().Get(ctx, where.F("username", username)); err == nil {
		return nil, errno.ErrUserAlreadyExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errno.ErrDBRead.WithMessage("%s", err.Error())
	}

	password, err := authn.Encrypt(rq.Password)
	if err != nil {
		return nil, errno.ErrInternal.WithMessage("failed to encrypt password")
	}

	userM := &model.UserM{
		Username: username,
		Password: password,
		Nickname: strings.TrimSpace(rq.Nickname),
		Email:    strings.TrimSpace(rq.Email),
		Phone:    strings.TrimSpace(rq.Phone),
		Status:   1,
		IsAdmin:  0,
		Extend:   rq.Extend,
	}

	if err := b.store.User().Create(ctx, userM); err != nil {
		slog.ErrorContext(ctx, "failed to create user", "error", err)

		return nil, errno.ErrDBWrite.WithMessage("%s", err.Error())
	}

	return &v1.CreateUserResponse{User: conversion.UserMToUserV1(userM)}, nil
}

// Update modifies a user's profile.
func (b *userBiz) Update(ctx context.Context, rq *v1.UpdateUserRequest) (*v1.UpdateUserResponse, error) {
	username := strings.TrimSpace(rq.Name)
	if err := bizutil.AuthorizeSelf(ctx, username); err != nil {
		return nil, err
	}

	userM, err := b.getByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	userM.Nickname = strings.TrimSpace(rq.Nickname)
	userM.Email = strings.TrimSpace(rq.Email)
	userM.Phone = strings.TrimSpace(rq.Phone)
	userM.Extend = rq.Extend

	if err := b.store.User().Update(ctx, userM); err != nil {
		return nil, errno.ErrDBWrite.WithMessage("%s", err.Error())
	}

	return &v1.UpdateUserResponse{User: conversion.UserMToUserV1(userM)}, nil
}

// Delete removes a user and every policy it owns.
func (b *userBiz) Delete(ctx context.Context, rq *v1.DeleteUserRequest) (*v1.DeleteUserResponse, error) {
	username := strings.TrimSpace(rq.Username)
	if !contextx.IsAdmin(ctx) {
		return nil, errno.ErrPermissionDenied
	}

	if err := b.deleteUserData(ctx, []string{username}); err != nil {
		return nil, errno.ErrDBWrite.WithMessage("%s", err.Error())
	}

	b.notify(ctx)

	return &v1.DeleteUserResponse{}, nil
}

// DeleteCollection removes several users and their policies.
func (b *userBiz) DeleteCollection(ctx context.Context, rq *v1.DeleteUsersRequest) (*v1.DeleteUsersResponse, error) {
	if !contextx.IsAdmin(ctx) {
		return nil, errno.ErrPermissionDenied
	}

	usernames := make([]string, 0, len(rq.Usernames))
	for _, username := range rq.Usernames {
		usernames = append(usernames, strings.TrimSpace(username))
	}

	if err := b.deleteUserData(ctx, usernames); err != nil {
		return nil, errno.ErrDBWrite.WithMessage("%s", err.Error())
	}

	b.notify(ctx)

	return &v1.DeleteUsersResponse{}, nil
}

// ChangePassword verifies the old password and stores a new one.
func (b *userBiz) ChangePassword(ctx context.Context, rq *v1.ChangePasswordRequest) (*v1.ChangePasswordResponse, error) {
	username := contextx.Username(ctx)
	if username == "" {
		return nil, errno.ErrUnauthenticated
	}

	userM, err := b.getByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	if err := authn.Compare(userM.Password, rq.OldPassword); err != nil {
		return nil, errno.ErrPasswordIncorrect
	}

	password, err := authn.Encrypt(rq.NewPassword)
	if err != nil {
		return nil, errno.ErrInternal.WithMessage("failed to encrypt password")
	}

	userM.Password = password

	if err := b.store.User().Update(ctx, userM); err != nil {
		return nil, errno.ErrDBWrite.WithMessage("%s", err.Error())
	}

	return &v1.ChangePasswordResponse{}, nil
}

// Get returns a single user.
func (b *userBiz) Get(ctx context.Context, rq *v1.GetUserRequest) (*v1.GetUserResponse, error) {
	userM, err := b.getByUsername(ctx, strings.TrimSpace(rq.Username))
	if err != nil {
		return nil, err
	}

	return &v1.GetUserResponse{User: conversion.UserMToUserV1(userM)}, nil
}

// List returns users with their policy counts.
func (b *userBiz) List(ctx context.Context, rq *v1.ListUsersRequest) (*v1.ListUsersResponse, error) {
	count, users, err := b.store.User().List(ctx, bizutil.ListOptions(rq.Name, rq.Offset, rq.Limit))
	if err != nil {
		return nil, errno.ErrDBRead.WithMessage("%s", err.Error())
	}

	items := conversion.UserMListToUserV1List(users)
	for i, item := range items {
		total, _, err := b.store.Policy().List(ctx, where.F("username", users[i].Username))
		if err != nil {
			return nil, errno.ErrDBRead.WithMessage("%s", err.Error())
		}

		item.TotalPolicy = total
	}

	return &v1.ListUsersResponse{
		TotalCount: count,
		Items:      items,
	}, nil
}

// Login authenticates a user and returns a signed token.
func (b *userBiz) Login(ctx context.Context, rq *v1.LoginRequest) (*v1.LoginResponse, error) {
	userM, err := b.Authenticate(ctx, rq.Username, rq.Password)
	if err != nil {
		return nil, err
	}

	return b.issueToken(ctx, userM)
}

// RefreshToken issues a new token for the authenticated caller.
func (b *userBiz) RefreshToken(ctx context.Context) (*v1.LoginResponse, error) {
	userID := contextx.UserID(ctx)
	if userID == "" {
		return nil, errno.ErrUnauthenticated
	}

	userM, err := b.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return b.issueToken(ctx, userM)
}

// Authenticate verifies a username/password pair and records the login time.
func (b *userBiz) Authenticate(ctx context.Context, username, password string) (*model.UserM, error) {
	userM, err := b.getByUsername(ctx, strings.TrimSpace(username))
	if err != nil {
		return nil, errno.ErrUserNotFound
	}

	if userM.Status != 1 {
		return nil, errno.ErrUnauthenticated.WithMessage("the account has been disabled")
	}

	if err := authn.Compare(userM.Password, password); err != nil {
		return nil, errno.ErrPasswordIncorrect
	}

	now := time.Now()
	userM.LoginedAt = &now
	if err := b.store.User().Update(ctx, userM); err != nil {
		slog.WarnContext(ctx, "failed to record login time", "error", err)
	}

	return userM, nil
}

// GetByUserID returns a user by its stable identifier.
func (b *userBiz) GetByUserID(ctx context.Context, userID string) (*model.UserM, error) {
	userM, err := b.store.User().Get(ctx, where.F("user_id", userID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrUserNotFound
		}

		return nil, errno.ErrDBRead.WithMessage("%s", err.Error())
	}

	return userM, nil
}

// deleteUserData removes a user together with their policies and access keys.
// Deleted policies are archived so that iam-watcher can apply the retention
// window.
func (b *userBiz) deleteUserData(ctx context.Context, usernames []string) error {
	return b.store.TX(ctx, func(txCtx context.Context) error {
		_, policies, err := b.store.Policy().List(txCtx, where.NewWhere().Q("username IN (?)", usernames))
		if err != nil {
			return err
		}

		audits := make([]*model.PolicyAuditM, 0, len(policies))
		for _, policyM := range policies {
			audits = append(audits, policyM.ToAudit())
		}
		if err := b.store.PolicyAudit().CreateBatch(txCtx, audits); err != nil {
			return err
		}

		if err := b.store.Policy().Delete(txCtx, where.NewWhere().Q("username IN (?)", usernames)); err != nil {
			return err
		}

		if err := b.store.Secret().Delete(txCtx, where.NewWhere().Q("username IN (?)", usernames)); err != nil {
			return err
		}

		return b.store.User().Delete(txCtx, where.NewWhere().Q("username IN (?)", usernames))
	})
}

func (b *userBiz) notify(ctx context.Context) {
	if err := b.notifier.Notify(ctx, notifier.SecretChanged); err != nil {
		slog.WarnContext(ctx, "failed to publish secret change notification", "error", err)
	}
	if err := b.notifier.Notify(ctx, notifier.PolicyChanged); err != nil {
		slog.WarnContext(ctx, "failed to publish policy change notification", "error", err)
	}
}

func (b *userBiz) getByUsername(ctx context.Context, username string) (*model.UserM, error) {
	userM, err := b.store.User().Get(ctx, where.F("username", username))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrUserNotFound
		}

		return nil, errno.ErrDBRead.WithMessage("%s", err.Error())
	}

	return userM, nil
}

func (b *userBiz) issueToken(ctx context.Context, userM *model.UserM) (*v1.LoginResponse, error) {
	tokenStr, expireAt, err := token.Sign(userM.UserID)
	if err != nil {
		return nil, errno.ErrSignToken.WithMessage("%s", err.Error())
	}

	slog.InfoContext(ctx, "issued login token", "username", userM.Username)

	return &v1.LoginResponse{
		Token:    tokenStr,
		ExpireAt: expireAt,
		User:     conversion.UserMToUserV1(userM),
	}, nil
}
