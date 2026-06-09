package database

import (
	"database/sql"
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)




func (db *appdbimpl) GetConversationById(conversationID int) (conversation Conversation, err error) {
	query := `
		SELECT id, lastconvo, is_group, photo, name
		FROM conversations
		WHERE id = ?;
	`

	err = db.c.QueryRow(query, conversationID).Scan(
		&conversation.ID,
		&conversation.LastConvo,
		&conversation.IsGroup,
		&conversation.Photo,
		&conversation.Name,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Conversation{}, nil // Return an empty conversation if no rows are found
		}
		return Conversation{}, err
	}

	return conversation, nil
}

// -------Messages-----

func (db *appdbimpl) CreateConversation_db(isGroup bool, name string, photo string) (conversation Conversation, err error) {
	var photoURL sql.NullString
	if photo != "" {
		photoURL = sql.NullString{String: photo, Valid: true} // Set the photo URL correctly
	} else {
		photoURL = sql.NullString{Valid: false} // If there's no photo, set it as NULL
	}
	query := `
		INSERT INTO conversations (lastconvo, is_group, name, photo)
		VALUES (current_timestamp, ?, ?, ?)
		RETURNING id, lastconvo, is_group, name, photo;
	`
	err = db.c.QueryRow(query, isGroup, name, photoURL).Scan(
		&conversation.ID,
		&conversation.LastConvo,
		&conversation.IsGroup,
		&conversation.Name,
		&conversation.Photo,
	)
	return
}
func (db *appdbimpl) AddUsersToConversation(userID string, conversationID int) (err error) {
	query := `
		INSERT INTO convmembers (user_id, conversation_id)
		VALUES (?, ?);
	`
	_, err = db.c.Exec(query, userID, conversationID)
	return
}

// func (db *appdbimpl) SendMessage(conversationID int, senderID string, content string) error {
// 	query := `
// 		INSERT INTO messages (conversation_id, sender, content, datetime, status)
// 		VALUES (?, ?, ?, CURRENT_TIMESTAMP, 'sent');
// 	`
// 	_, err := db.c.Exec(query, conversationID, senderID, content)
// 	return err
// }

// Check if a conversation already exists between two users
func (db *appdbimpl) ConversationExists(senderID string, recipientID string) (bool, error) {
	query := `
        SELECT COUNT(*) 
        FROM convmembers cm1
        JOIN convmembers cm2 ON cm1.conversation_id = cm2.conversation_id
        JOIN conversations c ON cm1.conversation_id = c.id
        WHERE cm1.user_id = ? 
        AND cm2.user_id = ? 
        AND c.is_group = FALSE;
    `

	var count int
	err := db.c.QueryRow(query, senderID, recipientID).Scan(&count)
	if err != nil {
		return false, err // Returning the error, so it gets logged in the caller function
	}

	return count > 0, nil
}

