/*
Package database is the middleware between the app database and the code. All data (de)serialization (save/load) from a
persistent database are handled here. Database specific logic should never escape this package.

To use this package you need to apply migrations to the database if needed/wanted, connect to it (using the database
data source name from config), and then initialize an instance of AppDatabase from the DB connection.

For example, this code adds a parameter in `webapi` executable for the database data source name (add it to the
main.WebAPIConfiguration structure):

	DB struct {
		Filename string `conf:""`
	}

This is an example on how to migrate the DB and connect to it:

	// Start Database
	logger.Println("initializing database support")
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		logger.WithError(err).Error("error opening SQLite DB")
		return fmt.Errorf("opening SQLite: %w", err)
	}
	defer func() {
		logger.Debug("database stopping")
		_ = db.Close()
	}()

Then you can initialize the AppDatabase and pass it to the api package.
*/
package database

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"strings"
)

// AppDatabase is the high level interface for the DB
// all the function that I creat in db nust be declated here
type AppDatabase interface {
	// User-related methods
	CreateUser(username string) (User, error)
	GetUser(username string) (User, error)
	GetUserId(id string) (User, error) // Fetch a user by ID (UUID)
	UpdateUserPhoto(userID string, filePath string) error
	GetUserByID(userID string) (User, error) // ✅ Add this function
	GetUserIDByUsername(username string) (string, error)

	// Conversation-related methods
	// GetMyConversations_db(userID string) (conversations []Conversation, err error)
	CreateConversation_db(isGroup bool, name string, photo string) (conversation Conversation, err error)
	AddUsersToConversation(userID string, conversationID int) (err error)
	GetConversationById(conversationID int) (conversation Conversation, err error)
	ConversationExists(senderID string, recipientID string) (bool, error)
	GetMyConversations_db(userID string) ([]Conversation, error)
	RemoveUserFromGroup(userID string, groupID int) error
	GetGroupMemberCount(groupID int) (int, error)
	DeleteGroup(groupID int) error
	IsConversationGroup(conversationID int) (bool, error)
	UpdateGroupName(groupID int, newName string) error
	GroupNameExists(name string) (bool, error)
	DoesConversationExist(conversationID int) (bool, error)
	// Message-related methods
	SendMessage(conversationID int, senderID string, content string) error
	IsUserInConversation(userID string, conversationID int) (bool, error)
	SendMessageFull(conversationID int, senderID string, content string) error
	GetMessagesByConversationId(conversationID int) ([]MessageWithSender, error)
	IsMessageOwner(userID string, messageID int) (bool, error)
	DeleteMessage(messageID int) error
	GetMessageContent(messageID int) (string, error)
	ForwardMessage(targetConversationID int, senderID string, content string) error
	UpdateGroupPhoto(groupID int, photoPath string) error
	DoesMessageExist(messageID int) (bool, error)
	CommentOnMessage(messageID int, userID string, contentType string, content string) error
	ConvertCommentsToMessages(messageID int, conversationID int) error
	IsCommentOwner(userID string, commentID int) (bool, error)
	DeleteComment(commentID int) error
	SendMessageWithType(conversationID int, senderID string, content string, contentType string, caption string, replyTo *int) error
	SendMessageWithMedia(conversationID int, senderID string, contentType string, content string, caption string) error
	SaveUploadedFile(file io.Reader, header *multipart.FileHeader, userID string) (string, string, error)
	GetCommentsByMessageID(messageID int) ([]MessageComment, error)
	GetConversationBetweenUsers(user1 string, user2 string) (Conversation, error)
	GetGroupByName(groupName string) (Conversation, error)
	GetGroupNameById(conversationID int) (string, error)
	MarkMessagesAsRead(conversationID int, readerID string) error

	// Reaction-related methods
	SetReaction(messageID int, userID string, emoji string) error
	RemoveReaction(messageID int, userID string) error
	GetReactionsByMessageID(messageID int) ([]Reaction, error)

	// User updates
	UpdateUserName(id string, newname string) (err error)

	// Connection health
	Ping() error
}
type appdbimpl struct {
	c *sql.DB
}

