package db

import "database/sql"

// Facilitates dependency injection for repository
type Storage struct {
        DB             *sql.DB
        UserRepository UserRepository
        TweetRepository TweetRepository
        TagRepository  TagRepository
}

func NewStorage(db *sql.DB) *Storage {
        return &Storage{
                DB:             db,
                UserRepository: NewUserRepository(db),
                TweetRepository: NewTweetRepository(db),
                TagRepository:  NewTagRepository(db),
        }
}
