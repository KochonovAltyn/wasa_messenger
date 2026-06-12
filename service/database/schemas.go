package database

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
	ID        int            `json:"id"`
	LastConvo time.Time      `json:"last_convo"`
	IsGroup   bool           `json:"is_group"`
	Photo     sql.NullString `json:"photo"`
	Name      string         `json:"name"`
	// NEW FIELDS:
	LastMessage     sql.NullString `json:"last_message"`
	LastMessageType sql.NullString `json:"last_message_type"`
}

type Convmember struct {
	ID             int `json:"id"`
	ConversationID int `json:"conversation_id"`
	UserID         int `json:"user_id"`
}

type Message struct {
	ID             int           `json:"id"`
	Datetime       time.Time     `json:"datetime"`
	Content        string        `json:"content"`
	Sender         int           `json:"sender"`
	ConversationID int           `json:"conversation_id"`
	Status         string        `json:"status"`
	ReplyTo        sql.NullInt64 `json:"reply_to"`
}

type ConversationInfo struct {
	OtherUserName string    `json:"other_user_name"` // Name of the other user in the conversation
	LastConvo     time.Time `json:"last_convo"`      // Timestamp of the last conversation
}

// type Participant struct {
//     ID       string `json:"id"`
//     Username string `json:"username"`
//     Photo    string `json:"photo"`
// }

type MessageWithSender struct {
	ID             int            `json:"id"`
	Datetime       time.Time      `json:"datetime"`
	Content        string         `json:"content"`
	Caption        string         `json:"caption"` // optional text attached to a media message
	Status         string         `json:"status"`
	SenderID       string         `json:"sender_id"`
	SenderUsername string         `json:"sender_username"`
	SenderPhoto    sql.NullString `json:"sender_photo"` // Use sql.NullString for nullable fields
	// The ID of the message this one is replying to (if any).
	ReplyTo sql.NullInt64 `json:"reply_to,omitempty"`

	// Optionally, show snippet from original message (if reply_to is valid).
	ReplyToContent        sql.NullString `json:"reply_to_content,omitempty"`
	ReplyToSenderUsername sql.NullString `json:"reply_to_sender,omitempty"`

	// Emoji reactions placed on this message by users in the conversation.
	Reactions []Reaction `json:"reactions"`
}

type MessageComment struct {
	ID        int       `json:"id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}