// New returns a new instance of AppDatabase based on the SQLite connection `db`.
// `db` is required - an error will be returned if `db` is `nil`.
func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building a AppDatabase")
	}

	// Check if table exists. If not, the database is empty, and we need to create the structure
	var tableName string
	err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='users';`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		err = createDatabase(db)
		if err != nil {
			return nil, fmt.Errorf("error creating database structure: %w", err)
		}
	}

	// Lightweight migrations for databases created before these features:
	// per-user read receipts and image captions.
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS message_reads (
		message_id INTEGER NOT NULL,
		user_id VARCHAR(64) NOT NULL,
		read_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(message_id, user_id),
		FOREIGN KEY(message_id) REFERENCES messages(id),
		FOREIGN KEY(user_id) REFERENCES users(id)
	);`); err != nil {
		return nil, fmt.Errorf("error creating message_reads table: %w", err)
	}
	if _, err := db.Exec(`ALTER TABLE messages ADD COLUMN caption TEXT DEFAULT '';`); err != nil &&
		!strings.Contains(err.Error(), "duplicate column") {
		return nil, fmt.Errorf("error adding caption column: %w", err)
	}

	return &appdbimpl{
		c: db,
	}, nil
}

func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}

func createDatabase(db *sql.DB) error {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS users(
			id VARCHAR(64), 
			name VARCHAR(25) NOT NULL,
			photo VARCHAR(255)
		);`,

		`CREATE TABLE IF NOT EXISTS conversations(
			id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, 
			lastconvo TIMESTAMP NOT NULL,
			is_group BOOLEAN DEFAULT FALSE,
			photo VARCHAR(255),
			name VARCHAR(255)


		);`,

		`CREATE TABLE IF NOT EXISTS convmembers (
			id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, 
			conversation_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			FOREIGN KEY (conversation_id) REFERENCES conversations (id),
			FOREIGN KEY (user_id) REFERENCES users (id)
		);`,

		`CREATE TABLE IF NOT EXISTS messages(
			id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
			datetime TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			content TEXT NOT NULL,
			content_type TEXT DEFAULT 'text',  -- ✅ New column to store message type
			caption TEXT DEFAULT '',           -- optional text sent together with a media message
			sender INTEGER NOT NULL,
			conversation_id INTEGER NOT NULL,
			status VARCHAR(10) DEFAULT 'delivered',
			reply_to INTEGER DEFAULT NULL,
			FOREIGN KEY(sender) REFERENCES users(id),
			FOREIGN KEY(conversation_id) REFERENCES conversations(id)
			FOREIGN KEY(reply_to) REFERENCES messages(id)
		);`,
		`CREATE TABLE IF NOT EXISTS message_comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
			message_id INTEGER NOT NULL,
			user_id VARCHAR(64) NOT NULL,
			content_type VARCHAR(10) CHECK (content_type IN ('text', 'emoji', 'photo', 'gif')) NOT NULL,
			content TEXT NOT NULL,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(message_id) REFERENCES messages(id),
			FOREIGN KEY(user_id) REFERENCES users(id)
		);`,
		`CREATE TABLE IF NOT EXISTS reactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
			message_id INTEGER NOT NULL,
			user_id VARCHAR(64) NOT NULL,
			emoji VARCHAR(16) NOT NULL,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(message_id, user_id),
			FOREIGN KEY(message_id) REFERENCES messages(id),
			FOREIGN KEY(user_id) REFERENCES users(id)
		);`,

		`CREATE TABLE IF NOT EXISTS message_reads (
			message_id INTEGER NOT NULL,
			user_id VARCHAR(64) NOT NULL,
			read_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(message_id, user_id),
			FOREIGN KEY(message_id) REFERENCES messages(id),
			FOREIGN KEY(user_id) REFERENCES users(id)
		);`,
	}
	for t := 0; t < len(tables); t++ {
		sqlStmt := tables[t]
		_, err := db.Exec(sqlStmt)
		if err != nil {
			return err
		}
	}
	return nil
}