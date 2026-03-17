package services

import (
	db "GoTwitter/db/repositories"
	"GoTwitter/models"
	"GoTwitter/utils"
	"context"
	"errors"
	"fmt"
)

type UserService interface {
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
}

type UserServiceImpl struct {
	userRepository db.UserRepository
}

func NewUserService(_userRepository db.UserRepository) UserService {
	return &UserServiceImpl{
		userRepository: _userRepository,
	}
}

func (u *UserServiceImpl) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	fmt.Println("Creating user in userService")
	if user == nil {
		return nil, errors.New("user is nil")
	}
	if user.Password == "" {
		return nil, errors.New("password is required")
	}

	hashed, err := utils.HashPassword(user.Password)
	if err != nil {
		return nil, err
	}
	user.Password = string(hashed)

	return u.userRepository.Create(ctx, user)
}
