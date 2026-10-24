// Package conversion translates storage models into public API types.
package conversion

import (
	"github.com/onexdemo/iamv2/internal/apiserver/model"
	v1 "github.com/onexdemo/iamv2/pkg/api/apiserver/v1"
)

// UserMToUserV1 converts a user model to its public representation.
func UserMToUserV1(userM *model.UserM) *v1.User {
	if userM == nil {
		return nil
	}

	user := &v1.User{
		ID:         userM.ID,
		InstanceID: userM.UserID,
		Name:       userM.Username,
		Extend:     userM.Extend,
		CreatedAt:  userM.CreatedAt,
		UpdatedAt:  userM.UpdatedAt,
		Nickname:   userM.Nickname,
		Email:      userM.Email,
		Phone:      userM.Phone,
		Status:     userM.Status,
		IsAdmin:    userM.IsAdmin,
	}

	if userM.LoginedAt != nil {
		user.LoginedAt = *userM.LoginedAt
	}

	return user
}

// UserMListToUserV1List converts a slice of user models to public
// representations.
func UserMListToUserV1List(users []*model.UserM) []*v1.User {
	out := make([]*v1.User, 0, len(users))
	for _, user := range users {
		out = append(out, UserMToUserV1(user))
	}

	return out
}
