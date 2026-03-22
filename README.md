# GoTwitter API

A robust, Twitter-like RESTful API built with Go, focusing on clean architecture, performance, and a seamless hashtag system.

## 🚀 Features

### 👤 User Management
- **Authentication**: JWT-based authentication using HTTP-only cookies.
- **Profiles**: Signup, login, logout, user listing/detail, and protected self-service profile updates/deletes.

### 🐦 Tweet Management
- **Creation**: Create tweets with a **280-character limit**.
- **Hashtag System**: **Automatic extraction** of `#hashtags` from tweet content.
- **Discovery**: List tweets with advanced filtering by `user_id`, `tag`, or search query (`q`) and pagination metadata.
- **Control**: Update and delete tweets (restricted to the original author).
- **Consistency**: Tweet and hashtag writes are handled transactionally.

### 🏷️ Tag Management
- **Popularity**: Track and retrieve the most used hashtags in the system.
- **Association**: Get all tweets associated with a specific hashtag.
- **Maintenance**: List tags publicly and delete tags through authenticated routes.

---

## 🛠️ Tech Stack

- **Language**: Go (1.22+)
- **Router**: [go-chi/chi](https://github.com/go-chi/chi)
- **Database**: MySQL
- **Migrations**: [goose](https://github.com/pressly/goose)
- **Validation**: [go-playground/validator](https://github.com/go-playground/validator)
- **Auth**: JWT (JSON Web Tokens)

---

## 🏁 Getting Started

### Prerequisites
- Go installed on your machine.
- MySQL server running.
- `goose` installed (optional, for manual migrations).

### 1. Environment Setup
Create a `.env` file in the root directory (refer to `.env.example` if available):
```env
PORT=3001
DB_ADDR=127.0.0.1:3306
DB_USER=root
DB_PASSWORD=yourpassword
DB_NAME=twitter_dev
JWT_SECRET=your_super_secret_key
```

### 2. Database Migrations
Run the migrations to set up your schema:
```bash
# Using makefile (if goose is installed)
make migrate-up
```

### 3. Run the Application
```bash
go run main.go
```
The server will start on `http://localhost:3001`.

---

## 📖 API Documentation

### Authentication
| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `POST` | `/signup` | Register a new user |
| `POST` | `/login` | Login and receive an auth cookie |
| `POST` | `/logout` | Clear the auth cookie |

### Users
| Method | Endpoint | Auth | Description |
| :--- | :--- | :--- | :--- |
| `GET` | `/users` | No | List users (paginated) |
| `GET` | `/users/{id}` | No | Get user details |
| `PUT` | `/users/{id}` | Yes | Update own profile |
| `DELETE` | `/users/{id}` | Yes | Delete own profile |

### Tweets
| Method | Endpoint | Auth | Description |
| :--- | :--- | :--- | :--- |
| `POST` | `/tweets` | Yes | Create a tweet (auto-extracts tags) |
| `GET` | `/tweets` | No | List tweets (filters: `user_id`, `tag`, `q`; paginated metadata) |
| `GET` | `/tweets/{id}` | No | Get detailed tweet info |
| `PUT` | `/tweets/{id}` | Yes | Update tweet (Author only) |
| `DELETE` | `/tweets/{id}` | Yes | Delete tweet (Author only) |

### Tags
| Method | Endpoint | Auth | Description |
| :--- | :--- | :--- | :--- |
| `GET` | `/tags` | No | List all tags (paginated) |
| `GET` | `/tags/popular` | No | Get top hashtags by usage count |
| `GET` | `/tags/{id}` | No | Get tag details and associated tweets (paginated tweets metadata) |
| `DELETE` | `/tags/{id}` | Yes | Delete a tag |

### Response Shape

List endpoints now return a consistent envelope with `items` and `meta`:

```json
{
  "status": "success",
  "message": "Tweets fetched successfully",
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

`GET /tags/popular` uses a similar envelope, but the metadata contains `limit` and `count`.

---

## 🧪 Example Usage

### Create a Tweet
```bash
curl -X POST http://localhost:3001/tweets \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{"tweet": "Building a #Twitter clone in #Go is fun! #backend"}'
```

### Search for Tweets
```bash
curl -G "http://localhost:3001/tweets" \
  -d "tag=backend" \
  -d "q=clone"
```

### List Tweets Response
```json
{
  "status": "success",
  "message": "Tweets fetched successfully",
  "data": {
    "items": [
      {
        "id": 1,
        "user_id": 1,
        "tweet": "Building a #Twitter clone in #Go is fun! #backend",
        "created_at": "2026-03-22T10:00:00Z",
        "updated_at": "2026-03-22T10:00:00Z",
        "tags": [
          { "id": 1, "name": "twitter" },
          { "id": 2, "name": "go" },
          { "id": 3, "name": "backend" }
        ]
      }
    ],
    "meta": {
      "page": 1,
      "page_size": 10,
      "count": 1
    }
  }
}
```

---

## 📂 Project Structure
- `/app`: Application bootstrapping and DI.
- `/controllers`: Request handlers and payload validation.
- `/services`: Business logic and orchestration.
- `/db/repositories`: Database abstraction layer (SQL queries).
- `/models`: Domain entities (Structs).
- `/router`: Route definitions and middleware.
- `/utils`: Helpers for JWT, hashing, and JSON.
- `/migrations`: SQL migration files.
