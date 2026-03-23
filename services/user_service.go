package services

import (
	db "GoTwitter/db/repositories"
	apperrors "GoTwitter/errors"
	"GoTwitter/models"
	"GoTwitter/utils"
	"context"
	"net/http"
)

type UserService interface {
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	Login(ctx context.Context, email string, password string) (*models.User, string, error)
	ListUsers(ctx context.Context, page int, pageSize int, viewerID *int64) ([]*models.User, error)
	GetUserByID(ctx context.Context, id int64, viewerID *int64) (*models.User, error)
	UpdateUser(ctx context.Context, actorID int64, user *models.User) error
	DeleteUser(ctx context.Context, actorID int64, id int64) error
	FollowUser(ctx context.Context, followerID int64, followeeID int64) (*models.User, error)
	UnfollowUser(ctx context.Context, followerID int64, followeeID int64) (*models.User, error)
	ListFollowers(ctx context.Context, userID int64, page int, pageSize int, viewerID *int64) ([]*models.User, error)
	ListFollowing(ctx context.Context, userID int64, page int, pageSize int, viewerID *int64) ([]*models.User, error)
}

type UserServiceImpl struct {
	userRepository   db.UserRepository
	socialRepository db.SocialRepository
}

func NewUserService(userRepository db.UserRepository, socialRepository db.SocialRepository) UserService {
	return &UserServiceImpl{
		userRepository:   userRepository,
		socialRepository: socialRepository,
	}
}

func (u *UserServiceImpl) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	if user == nil {
		return nil, apperrors.NewAppError("user is nil", http.StatusBadRequest, nil)
	}
	if user.Password == "" {
		return nil, apperrors.NewAppError("password is required", http.StatusBadRequest, nil)
	}

	existingEmail, err := u.userRepository.GetUserByEmail(ctx, user.Email)
	if err != nil {
		return nil, apperrors.NewAppError("error checking email uniqueness", http.StatusInternalServerError, err)
	}
	if existingEmail != nil {
		return nil, apperrors.NewAppError("email already in use", http.StatusConflict, nil)
	}

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
	user.Password = hashed

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

func (u *UserServiceImpl) ListUsers(ctx context.Context, page int, pageSize int, viewerID *int64) ([]*models.User, error) {
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
	if err := u.enrichUsers(ctx, users, viewerID); err != nil {
		return nil, err
	}
	return users, nil
}

