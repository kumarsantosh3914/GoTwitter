package db

import (
	env "GoTwitter/config/env"
	"database/sql"
	"log"

	"github.com/go-sql-driver/mysql"
)

func SetupDB() (*sql.DB, error) {
	cfg := mysql.NewConfig()

	cfg.User = env.GetString("DB_USER", "root")
	cfg.Passwd = env.GetString("DB_PASSWORD", "")
	cfg.Net = env.GetString("DB_NET", "tcp")
	cfg.Addr = env.GetString("DB_ADDR", "127.0.0.1:3306")

	cfg.DBName = env.GetString("DB_NAME", env.GetString("DBName", "twitter_dev"))
	cfg.ParseTime = true

	log.Println("[INFO] Connecting to database:", cfg.DBName)

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	log.Println("[INFO] Connected to database successfully:", cfg.DBName)

	return db, nil
}
