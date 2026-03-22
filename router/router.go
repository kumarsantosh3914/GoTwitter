package router

import (
	"GoTwitter/controllers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Router interface {
	Register(r chi.Router)
}

func SetupRouter(UserRouter Router, TweetRouter Router, TagRouter Router) *chi.Mux {
	chiRouter := chi.NewRouter()

	chiRouter.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	chiRouter.Use(middleware.Logger)

	chiRouter.Get("/ping", controllers.PingHandler)

	UserRouter.Register(chiRouter)
	TweetRouter.Register(chiRouter)
	TagRouter.Register(chiRouter)

	return chiRouter
}
