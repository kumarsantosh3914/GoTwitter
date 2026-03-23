# GoTwitter API

A Twitter-like REST API built with Go, Chi, and MySQL. The application supports authentication with HTTP-only cookie-based JWTs, tweet CRUD, hashtag extraction, likes, retweets, replies with threaded conversations, follow relationships, and S3 media uploads via presigned URLs.

## Features

### Authentication and users
- Signup, login, and logout
- JWT authentication stored in an HTTP-only `auth_token` cookie
- Public user listing and profile lookup
- Protected self-service user update and delete
- Follow and unfollow relationships
- Follower and following lists
- User response metadata including `follower_count`, `following_count`, and viewer-specific `is_following`

### Tweets
- Create, list, fetch, update, and delete tweets
- 280 character tweet validation
- Automatic hashtag extraction and association
- Filtering by `user_id`, `tag`, and search query `q`
- Like and unlike tweets
- Retweet and unretweet tweets
- Reply to a tweet with `parent_tweet_id`
- Nested replies on `GET /tweets/{id}`
- Thread view on `GET /tweets/{id}/thread`
- Tweet response metadata including `like_count`, `retweet_count`, `reply_count`, `is_liked`, and `is_retweeted`

### Media
- Create S3 presigned upload URLs through the API
- Persist media metadata before upload
- Attach uploaded media to tweets using `media_ids`
- Support for up to 4 media attachments per tweet

### Tags
- Public tag listing
- Popular tags by usage count
- Tag details with paginated associated tweets
- Authenticated tag deletion

## Tech stack

- Go 1.25
- Chi router
- MySQL
- Goose migrations
- go-playground validator
- JWT
- AWS SDK for Go v2 for S3 presigning

## Getting started

### Prerequisites

- Go installed locally
- MySQL running locally or remotely
- `goose` installed if you want to run migrations manually

### Environment

Create a `.env` file in the project root.

```env
# Server
PORT=8080

# Database
DB_ADDR=127.0.0.1:3306
DB_USER=root
DB_PASSWORD=yourpassword
DB_NAME=twitter_dev
DB_NET=tcp

# Auth
JWT_SECRET=your_super_secret_key
COOKIE_SECURE=false

# AWS S3 media uploads
AWS_REGION=ap-south-1
AWS_S3_BUCKET=twitter-bucket
AWS_ACCESS_KEY_ID=your_access_key_id
AWS_SECRET_ACCESS_KEY=your_secret_access_key
AWS_S3_PUBLIC_BASE_URL=
```

Notes:
- `PORT` can be `8080` or `:8080`. The app normalizes it.
- `AWS_S3_PUBLIC_BASE_URL` is optional. If empty, the app uses `https://<bucket>.s3.<region>.amazonaws.com`.
- If the AWS variables are missing, `POST /media/presign` returns `501 Not Implemented`.
- The env loader supports both `#` and `//` comment lines.

### Install dependencies

```bash
go mod tidy
```

### Run migrations

```bash
make migrate-up
```

### Start the server

```bash
go run main.go
```

By default the API listens on `http://localhost:8080`.

## Authentication model

Protected endpoints require the `auth_token` cookie. The simplest manual test flow is to store cookies in a file:

```bash
BASE="http://localhost:8080"
COOKIE_JAR="./cookies.txt"
```

Login example:

```bash
curl -i -X POST "$BASE/login" \
  -H "Content-Type: application/json" \
  -c "$COOKIE_JAR" \
  -d '{
    "email": "alice@example.com",
    "password": "password123"
  }'
```

Then call protected routes with:

```bash
-b "$COOKIE_JAR"
```

## API reference

### Health

| Method | Endpoint | Auth | Description |
| :--- | :--- | :--- | :--- |
| `GET` | `/ping` | No | Health check |

### Auth

| Method | Endpoint | Auth | Description |
| :--- | :--- | :--- | :--- |
| `POST` | `/signup` | No | Register a user and issue an auth cookie |
| `POST` | `/login` | No | Login and issue an auth cookie |
| `POST` | `/logout` | No | Clear the auth cookie |

### Users

| Method | Endpoint | Auth | Description |
| :--- | :--- | :--- | :--- |
| `GET` | `/users` | No | List users |
| `GET` | `/users/{id}` | No | Get user details |
| `PUT` | `/users/{id}` | Yes | Update own profile |
| `DELETE` | `/users/{id}` | Yes | Delete own profile |
| `GET` | `/users/{id}/followers` | No | List followers |
| `GET` | `/users/{id}/following` | No | List followed users |
| `POST` | `/users/{id}/follow` | Yes | Follow a user |
| `DELETE` | `/users/{id}/follow` | Yes | Unfollow a user |

### Tweets

