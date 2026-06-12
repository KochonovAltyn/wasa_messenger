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

// MarkMessagesAsRead records that readerID has read every message in the
// conversation that was not sent by them. It is called when a user opens a
// conversation. Read receipts are tracked PER USER: the sender's double
// checkmark lights up only once every other member has a row here (see the
// computed status in GetMessagesByConversationId).
func (db *appdbimpl) MarkMessagesAsRead(conversationID int, readerID string) error {
	query := `
		INSERT INTO message_reads (message_id, user_id)
		SELECT m.id, ?
		FROM messages m
		WHERE m.conversation_id = ?
		  AND m.sender <> ?
		ON CONFLICT(message_id, user_id) DO NOTHING;`
	_, err := db.c.Exec(query, readerID, conversationID, readerID)
	return err
}
