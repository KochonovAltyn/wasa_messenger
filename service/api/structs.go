package api

import (
	"database/sql"
	"time"
)

type User struct {
	ID       string         `json:"id"`
	Username string         `json:"username"`
	Photo    sql.NullString `json:"photo"`
}

type Conversation struct {
	ID        int       `json:"id"`
	LastConvo time.Time `json:"last_convo"`
	IsGroup   bool      `json:"is_group"`
	Photo     string    `json:"photo"`
}

type Convmember struct {
	ID             int `json:"id"`
	ConversationID int `json:"conversation_id"`
	UserID         int `json:"user_id"`
}

type Message struct {
	ID             int       `json:"id"`
	Datetime       time.Time `json:"datetime"`
	Content        string    `json:"content"`
	Sender         int       `json:"sender"`
	ConversationID int       `json:"conversation_id"`
	Status         string    `json:"status"`
}

type Participant struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Photo    string `json:"photo"`
}


// узнать Это файл с моделями (структурами данных) твоего приложения. Он нужен, чтобы описать “какие сущности существуют” (User, Conversation, Message…) и в каком виде они:
// хранятся/читаются из базы данных отдаются в JSON через API передаются между функциями внутри кода