| Method | Endpoint | Auth | Description |
| :--- | :--- | :--- | :--- |
| `POST` | `/tweets/` | Yes | Create a tweet or reply |
| `GET` | `/tweets/` | No | List tweets |
| `GET` | `/tweets/{id}` | No | Get a tweet with nested replies |
| `GET` | `/tweets/{id}/thread` | No | Get the tweet plus its parent chain |
| `PUT` | `/tweets/{id}` | Yes | Update own tweet |
| `DELETE` | `/tweets/{id}` | Yes | Delete own tweet |
| `POST` | `/tweets/{id}/like` | Yes | Like a tweet |
| `DELETE` | `/tweets/{id}/like` | Yes | Unlike a tweet |
| `POST` | `/tweets/{id}/retweet` | Yes | Retweet a tweet |
| `DELETE` | `/tweets/{id}/retweet` | Yes | Undo a retweet |

List filters for `GET /tweets/`:
- `page`
- `page_size`
- `user_id`
- `tag`
- `q`

### Media

| Method | Endpoint | Auth | Description |
| :--- | :--- | :--- | :--- |
| `POST` | `/media/presign` | Yes | Create a presigned S3 upload URL and attachment record |

### Tags

| Method | Endpoint | Auth | Description |
| :--- | :--- | :--- | :--- |
| `GET` | `/tags/` | No | List tags |
| `GET` | `/tags/popular` | No | List popular tags |
| `GET` | `/tags/{id}` | No | Get tag details and tagged tweets |
| `DELETE` | `/tags/{id}` | Yes | Delete a tag |

## Request examples

### Signup

```bash
curl -i -X POST "$BASE/signup" \
  -H "Content-Type: application/json" \
  -c "$COOKIE_JAR" \
  -d '{
    "username": "alice",
    "email": "alice@example.com",
    "password": "password123"
  }'
```

### Create a tweet

```bash
curl -i -X POST "$BASE/tweets/" \
  -H "Content-Type: application/json" \
  -b "$COOKIE_JAR" \
  -d '{
    "tweet": "hello from GoTwitter #golang"
  }'
```

### Reply to a tweet

```bash
curl -i -X POST "$BASE/tweets/" \
  -H "Content-Type: application/json" \
  -b "$COOKIE_JAR" \
  -d '{
    "tweet": "this is a reply",
    "parent_tweet_id": 1
  }'
```

### Like and retweet

```bash
curl -i -X POST "$BASE/tweets/1/like" -b "$COOKIE_JAR"
curl -i -X POST "$BASE/tweets/1/retweet" -b "$COOKIE_JAR"
```

### Follow a user

```bash
curl -i -X POST "$BASE/users/2/follow" \
  -b "$COOKIE_JAR"
```

### Create a presigned upload

```bash
curl -i -X POST "$BASE/media/presign" \
  -H "Content-Type: application/json" \
  -b "$COOKIE_JAR" \
  -d '{
    "filename": "photo.png",
    "content_type": "image/png",
    "size_bytes": 2048
  }'
```

After that, upload the file directly to the returned `upload_url` using `PUT`, then attach the returned `attachment.id` to a tweet:

```bash
curl -i -X POST "$BASE/tweets/" \
  -H "Content-Type: application/json" \
  -b "$COOKIE_JAR" \
  -d '{
    "tweet": "tweet with image #photo",
    "media_ids": [1]
  }'
```

## Response shape

Successful responses use a consistent envelope:

```json
{
  "status": "success",
  "message": "Tweets fetched successfully",
  "data": {}
}
```

Paginated list responses use:

```json
{
  "status": "success",
  "message": "Users fetched successfully",
  "data": {
    "items": [],
    "meta": {
      "page": 1,
      "page_size": 10,
      "count": 0
    }
  }
}
```

Popular tags use:

```json
{
  "status": "success",
  "message": "Popular tags fetched successfully",
  "data": {
    "items": [],
    "meta": {
      "limit": 10,
      "count": 0
    }
  }
}
```

## Important behavior notes

- Tweet list endpoints currently return top-level tweets only. Replies are fetched through `GET /tweets/{id}` and `GET /tweets/{id}/thread`.
- The authenticated user can only update or delete their own user account and tweets.
- A user cannot follow themselves.
- Media attachments must belong to the authenticated user before they can be attached to a tweet.
- A tweet can include at most 4 media attachments.
- Hashtags are normalized to lowercase.

## Project structure

- `app/`: application bootstrapping and dependency wiring
- `config/`: env and database configuration
- `controllers/`: HTTP handlers and payload validation
- `services/`: business logic
- `db/repositories/`: SQL-backed repository layer
- `models/`: domain models and response entities
- `router/`: route registration and middleware
- `utils/`: auth, validation, and JSON helpers
- `migrations/`: database migrations

## Verification

Current codebase verification:

```bash
go test ./...
```