func (u *UserServiceImpl) GetUserByID(ctx context.Context, id int64, viewerID *int64) (*models.User, error) {
	user, err := u.userRepository.GetByID(ctx, id)
	if err != nil {
		return nil, apperrors.NewAppError("failed to fetch user", http.StatusInternalServerError, err)
	}
	if user == nil {
		return nil, apperrors.NewAppError("user not found", http.StatusNotFound, nil)
	}
	if err := u.enrichUsers(ctx, []*models.User{user}, viewerID); err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserServiceImpl) UpdateUser(ctx context.Context, actorID int64, user *models.User) error {
	if user == nil {
		return apperrors.NewAppError("user is nil", http.StatusBadRequest, nil)
	}
	if actorID != user.Id {
		return apperrors.NewAppError("unauthorized: only the account owner can update the user", http.StatusForbidden, nil)
	}

	existing, err := u.userRepository.GetByID(ctx, user.Id)
	if err != nil {
		return apperrors.NewAppError("failed to fetch user", http.StatusInternalServerError, err)
	}
	if existing == nil {
		return apperrors.NewAppError("user not found", http.StatusNotFound, nil)
	}
	if user.Email != existing.Email {
		existingEmail, _ := u.userRepository.GetUserByEmail(ctx, user.Email)
		if existingEmail != nil {
			return apperrors.NewAppError("email already in use", http.StatusConflict, nil)
		}
	}
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

func (u *UserServiceImpl) DeleteUser(ctx context.Context, actorID int64, id int64) error {
	if actorID != id {
		return apperrors.NewAppError("unauthorized: only the account owner can delete the user", http.StatusForbidden, nil)
	}

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

func (u *UserServiceImpl) FollowUser(ctx context.Context, followerID int64, followeeID int64) (*models.User, error) {
	if followerID == followeeID {
		return nil, apperrors.NewAppError("users cannot follow themselves", http.StatusBadRequest, nil)
	}
	followee, err := u.userRepository.GetByID(ctx, followeeID)
	if err != nil {
		return nil, apperrors.NewAppError("failed to fetch user", http.StatusInternalServerError, err)
	}
	if followee == nil {
		return nil, apperrors.NewAppError("user not found", http.StatusNotFound, nil)
	}
	if err := u.socialRepository.FollowUser(ctx, followerID, followeeID); err != nil {
		return nil, apperrors.NewAppError("failed to follow user", http.StatusInternalServerError, err)
	}
	return u.GetUserByID(ctx, followeeID, &followerID)
}

func (u *UserServiceImpl) UnfollowUser(ctx context.Context, followerID int64, followeeID int64) (*models.User, error) {
	followee, err := u.userRepository.GetByID(ctx, followeeID)
	if err != nil {
		return nil, apperrors.NewAppError("failed to fetch user", http.StatusInternalServerError, err)
	}
	if followee == nil {
		return nil, apperrors.NewAppError("user not found", http.StatusNotFound, nil)
	}
	if err := u.socialRepository.UnfollowUser(ctx, followerID, followeeID); err != nil {
		return nil, apperrors.NewAppError("failed to unfollow user", http.StatusInternalServerError, err)
	}
	return u.GetUserByID(ctx, followeeID, &followerID)
}

func (u *UserServiceImpl) ListFollowers(ctx context.Context, userID int64, page int, pageSize int, viewerID *int64) ([]*models.User, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	user, err := u.userRepository.GetByID(ctx, userID)
	if err != nil {
		return nil, apperrors.NewAppError("failed to fetch user", http.StatusInternalServerError, err)
	}
	if user == nil {
		return nil, apperrors.NewAppError("user not found", http.StatusNotFound, nil)
	}

	ids, err := u.socialRepository.ListFollowers(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, apperrors.NewAppError("failed to fetch followers", http.StatusInternalServerError, err)
	}
	return u.usersByIDs(ctx, ids, viewerID)
}

func (u *UserServiceImpl) ListFollowing(ctx context.Context, userID int64, page int, pageSize int, viewerID *int64) ([]*models.User, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	user, err := u.userRepository.GetByID(ctx, userID)
	if err != nil {
		return nil, apperrors.NewAppError("failed to fetch user", http.StatusInternalServerError, err)
	}
	if user == nil {
		return nil, apperrors.NewAppError("user not found", http.StatusNotFound, nil)
	}

	ids, err := u.socialRepository.ListFollowing(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, apperrors.NewAppError("failed to fetch following users", http.StatusInternalServerError, err)
	}
	return u.usersByIDs(ctx, ids, viewerID)
}

func (u *UserServiceImpl) usersByIDs(ctx context.Context, ids []int64, viewerID *int64) ([]*models.User, error) {
	users, err := u.userRepository.GetByIDs(ctx, ids)
	if err != nil {
		return nil, apperrors.NewAppError("failed to fetch users", http.StatusInternalServerError, err)
	}
	if err := u.enrichUsers(ctx, users, viewerID); err != nil {
		return nil, err
	}
	return users, nil
}

func (u *UserServiceImpl) enrichUsers(ctx context.Context, users []*models.User, viewerID *int64) error {
	if len(users) == 0 || viewerID == nil {
		return nil
	}

	userIDs := make([]int64, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.Id)
	}

	states, err := u.socialRepository.GetUserFollowStates(ctx, *viewerID, userIDs)
	if err != nil {
		return apperrors.NewAppError("failed to fetch follow relationships", http.StatusInternalServerError, err)
	}
	for _, user := range users {
		user.IsFollowing = states[user.Id]
	}
	return nil
}
