-- +goose Up
ALTER TABLE tweets
    ADD COLUMN parent_tweet_id BIGINT UNSIGNED NULL AFTER user_id,
    ADD CONSTRAINT fk_tweets_parent_tweet
        FOREIGN KEY (parent_tweet_id) REFERENCES tweets(id) ON DELETE CASCADE;

CREATE INDEX idx_tweets_parent_tweet_id ON tweets(parent_tweet_id);

CREATE TABLE tweet_likes (
    tweet_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (tweet_id, user_id),
    FOREIGN KEY (tweet_id) REFERENCES tweets(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE tweet_retweets (
    tweet_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (tweet_id, user_id),
    FOREIGN KEY (tweet_id) REFERENCES tweets(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE user_follows (
    follower_id BIGINT UNSIGNED NOT NULL,
    followee_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (follower_id, followee_id),
    FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (followee_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE media_attachments (
    id SERIAL PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    tweet_id BIGINT UNSIGNED NULL,
    s3_key VARCHAR(512) NOT NULL UNIQUE,
    url VARCHAR(1024) NOT NULL,
    mime_type VARCHAR(255) NOT NULL,
    size_bytes BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (tweet_id) REFERENCES tweets(id) ON DELETE CASCADE
);

CREATE INDEX idx_media_attachments_tweet_id ON media_attachments(tweet_id);

-- +goose Down
DROP TABLE media_attachments;
DROP TABLE user_follows;
DROP TABLE tweet_retweets;
DROP TABLE tweet_likes;
DROP INDEX idx_tweets_parent_tweet_id ON tweets;
ALTER TABLE tweets
    DROP FOREIGN KEY fk_tweets_parent_tweet,
    DROP COLUMN parent_tweet_id;
