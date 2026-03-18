-- +goose Up
CREATE TABLE tweet_tags (
    tweet_id BIGINT UNSIGNED NOT NULL,
    tag_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (tweet_id, tag_id),
    FOREIGN KEY (tweet_id) REFERENCES tweets(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE tweet_tags;