// GetMyConversations_db retrieves all conversations for a specific user.
func (db *appdbimpl) GetMyConversations_db(userID string) ([]Conversation, error) {
	query := `
        SELECT 
            c.id, 
            COALESCE(
                (SELECT MAX(m.datetime) 
                 FROM messages m 
                 WHERE m.conversation_id = c.id),
                c.lastconvo
            ) AS lastconvo,
            c.is_group, 
            c.photo,
            CASE 
                WHEN c.is_group = TRUE THEN c.name 
                ELSE (SELECT u.name FROM users u 
                      JOIN convmembers cm ON u.id = cm.user_id 
                      WHERE cm.conversation_id = c.id AND u.id != ? LIMIT 1)
            END AS name,
            CASE
                WHEN c.is_group = FALSE THEN (SELECT u.photo FROM users u 
                                              JOIN convmembers cm ON u.id = cm.user_id 
                                              WHERE cm.conversation_id = c.id AND u.id != ? LIMIT 1)
                ELSE NULL
            END AS user_photo,
            (SELECT m.content FROM messages m WHERE m.conversation_id = c.id ORDER BY m.datetime DESC LIMIT 1) AS last_message,
            (SELECT m.content_type FROM messages m WHERE m.conversation_id = c.id ORDER BY m.datetime DESC LIMIT 1) AS last_message_type
        FROM 
            conversations c
        JOIN 
            convmembers cm ON c.id = cm.conversation_id
        WHERE 
            cm.user_id = ?;
    `

	// Pass userID three times: first two for the subqueries and the third for the WHERE clause.
	rows, err := db.c.Query(query, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []Conversation
	for rows.Next() {
		var convo Conversation
		var lastConvoStr string // temporary variable for the timestamp as string
		var userPhoto sql.NullString
		var lastMessage sql.NullString
		var lastMessageType sql.NullString

		err := rows.Scan(&convo.ID, &lastConvoStr, &convo.IsGroup, &convo.Photo, &convo.Name, &userPhoto, &lastMessage, &lastMessageType)
		if err != nil {
			return nil, err
		}

		// Parse the last conversation timestamp string into a time.Time.
		// Adjust the layout if your database returns a different format.
		parsedTime, err := time.Parse("2006-01-02 15:04:05", lastConvoStr)
		if err != nil {
			return nil, err
		}
		convo.LastConvo = parsedTime

		// Correct the photo URL by avoiding double '/uploads/'
		if userPhoto.Valid && userPhoto.String != "" {
			if !strings.HasPrefix(userPhoto.String, "/uploads/") {
				convo.Photo.String = "/uploads/" + userPhoto.String
			} else {
				convo.Photo.String = userPhoto.String
			}
		} else if convo.Photo.Valid && convo.Photo.String != "" {
			if !strings.HasPrefix(convo.Photo.String, "/uploads/") {
				convo.Photo.String = "/uploads/" + convo.Photo.String
			}
		} else {
			convo.Photo.String = "/default-profile.png"
		}

		// Set the new last message fields.
		convo.LastMessage = lastMessage
		convo.LastMessageType = lastMessageType

		conversations = append(conversations, convo)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return conversations, nil
}

// SendMessage inserts a new message into the database.
func (db *appdbimpl) SendMessage(conversationID int, senderID string, content string) error {
	query := `
        INSERT INTO messages (conversation_id, sender, content, datetime, status)
        VALUES (?, ?, ?, CURRENT_TIMESTAMP, 'delivered');
    `
	_, err := db.c.Exec(query, conversationID, senderID, content)
	return err
}

func (db *appdbimpl) IsUserInConversation(userID string, conversationID int) (bool, error) {
	query := `
        SELECT COUNT(*) 
        FROM convmembers 
        WHERE user_id = ? AND conversation_id = ?;
    `
	var count int
	err := db.c.QueryRow(query, userID, conversationID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
func (db *appdbimpl) SendMessageFull(conversationID int, senderID string, content string) error {
	query := `
        INSERT INTO messages (conversation_id, sender, content, datetime, status)
        VALUES (?, ?, ?, CURRENT_TIMESTAMP, 'delivered');
    `
	_, err := db.c.Exec(query, conversationID, senderID, content)
	return err
}

// GetMessagesByConversationId: now does a LEFT JOIN on the "parent" message
func (db *appdbimpl) GetMessagesByConversationId(conversationID int) ([]MessageWithSender, error) {
	query := `
	SELECT 
	  m.id,
	  m.datetime,
	  m.content,
	  m.status,
	  u.id         AS sender_id,
	  u.name       AS sender_username,
	  u.photo      AS sender_photo,
	
	  m.reply_to,
	  pm.content   AS reply_to_content,
	  pu.name      AS reply_to_sender_username
	
	FROM messages m
	JOIN users u ON m.sender = u.id
	
	LEFT JOIN messages pm ON m.reply_to = pm.id
	LEFT JOIN users pu    ON pm.sender = pu.id
	
	WHERE m.conversation_id = ?
	ORDER BY m.datetime ASC;
    `

	rows, err := db.c.Query(query, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []MessageWithSender
	for rows.Next() {
		var msg MessageWithSender

		// We scan the joined columns
		err := rows.Scan(
			&msg.ID,
			&msg.Datetime,
			&msg.Content,
			&msg.Status,
			&msg.SenderID,
			&msg.SenderUsername,
			&msg.SenderPhoto,

			// new columns for the reply
			&msg.ReplyTo,
			&msg.ReplyToContent,
			&msg.ReplyToSenderUsername,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (db *appdbimpl) IsMessageOwner(userID string, messageID int) (bool, error) {
	query := `
        SELECT COUNT(*) 
        FROM messages 
        WHERE id = ? AND sender = ?;
    `
	var count int
	err := db.c.QueryRow(query, messageID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (db *appdbimpl) DeleteMessage(messageID int) error {
	query := `DELETE FROM messages WHERE id = ?;`
	_, err := db.c.Exec(query, messageID)
	return err
}

func (db *appdbimpl) GetMessageContent(messageID int) (string, error) {
	query := `
        SELECT content 
        FROM messages 
        WHERE id = ?;
    `
	var content string
	err := db.c.QueryRow(query, messageID).Scan(&content)
	if err != nil {
		return "", err
	}
	return content, nil
}

func (db *appdbimpl) ForwardMessage(targetConversationID int, senderID string, content string) error {
	query := `
        INSERT INTO messages (conversation_id, sender, content, datetime, status)
        VALUES (?, ?, ?, CURRENT_TIMESTAMP, 'forwarded');
    `
	_, err := db.c.Exec(query, targetConversationID, senderID, content)
	return err
}

// ✅ Remove a user from a group
func (db *appdbimpl) RemoveUserFromGroup(userID string, groupID int) error {
	query := `DELETE FROM convmembers WHERE user_id = ? AND conversation_id = ?;`
	_, err := db.c.Exec(query, userID, groupID)
	return err
}

// ✅ Get the count of remaining members in a group
func (db *appdbimpl) GetGroupMemberCount(groupID int) (int, error) {
	query := `SELECT COUNT(*) FROM convmembers WHERE conversation_id = ?;`
	var count int
	err := db.c.QueryRow(query, groupID).Scan(&count)
	return count, err
}

// ✅ Delete a group if empty
func (db *appdbimpl) DeleteGroup(groupID int) error {
	query := `DELETE FROM conversations WHERE id = ?;`
	_, err := db.c.Exec(query, groupID)
	return err
}

// ✅ Check if a conversation is a group
func (db *appdbimpl) IsConversationGroup(conversationID int) (bool, error) {
	query := `SELECT is_group FROM conversations WHERE id = ?;`
	var isGroup bool
	err := db.c.QueryRow(query, conversationID).Scan(&isGroup)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil // No conversation found
		}
		return false, err
	}
	return isGroup, nil
}

// ✅ Check if a group with the given name already exists
func (db *appdbimpl) GroupNameExists(name string) (bool, error) {
	query := `SELECT COUNT(*) FROM conversations WHERE is_group = TRUE AND name = ?;`
	var count int
	err := db.c.QueryRow(query, name).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ✅ Update the name of a group conversation
func (db *appdbimpl) UpdateGroupName(groupID int, newName string) error {
	query := `UPDATE conversations SET name = ? WHERE id = ? AND is_group = TRUE;`
	_, err := db.c.Exec(query, newName, groupID)
	return err
}

// ✅ Update the group photo
func (db *appdbimpl) UpdateGroupPhoto(groupID int, photoPath string) error {
	query := `UPDATE conversations SET photo = ? WHERE id = ? AND is_group = TRUE;`
	_, err := db.c.Exec(query, photoPath, groupID)
	return err
}

// Check if a message exists
func (db *appdbimpl) DoesMessageExist(messageID int) (bool, error) {
	query := `SELECT COUNT(*) FROM messages WHERE id = ?;`
	var count int
	err := db.c.QueryRow(query, messageID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Add a comment to a message
func (db *appdbimpl) CommentOnMessage(messageID int, userID string, contentType string, content string) error {
	query := `
        INSERT INTO message_comments (message_id, user_id, content_type, content, timestamp)
        VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP);
    `
	_, err := db.c.Exec(query, messageID, userID, contentType, content)
	return err
}

func (db *appdbimpl) DoesConversationExist(conversationID int) (bool, error) {
	query := `SELECT COUNT(*) FROM conversations WHERE id = ?;`
	var count int
	err := db.c.QueryRow(query, conversationID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (db *appdbimpl) ConvertCommentsToMessages(messageID int, conversationID int) error {
	querySelect := `
        SELECT user_id, content_type, content FROM message_comments WHERE message_id = ?;
    `

	// Fetch comments first
	rows, err := db.c.Query(querySelect, messageID)
	if err != nil {
		return err
	}

	// Store comments in memory before closing rows
	var comments []struct {
		UserID      string
		ContentType string
		Content     string
	}

	for rows.Next() {
		var comment struct {
			UserID      string
			ContentType string
			Content     string
		}
		if err := rows.Scan(&comment.UserID, &comment.ContentType, &comment.Content); err != nil {
			rows.Close() // Ensure we close before returning an error
			return err
		}
		comments = append(comments, comment)
	}
	if err = rows.Err(); err != nil {
		return err
	}

	rows.Close() // ✅ Properly close before inserting new messages

	// Insert each comment as a new message
	for _, comment := range comments {
		queryInsert := `
            INSERT INTO messages (conversation_id, sender, content, datetime, status)
            VALUES (?, ?, ?, CURRENT_TIMESTAMP, 'comment-converted');
        `
		_, err = db.c.Exec(queryInsert, conversationID, comment.UserID, comment.Content)
		if err != nil {
			return err
		}
	}

	// Finally, delete comments from message_comments
	queryDelete := `DELETE FROM message_comments WHERE message_id = ?;`
	_, err = db.c.Exec(queryDelete, messageID)

	return err
}

// ✅ Check if a user is the owner of a comment
func (db *appdbimpl) IsCommentOwner(userID string, commentID int) (bool, error) {
	query := `SELECT COUNT(*) FROM message_comments WHERE id = ? AND user_id = ?;`
	var count int
	err := db.c.QueryRow(query, commentID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ✅ Delete a comment
func (db *appdbimpl) DeleteComment(commentID int) error {
	query := `DELETE FROM message_comments WHERE id = ?;`
	_, err := db.c.Exec(query, commentID)
	return err
}

// SendMessageWithType is extended to accept an optional replyTo parameter
func (db *appdbimpl) SendMessageWithType(
	conversationID int,
	senderID string,
	content string,
	contentType string,
	replyTo *int,
) error {
	// We'll include reply_to in the INSERT
	query := `
        INSERT INTO messages (
            conversation_id,
            sender,
            content,
            content_type,
            datetime,
            status,
            reply_to
        )
        VALUES (
            ?,
            ?,
            ?,
            ?,
            CURRENT_TIMESTAMP,
            'delivered',
            ?
        );
    `
	// If replyTo is nil, pass NULL; otherwise pass the integer value
	var replyToParam interface{}
	if replyTo == nil {
		replyToParam = nil
	} else {
		replyToParam = *replyTo
	}

	_, err := db.c.Exec(query, conversationID, senderID, content, contentType, replyToParam)
	return err
}

func (db *appdbimpl) SendMessageWithMedia(conversationID int, senderID string, contentType string, content string) error {
	query := `
        INSERT INTO messages (conversation_id, sender, content, content_type, datetime, status)
        VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, 'delivered');
    `
	_, err := db.c.Exec(query, conversationID, senderID, content, contentType)
	return err
}

// SaveUploadedFile saves an uploaded file (photo or GIF) and returns content type & file path
func (db *appdbimpl) SaveUploadedFile(file io.Reader, header *multipart.FileHeader, userID string) (string, string, error) {
	// Allowed file types
	fileExt := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := map[string]string{
		".jpg":  "photo",
		".jpeg": "photo",
		".png":  "photo",
		".gif":  "gif",
	}

	contentType, ok := allowedExts[fileExt]
	if !ok {
		return "", "", errors.New("invalid file type")
	}

	// Ensure uploads directory exists
	uploadDir := "webui/public/uploads"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {

		return "", "", errors.New("fail to create a directory")
	}

	// Generate unique filename (userID + timestamp)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	fileName := userID + "_" + timestamp + fileExt
	filePath := filepath.Join(uploadDir, fileName)

	// Save file
	out, err := os.Create(filePath)
	if err != nil {
		return "", "", err
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		return "", "", err
	}

	// Return content type and file path
	return contentType, "/uploads/" + fileName, nil
}

func (db *appdbimpl) GetCommentsByMessageID(messageID int) ([]MessageComment, error) {
	query := `
        SELECT 
            mc.id, 
            u.id AS user_id,
            u.name, 
            mc.content, 
            mc.timestamp
        FROM 
            message_comments mc
        JOIN 
            users u ON mc.user_id = u.id
        WHERE 
            mc.message_id = ?;
    `
	rows, err := db.c.Query(query, messageID)
	if err != nil {
		// Print the error for debugging purposes
		return nil, err
	}
	defer rows.Close()

	var comments []MessageComment
	for rows.Next() {
		var comment MessageComment
		if err := rows.Scan(&comment.ID, &comment.UserID, &comment.Username, &comment.Content, &comment.Timestamp); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return comments, nil
}

func (db *appdbimpl) GetConversationBetweenUsers(user1 string, user2 string) (Conversation, error) {
	query := `
        SELECT c.id, c.lastconvo, c.is_group, c.photo, c.name
        FROM conversations c
        JOIN convmembers cm1 ON c.id = cm1.conversation_id
        JOIN convmembers cm2 ON c.id = cm2.conversation_id
        WHERE cm1.user_id = ? AND cm2.user_id = ? AND c.is_group = FALSE
        LIMIT 1;
    `
	var conv Conversation
	// Directly scan lastconvo into conv.LastConvo (of type time.Time)
	err := db.c.QueryRow(query, user1, user2).Scan(&conv.ID, &conv.LastConvo, &conv.IsGroup, &conv.Photo, &conv.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return Conversation{}, nil
		}
		return Conversation{}, err
	}
	return conv, nil
}

func (db *appdbimpl) GetGroupByName(groupName string) (Conversation, error) {
	query := `
        SELECT id, lastconvo, is_group, photo, name
        FROM conversations
        WHERE is_group = TRUE AND name = ?
        LIMIT 1;
    `
	var conv Conversation
	err := db.c.QueryRow(query, groupName).Scan(&conv.ID, &conv.LastConvo, &conv.IsGroup, &conv.Photo, &conv.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return Conversation{}, nil
		}
		return Conversation{}, err
	}
	return conv, nil
}

// GetGroupNameById returns the name of the group conversation given its id.
// It assumes the conversation exists and is a group.
func (db *appdbimpl) GetGroupNameById(conversationID int) (string, error) {
	var name string
	query := `SELECT name FROM conversations WHERE id = ? AND is_group = TRUE;`
	err := db.c.QueryRow(query, conversationID).Scan(&name)
	if err != nil {
		return "", err
	}
	return name, nil
}
