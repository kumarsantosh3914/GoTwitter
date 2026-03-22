package router

import (
	"GoTwitter/controllers"

	"github.com/go-chi/chi/v5"
)

type UserRouter struct {
	userController *controllers.UserController
}

func NewUserRouter(_userController *controllers.UserController) Router {
	return &UserRouter{
		userController: _userController,
	}
}

func (ur *UserRouter) Register(r chi.Router) {
	r.Post("/signup", ur.userController.RegisterUser)
	r.Post("/login", ur.userController.Login)
	r.Post("/logout", ur.userController.Logout)

	r.Route("/users", func(r chi.Router) {
		r.Get("/", ur.userController.ListUsers)
		r.Get("/{id}", ur.userController.GetUser)

		r.Group(func(r chi.Router) {
			r.Use(AuthMiddleware)
			r.Put("/{id}", ur.userController.UpdateUser)
			r.Delete("/{id}", ur.userController.DeleteUser)
		})
	})
}
