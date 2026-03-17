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
	ListUsers(ctx context.Context, page, pageSize int) ([]*models.User, error)
	GetUserByID(ctx context.Context, id int64) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id int64) error
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

	// Check email uniqueness
	existingEmail, err := u.userRepository.GetUserByEmail(ctx, user.Email)
	if err != nil {
		return nil, apperrors.NewAppError("error checking email uniqueness", http.StatusInternalServerError, err)
	}
	if existingEmail != nil {
		return nil, apperrors.NewAppError("email already in use", http.StatusConflict, nil)
	}

	// Check username uniqueness
	existingUsername, err := u.userRepository.GetUserByUsername(ctx, user.Username)
	if err != nil {
		return nil, apperrors.NewAppError("error checking username uniqueness", http.StatusInternalServerError, err)
	}
	if existingUsername != nil {
		return nil, apperrors.NewAppError("username already in use", http.StatusConflict, nil)
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

func (u *UserServiceImpl) ListUsers(ctx context.Context, page, pageSize int) ([]*models.User, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	users, err := u.userRepository.GetAll(ctx, pageSize, offset)
	if err != nil {
		return nil, apperrors.NewAppError("failed to fetch users", http.StatusInternalServerError, err)
	}
	return users, nil
}

func (u *UserServiceImpl) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	user, err := u.userRepository.GetByID(ctx, id)
	if err != nil {
		return nil, apperrors.NewAppError("failed to fetch user", http.StatusInternalServerError, err)
	}
	if user == nil {
		return nil, apperrors.NewAppError("user not found", http.StatusNotFound, nil)
	}
	return user, nil
}

func (u *UserServiceImpl) UpdateUser(ctx context.Context, user *models.User) error {
	if user == nil {
		return apperrors.NewAppError("user is nil", http.StatusBadRequest, nil)
	}

	// Check if user exists
	existing, err := u.userRepository.GetByID(ctx, user.Id)
	if err != nil {
		return apperrors.NewAppError("failed to fetch user", http.StatusInternalServerError, err)
	}
	if existing == nil {
		return apperrors.NewAppError("user not found", http.StatusNotFound, nil)
	}

	// Check email uniqueness if changed
	if user.Email != existing.Email {
		existingEmail, _ := u.userRepository.GetUserByEmail(ctx, user.Email)
		if existingEmail != nil {
			return apperrors.NewAppError("email already in use", http.StatusConflict, nil)
		}
	}

	// Check username uniqueness if changed
	if user.Username != existing.Username {
		existingUsername, _ := u.userRepository.GetUserByUsername(ctx, user.Username)
		if existingUsername != nil {
			return apperrors.NewAppError("username already in use", http.StatusConflict, nil)
		}
	}

	if err := u.userRepository.Update(ctx, user); err != nil {
		return apperrors.NewAppError("failed to update user", http.StatusInternalServerError, err)
	}
	return nil
}

func (u *UserServiceImpl) DeleteUser(ctx context.Context, id int64) error {
	// Check if user exists
	existing, err := u.userRepository.GetByID(ctx, id)
	if err != nil {
		return apperrors.NewAppError("failed to fetch user", http.StatusInternalServerError, err)
	}
	if existing == nil {
		return apperrors.NewAppError("user not found", http.StatusNotFound, nil)
	}

	if err := u.userRepository.DeleteByID(ctx, id); err != nil {
		return apperrors.NewAppError("failed to delete user", http.StatusInternalServerError, err)
	}
	return nil
}
