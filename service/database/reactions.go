package database

// Reaction represents a single emoji reaction placed by a user on a message.
type Reaction struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Emoji    string `json:"emoji"`
}

// SetReaction adds or updates the emoji reaction of a user on a message.
// Because of the UNIQUE(message_id, user_id) constraint, a user can only have
// one reaction per message: reacting again replaces the previous emoji.
func (db *appdbimpl) SetReaction(messageID int, userID string, emoji string) error {
	query := `
		INSERT INTO reactions (message_id, user_id, emoji)
		VALUES (?, ?, ?)
		ON CONFLICT(message_id, user_id)
		DO UPDATE SET emoji = excluded.emoji, timestamp = CURRENT_TIMESTAMP;`
	_, err := db.c.Exec(query, messageID, userID, emoji)
	return err
}

// RemoveReaction deletes the reaction a user placed on a message (if any).
func (db *appdbimpl) RemoveReaction(messageID int, userID string) error {
	query := `DELETE FROM reactions WHERE message_id = ? AND user_id = ?;`
	_, err := db.c.Exec(query, messageID, userID)
	return err
}

// GetReactionsByMessageID returns all reactions on a message, including the
// username of each reacting user so the frontend can show who reacted.
func (db *appdbimpl) GetReactionsByMessageID(messageID int) ([]Reaction, error) {
	query := `
		SELECT r.user_id, u.name, r.emoji
		FROM reactions r
		JOIN users u ON r.user_id = u.id
		WHERE r.message_id = ?
		ORDER BY r.timestamp ASC;`
	rows, err := db.c.Query(query, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reactions []Reaction
	for rows.Next() {
		var r Reaction
		if err := rows.Scan(&r.UserID, &r.Username, &r.Emoji); err != nil {
			return nil, err
		}
		reactions = append(reactions, r)
	}
	return reactions, rows.Err()
}

// MarkMessagesAsRead marks every message in the given conversation that was NOT
// sent by readerID as 'read'. It is called when a user opens a conversation, so
// that the sender's "read" checkmark lights up. Messages already 'read' are
// left untouched.
func (db *appdbimpl) MarkMessagesAsRead(conversationID int, readerID string) error {
	query := `
		UPDATE messages
		SET status = 'read'
		WHERE conversation_id = ?
		  AND sender <> ?
		  AND status <> 'read';`
	_, err := db.c.Exec(query, conversationID, readerID)
	return err
}
