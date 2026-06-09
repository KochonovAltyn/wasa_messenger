# 💬 WASA - WhatsApp Clone

A full-stack WhatsApp Web clone built with Go backend and Vue.js frontend.

![WhatsApp Interface](https://img.shields.io/badge/Interface-WhatsApp-25D366?style=for-the-badge&logo=whatsapp)
![Go](https://img.shields.io/badge/Go-1.19-00ADD8?style=for-the-badge&logo=go)
![Vue.js](https://img.shields.io/badge/Vue.js-3.2-4FC08D?style=for-the-badge&logo=vue.js)
![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker)

---

## 🚀 Features

### ✨ Frontend (Vue.js)
- 🎨 **Pixel-perfect WhatsApp UI** - Dark theme with authentic colors
- 💬 **Real-time messaging** - Send and receive messages
- 👥 **Group chats** - Create and manage group conversations
- 🔍 **Search functionality** - Search messages and contacts
- 😊 **Emoji support** - Full emoji picker
- 📎 **File attachments** - Attach files to messages
- 🔔 **Browser notifications** - Desktop notifications for new messages
- 📱 **Responsive design** - Works on mobile and desktop
- 🌐 **Multi-language** - English interface

### 🔧 Backend (Go)
- 🗄️ **SQLite database** - Lightweight data storage
- 🔐 **Authentication** - User login/register with tokens
- 📨 **REST API** - RESTful endpoints for all operations
- 👤 **User management** - Profile, username, photos
- 💬 **Conversations** - Private and group chats
- 📝 **Messages** - Send, forward, delete, comment
- 👥 **Groups** - Create groups, add/remove members

---

## 🛠️ Tech Stack

### Backend
- **Language:** Go 1.19
- **Database:** SQLite
- **Framework:** Standard library + custom router
- **API:** RESTful

### Frontend
- **Framework:** Vue.js 3.2
- **Build Tool:** Vite
- **HTTP Client:** Axios
- **Router:** Vue Router 4
- **Styling:** Scoped CSS (WhatsApp theme)

### DevOps
- **Containerization:** Docker + Docker Compose
- **Package Manager:** npm
- **Version Control:** Git

---

## 📦 Project Structure

```
wasa_project/
├── cmd/
│   ├── healthcheck/          # Health check utility
│   └── webapi/               # Main API server
│       ├── main.go
│       ├── cors.go
│       ├── register-web-ui.go
│       └── load-configuration.go
├── service/
│   ├── api/                  # API handlers
│   │   ├── api.go
│   │   ├── users.go
│   │   ├── conversations.go
│   │   └── ...
│   └── database/             # Database layer
│       ├── database.go
│       ├── users.go
│       ├── conversations.go
│       └── schemas.go
├── webui/                    # Frontend application
│   ├── src/
│   │   ├── components/       # Vue components
│   │   │   └── WhatsAppView.vue
│   │   ├── views/            # Page views
│   │   │   └── HomeView.vue
│   │   ├── router/           # Vue Router
│   │   ├── services/         # API services
│   │   │   └── axios.js
│   │   ├── App.vue
│   │   └── main.js
│   ├── public/
│   ├── package.json
│   └── vite.config.js
├── Dockerfile.backend        # Backend Docker image
├── Dockerfile.frontend       # Frontend Docker image
├── docker-compose.yml        # Docker Compose config
├── go.mod                    # Go dependencies
└── README.md                 # This file
```

---

## 🚀 Quick Start

### Prerequisites

- **Go 1.19+** - [Download](https://golang.org/dl/)
- **Node.js 20+** - [Download](https://nodejs.org/)
- **npm** - Comes with Node.js
- **Docker** (optional) - [Download](https://www.docker.com/)

### 🏃 Running Locally

#### 1. Clone the repository

```bash
git clone https://github.com/KochonovAltyn/wasa_messenger.git
cd wasa_project
```

#### 2. Start Backend

```bash
# Install dependencies
go mod download

# Run backend server
go run ./cmd/webapi

# Backend will start on http://localhost:3000
```

#### 3. Start Frontend (in a new terminal)

```bash
cd webui

# Install dependencies
npm install

# Run development server
npm run dev

# Frontend will start on http://localhost:5173
```

#### 4. Open in browser

Navigate to: **http://localhost:5173**

---

## 🐳 Running with Docker

### Option 1: Docker Compose (Recommended)

```bash
# Build and start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

**Access:**
- Frontend: http://localhost:8080
- Backend API: http://localhost:3000

### Option 2: Manual Docker Commands

```bash
# Build images
docker build -f Dockerfile.backend -t wasa-backend .
docker build -f Dockerfile.frontend -t wasa-frontend .

# Run containers
docker run -d --name wasa-backend -p 3000:3000 wasa-backend
docker run -d --name wasa-frontend -p 8080:80 wasa-frontend

# Check status
docker ps
```

---

## 📡 API Endpoints

### Authentication
- `POST /session` - Login/Register

### Users
- `GET /users/{id}` - Get user details
- `PUT /users/me/username` - Update username
- `PUT /users/me/photo` - Update profile photo

### Conversations
- `GET /users/{id}/conversations` - List all chats
- `GET /conversations/{c_id}` - Get specific conversation
- `POST /conversations/{c_id}/messages` - Send message
- `POST /users/{id}/conversations/first-message` - Start new chat

### Messages
- `POST /conversations/{c_id}/messages/{id}/forward/{target_id}` - Forward message
- `POST /conversations/{c_id}/messages/{id}/comments` - Add comment
- `DELETE /conversations/{c_id}/messages/{id}` - Delete message

### Groups
- `POST /groups` - Create group
- `POST /groups/{c_id}/members` - Add member
- `DELETE /groups/{c_id}/leave` - Leave group
- `PUT /groups/{c_id}/name` - Update group name
- `PUT /conversations/{c_id}/set-group-photo` - Update group photo

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

---
**⭐ If you like this project, please give it a star!**


