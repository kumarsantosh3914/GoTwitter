package controllers

import (
	"GoTwitter/config/env"
	apperrors "GoTwitter/errors"
	"GoTwitter/models"
	"GoTwitter/services"
	"GoTwitter/utils"
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

type UserController struct {
	UserService services.UserService
}

func NewUserController(_userService services.UserService) *UserController {
	return &UserController{
		UserService: _userService,
	}
}

type userResponse struct {
	Id             int64     `json:"id"`
	Username       string    `json:"username"`
	Email          string    `json:"email"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	FollowerCount  int64     `json:"follower_count"`
	FollowingCount int64     `json:"following_count"`
	IsFollowing    bool      `json:"is_following"`
}

func toUserResponse(u *models.User) *userResponse {
	if u == nil {
		return nil
	}
	return &userResponse{
		Id:             u.Id,
		Username:       u.Username,
		Email:          u.Email,
		CreatedAt:      u.CreatedAt,
		UpdatedAt:      u.UpdatedAt,
		FollowerCount:  u.FollowerCount,
		FollowingCount: u.FollowingCount,
		IsFollowing:    u.IsFollowing,
	}
}

func toUserResponses(users []*models.User) []*userResponse {
	res := make([]*userResponse, len(users))
	for i, u := range users {
		res[i] = toUserResponse(u)
	}
	return res
}

func setAuthCookie(w http.ResponseWriter, token string) {
	secure := strings.EqualFold(env.GetString("COOKIE_SECURE", "false"), "true")
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})
}

func clearAuthCookie(w http.ResponseWriter) {
	secure := strings.EqualFold(env.GetString("COOKIE_SECURE", "false"), "true")
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
		MaxAge:   -1,
	})
}

func handleError(w http.ResponseWriter, err error) {
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		apperrors.WriteError(w, appErr)
		return
	}
	apperrors.WriteError(w, apperrors.NewAppError("internal server error", http.StatusInternalServerError, err))
}

func (uc *UserController) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Username string `json:"username" validate:"required,min=3,max=30"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
	}

	if err := utils.ReadJsonBody(r, &payload); err != nil {
		handleError(w, apperrors.NewAppError("invalid json body", http.StatusBadRequest, err))
		return
	}

	if err := utils.Validator.Struct(payload); err != nil {
		handleError(w, apperrors.NewAppError("validation failed", http.StatusBadRequest, err))
		return
	}

	created, err := uc.UserService.CreateUser(r.Context(), &models.User{
		Username: payload.Username,
		Email:    payload.Email,
		Password: payload.Password,
	})
	if err != nil {
		log.Printf("[ERROR] signup create user failed: %v", err)
		handleError(w, err)
		return
	}

	// Issue auth cookie on signup
	user, token, err := uc.UserService.Login(r.Context(), payload.Email, payload.Password)
	if err == nil && token != "" {
		setAuthCookie(w, token)
		created = user
	} else if err != nil {
		log.Printf("[WARN] signup auto-login failed: %v", err)
	}

	utils.WriteJsonSuccessResponse(w, http.StatusCreated, "User registered successfully", toUserResponse(created))
}

func (uc *UserController) Login(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	if err := utils.ReadJsonBody(r, &payload); err != nil {
		handleError(w, apperrors.NewAppError("invalid json body", http.StatusBadRequest, err))
		return
	}

	if err := utils.Validator.Struct(payload); err != nil {
		handleError(w, apperrors.NewAppError("validation failed", http.StatusBadRequest, err))
		return
	}

	user, token, err := uc.UserService.Login(r.Context(), payload.Email, payload.Password)
	if err != nil {
		log.Printf("[WARN] login failed: %v", err)
		handleError(w, err)
		return
	}

	setAuthCookie(w, token)
	utils.WriteJsonSuccessResponse(w, http.StatusOK, "Login successful", toUserResponse(user))
}

func (uc *UserController) Logout(w http.ResponseWriter, r *http.Request) {
	clearAuthCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (uc *UserController) ListUsers(w http.ResponseWriter, r *http.Request) {
	page, err := parsePositiveIntQuery(r.URL.Query().Get("page"), "page", 1)
	if err != nil {
		handleError(w, err)
		return
	}

	pageSize, err := parsePositiveIntQuery(r.URL.Query().Get("page_size"), "page_size", 10)
	if err != nil {
		handleError(w, err)
		return
	}

	var viewerID *int64
	if claims, ok := utils.GetUserFromRequest(r); ok {
		viewerID = &claims.UserID
	}

	users, err := uc.UserService.ListUsers(r.Context(), page, pageSize, viewerID)
	if err != nil {
		handleError(w, err)
		return
	}

	utils.WriteJsonSuccessResponse(w, http.StatusOK, "Users fetched successfully", map[string]any{
		"items": toUserResponses(users),
		"meta": paginationMeta{
			Page:     page,
			PageSize: pageSize,
			Count:    len(users),
		},
	})
}

func (uc *UserController) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		handleError(w, apperrors.NewAppError("invalid user id", http.StatusBadRequest, err))
		return
	}

	var viewerID *int64
	if claims, ok := utils.GetUserFromRequest(r); ok {
		viewerID = &claims.UserID
	}

	user, err := uc.UserService.GetUserByID(r.Context(), id, viewerID)
	if err != nil {
		handleError(w, err)
		return
	}

	utils.WriteJsonSuccessResponse(w, http.StatusOK, "User fetched successfully", toUserResponse(user))
}

