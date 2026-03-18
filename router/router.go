package router

import (
	"GoTwitter/controllers"

	"github.com/go-chi/chi/v5"
)

type Router interface {
	Register(r chi.Router)
}

func SetupRouter(UserRouter Router, TweetRouter Router, TagRouter Router) *chi.Mux {
        chiRouter := chi.NewRouter()

        chiRouter.Get("/ping", controllers.PingHandler)

        UserRouter.Register(chiRouter)
        TweetRouter.Register(chiRouter)
        TagRouter.Register(chiRouter)

        return chiRouter
}
