package controllers

import (
	"GoTwitter/config/env"
	"GoTwitter/services"
	"GoTwitter/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
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
	Id        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toUserResponse(u *models.User) *userResponse {
	if u == nil {
		return nil
	}
	return &userResponse{
		Id:        u.Id,
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
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

func (uc *UserController) RegisterUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Registeruser called in UserController")
	var payload struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	created, err := uc.UserService.CreateUser(r.Context(), &models.User{
		Username: payload.Username,
		Email:    payload.Email,
		Password: payload.Password,
	})
	if err != nil {
		log.Printf("[ERROR] signup create user failed: %v", err)
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}
	if created == nil {
		http.Error(w, "user not created", http.StatusInternalServerError)
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

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(toUserResponse(created))
}

func (uc *UserController) Login(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	user, token, err := uc.UserService.Login(r.Context(), payload.Email, payload.Password)
	if err != nil {
		log.Printf("[WARN] login failed: %v", err)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	setAuthCookie(w, token)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(toUserResponse(user))
}

func (uc *UserController) Logout(w http.ResponseWriter, r *http.Request) {
	clearAuthCookie(w)
	w.WriteHeader(http.StatusNoContent)
}