func (uc *UserController) UpdateUser(w http.ResponseWriter, r *http.Request) {
	claims, ok := utils.GetUserFromContext(r.Context())
	if !ok {
		handleError(w, apperrors.NewAppError("unauthorized", http.StatusUnauthorized, nil))
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		handleError(w, apperrors.NewAppError("invalid user id", http.StatusBadRequest, err))
		return
	}

	var payload struct {
		Username string `json:"username" validate:"required,min=3,max=30"`
		Email    string `json:"email" validate:"required,email"`
	}

	if err := utils.ReadJsonBody(r, &payload); err != nil {
		handleError(w, apperrors.NewAppError("invalid json body", http.StatusBadRequest, err))
		return
	}

	if err := utils.Validator.Struct(payload); err != nil {
		handleError(w, apperrors.NewAppError("validation failed", http.StatusBadRequest, err))
		return
	}

	user := &models.User{
		Id:       id,
		Username: payload.Username,
		Email:    payload.Email,
	}

	if err := uc.UserService.UpdateUser(r.Context(), claims.UserID, user); err != nil {
		handleError(w, err)
		return
	}

	// Fetch updated user to return
	updated, err := uc.UserService.GetUserByID(r.Context(), id, &claims.UserID)
	if err != nil {
		handleError(w, err)
		return
	}

	utils.WriteJsonSuccessResponse(w, http.StatusOK, "User updated successfully", toUserResponse(updated))
}

func (uc *UserController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	claims, ok := utils.GetUserFromContext(r.Context())
	if !ok {
		handleError(w, apperrors.NewAppError("unauthorized", http.StatusUnauthorized, nil))
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		handleError(w, apperrors.NewAppError("invalid user id", http.StatusBadRequest, err))
		return
	}

	if err := uc.UserService.DeleteUser(r.Context(), claims.UserID, id); err != nil {
		handleError(w, err)
		return
	}

	utils.WriteJsonSuccessResponse(w, http.StatusOK, "User deleted successfully", nil)
}

func (uc *UserController) FollowUser(w http.ResponseWriter, r *http.Request) {
	uc.toggleFollow(w, r, uc.UserService.FollowUser, "User followed successfully")
}

func (uc *UserController) UnfollowUser(w http.ResponseWriter, r *http.Request) {
	uc.toggleFollow(w, r, uc.UserService.UnfollowUser, "User unfollowed successfully")
}

func (uc *UserController) ListFollowers(w http.ResponseWriter, r *http.Request) {
	uc.listUserConnections(w, r, uc.UserService.ListFollowers, "Followers fetched successfully")
}

func (uc *UserController) ListFollowing(w http.ResponseWriter, r *http.Request) {
	uc.listUserConnections(w, r, uc.UserService.ListFollowing, "Following users fetched successfully")
}

func (uc *UserController) toggleFollow(
	w http.ResponseWriter,
	r *http.Request,
	action func(context.Context, int64, int64) (*models.User, error),
	message string,
) {
	claims, ok := utils.GetUserFromContext(r.Context())
	if !ok {
		handleError(w, apperrors.NewAppError("unauthorized", http.StatusUnauthorized, nil))
		return
	}

	targetID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		handleError(w, apperrors.NewAppError("invalid user id", http.StatusBadRequest, err))
		return
	}

	user, err := action(r.Context(), claims.UserID, targetID)
	if err != nil {
		handleError(w, err)
		return
	}

	utils.WriteJsonSuccessResponse(w, http.StatusOK, message, toUserResponse(user))
}

func (uc *UserController) listUserConnections(
	w http.ResponseWriter,
	r *http.Request,
	action func(context.Context, int64, int, int, *int64) ([]*models.User, error),
	message string,
) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		handleError(w, apperrors.NewAppError("invalid user id", http.StatusBadRequest, err))
		return
	}

	page, err := parsePositiveIntQuery(r.URL.Query().Get("page"), "page", 1)
	if err != nil {
		handleError(w, err)
		return
	}

	pageSize, err := parsePositiveIntQuery(r.URL.Query().Get("page_size"), "page_size", 10)
	if err != nil {
		handleError(w, err)
		return
	}

	var viewerID *int64
	if claims, ok := utils.GetUserFromRequest(r); ok {
		viewerID = &claims.UserID
	}

	users, err := action(r.Context(), userID, page, pageSize, viewerID)
	if err != nil {
		handleError(w, err)
		return
	}

	utils.WriteJsonSuccessResponse(w, http.StatusOK, message, map[string]any{
		"items": toUserResponses(users),
		"meta": paginationMeta{
			Page:     page,
			PageSize: pageSize,
			Count:    len(users),
		},
	})
}
