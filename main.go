package main

import (
	"log"

	"GoTwitter/app"
	dbconfig "GoTwitter/config/db"
	env "GoTwitter/config/env"
	db "GoTwitter/db/repositories"
)

func main() {
	env.Load()

	cfg := app.NewConfig()
	conn, err := dbconfig.SetupDB()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	store := db.NewStorage(conn)
	application := app.NewApplication(cfg, store)

	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
