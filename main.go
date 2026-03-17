package main

import (
	"log"

	"GoTwitter/app"
)

func main() {
	cfg := app.NewConfig()
	application := app.NewApplication(cfg)

	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
