# WASA Messenger

A web-based messaging application developed for the Web and Software
Architecture course. It consists of a Go backend exposing a REST API and a
Vue.js single-page frontend. Users can register, manage conversations and
groups, exchange messages with delivery/read status, and react to messages
with emojis.

## Tech stack

Backend:
- Go 1.21
- SQLite (pure-Go driver `modernc.org/sqlite`, no CGO required)
- `julienschmidt/httprouter` for routing
- `gorilla/handlers` for CORS
- Standard library `log/slog` for logging

Frontend:
- Vue.js 3.2
- Vite (build tool / dev server)
- Axios (HTTP client)
- Vue Router 4

## Features

- User login/registration with token-based authentication
- Profile management (username, profile photo)
- Private (one-to-one) and group conversations
- Sending text and media messages, with reply support
- Message delivery and read status (single check / double check / read)
- Emoji reactions on messages
- Forwarding and deleting messages
- Group management (create, add members, leave, rename, set photo)
- User search

## Project structure

```
wasa_project/
├── cmd/
│   ├── healthcheck/          # Health check utility
│   └── webapi/               # Main API server entry point
│       ├── main.go
│       ├── cors.go
│       ├── register-web-ui.go
│       └── load-configuration.go
├── service/
│   ├── api/                  # HTTP handlers (business logic)
│   │   ├── api.go
│   │   ├── users.go
│   │   ├── conversations.go
│   │   ├── reactions.go
│   │   └── ...
│   ├── database/             # Database layer
│   │   ├── database.go
│   │   ├── users.go
│   │   ├── conversations.go
│   │   ├── reactions.go
│   │   └── schemas.go
│   ├── applog/               # Logging facade over log/slog
│   └── uid/                  # UUID generation
├── webui/                    # Vue.js frontend
│   ├── src/
│   │   ├── views/            # Page views (login, conversations, messages, ...)
│   │   ├── components/       # Shared components
│   │   ├── services/         # Axios API client
│   │   ├── App.vue
│   │   └── main.js
│   ├── package.json
│   └── vite.config.js
├── Dockerfile.backend
├── Dockerfile.frontend
├── go.mod
└── README.md
```

## Running locally

Prerequisites: Go 1.21+, Node.js 20+, npm.

### 1. Clone

```bash
git clone https://github.com/KochonovAltyn/wasa_messenger.git
cd wasa_messenger
```

### 2. Backend

```bash
go mod download
go run ./cmd/webapi
```

The API server starts on `http://localhost:3000`. On first run it creates the
SQLite database file automatically.

### 3. Frontend (in a separate terminal)

```bash
cd webui
npm install
npm run dev
```

The frontend starts on `http://localhost:5173` and talks to the backend on
port 3000.

## Running with Docker

Build and run the two images separately:

```bash
docker build -f Dockerfile.backend -t wasa-backend .
docker build -f Dockerfile.frontend -t wasa-frontend .

docker run -d --name wasa-backend -p 3000:3000 wasa-backend
docker run -d --name wasa-frontend -p 8080:80 wasa-frontend
```

## Main API endpoints

Authentication:
- `POST /session` — login or register (returns the user identifier used as token)

Users:
- `GET /users/:id` — get user details
- `PUT /users/me/username` — update username
- `PUT /users/me/photo` — update profile photo
- `GET /search/users` — search users

Conversations and messages:
- `GET /users/:id/conversations` — list the user's conversations
- `GET /conversations/:c_id` — get a conversation with its messages
- `POST /conversations/:conversation_id/messages` — send a message
- `POST /users/:id/conversations/first-message` — start a new private chat
- `POST /conversations/:conversation_id/messages/:message_id/forward/:target_conversation_id` — forward
- `DELETE /conversations/:conversation_id/messages/:message_id` — delete a message

Reactions:
- `PUT /conversations/:c_id/messages/:message_id/reaction` — set an emoji reaction
- `DELETE /conversations/:conversation_id/messages/:message_id/reaction` — remove a reaction

Groups:
- `POST /groups` — create a group
- `POST /groups/:c_id/members` — add a member
- `DELETE /groups/:c_id/leave` — leave a group
- `PUT /groups/:c_id/name` — rename a group
- `PUT /conversations/:c_id/set-group-photo` — set group photo
