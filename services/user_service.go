package services

import (
	apperrors "GoTwitter/errors"
	db "GoTwitter/db/repositories"
	"GoTwitter/models"
	"GoTwitter/utils"
	"context"
	"fmt"
	"net/http"
)

type UserService interface {
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	Login(ctx context.Context, email string, password string) (*models.User, string, error)
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
		return nil, apperrors.NewAppError("user is nil", http.StatusBadRequest, nil)
	}
	if user.Password == "" {
		return nil, apperrors.NewAppError("password is required", http.StatusBadRequest, nil)
	}

	hashed, err := utils.HashPassword(user.Password)
	if err != nil {
		return nil, apperrors.NewAppError("failed to hash password", http.StatusInternalServerError, err)
	}
	user.Password = string(hashed)

	createdUser, err := u.userRepository.Create(ctx, user)
	if err != nil {
		return nil, apperrors.NewAppError("failed to create user", http.StatusInternalServerError, err)
	}

	return createdUser, nil
}

func (u *UserServiceImpl) Login(ctx context.Context, email string, password string) (*models.User, string, error) {
	if email == "" || password == "" {
		return nil, "", apperrors.NewAppError("email and password are required", http.StatusBadRequest, nil)
	}

	user, err := u.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, "", apperrors.NewAppError("error while fetching user", http.StatusInternalServerError, err)
	}
	if user == nil {
		return nil, "", apperrors.NewAppError("invalid credentials", http.StatusUnauthorized, nil)
	}

	if !utils.CheckPasswordHash(password, user.Password) {
		return nil, "", apperrors.NewAppError("invalid credentials", http.StatusUnauthorized, nil)
	}

	token, err := utils.GenerateJWT(user.Id, user.Email)
	if err != nil {
		return nil, "", apperrors.NewAppError("failed to generate token", http.StatusInternalServerError, err)
	}
	return user, token, nil
}
