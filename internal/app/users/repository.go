package users

import (
	"BD-v2/internal/app/users/models"
	"context"
)

type Repository interface {
	CreateUser(ctx context.Context, user *models.User) error
	CheckIfUserExist(ctx context.Context, user *models.User) ([]*models.User, error)
	FindUserNickname(ctx context.Context, nickname string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) (*models.User, error)
	GetUsers(limit int, forum, since string, desc bool) ([]*models.User, error)
}
