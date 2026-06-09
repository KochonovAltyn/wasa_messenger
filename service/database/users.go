package database

import (
	"database/sql"
	"errors"
	"fmt"

	uuid "github.com/KochonovAltyn/wasa_messenger/service/uid"
)

// func (db *appdbimpl) CreateUser(username string) (user User, err error) {

// 	query := `INSERT INTO users ( name ) VALUES (?);`
// 	_, err = db.c.Exec(query, username)
// 	if err != nil {
// 		return
// 	}
// 	return db.GetUser(username)
// }

func (db *appdbimpl) CreateUser(username string) (User, error) {
	// Start a transaction
	tx, err := db.c.Begin()
	if err != nil {
		return User{}, err
	}
	// defer tx.Rollback() // Rollback the transaction if it's not committed
	defer func() {
		if newerr := tx.Rollback(); newerr != nil {
			err = errors.New("failed to rollback transaction")
		}
	}()

	// Generate a UUID
	id, err := uuid.NewV4()
	if err != nil {
		return User{}, err
	}

	// Insert the new user
	query := `INSERT INTO users (id, name) VALUES (?, ?);`
	_, err = tx.Exec(query, id.String(), username)
	if err != nil {
		return User{}, err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return User{}, err
	}

	// Retrieve the created user
	return db.GetUser(username)
}

func (db *appdbimpl) GetUser(username string) (user User, err error) {
	query := `SELECT id, name FROM users WHERE name = ?;`
	row := db.c.QueryRow(query, username)

	// Attempt to scan the result into the User struct
	err = row.Scan(&user.ID, &user.Username)
	if errors.Is(err, sql.ErrNoRows) {
		// Return a clear error if no user is found
		return
	} else if err != nil {
		// Return the database error for debugging
		return
	}

	// Successfully found the user
	return user, nil
}


func (db *appdbimpl) GetUserId(id string) (user User, err error) {
	// Query the database for the user by ID (UUID)
	query := `SELECT id, name, photo FROM users WHERE id = ?;`
	row := db.c.QueryRow(query, id)

	// Attempt to scan the result into the User struct
	err = row.Scan(&user.ID, &user.Username, &user.Photo)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Return a clear error if no user is found
			return User{}, sql.ErrNoRows
		}
		// Return other database-related errors
		return User{}, err
	}

	// Successfully found the user
	return user, nil
}

// -------------
func (db *appdbimpl) UpdateUserName(id string, newname string) (err error) {
	tx, err := db.c.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if newerr := tx.Rollback(); newerr != nil && !errors.Is(newerr, sql.ErrTxDone) {
			err = fmt.Errorf("failed to rollback transaction: %w", newerr)
		}
	}()

	// Check if the username already exists
	var count int
	checkQuery := `SELECT COUNT(*) FROM users WHERE name = ? AND id <> ?;`
	err = tx.QueryRow(checkQuery, newname, id).Scan(&count)
	if err != nil {
		return errors.New("failed to check if username exists: " + err.Error())
	}
	if count > 0 {
		return errors.New("username '" + newname + "' is already taken by another user")
	}

	// Update the username in the users table
	updateUserQuery := `UPDATE users SET name = ? WHERE id = ?;`
	_, err = tx.Exec(updateUserQuery, newname, id)
	if err != nil {
		return fmt.Errorf("failed to update username in users table: %w", err)
	}

	// Update conversation names only for one-on-one conversations (is_group = FALSE)
	updateConversationQuery := `
		UPDATE conversations 
		SET name = ? 
		WHERE is_group = FALSE 
		AND id IN (
			SELECT conversation_id FROM convmembers WHERE user_id = ?
		);
	`
	_, err = tx.Exec(updateConversationQuery, newname, id)
	if err != nil {
		return fmt.Errorf("failed to update conversation names: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (db *appdbimpl) UpdateUserPhoto(userID string, photoPath string) error {
	query := `UPDATE users SET photo = ? WHERE id = ?;`
	_, err := db.c.Exec(query, photoPath, userID)
	return err
}

func GetUserByID(db *sql.DB, userID string) (*User, error) {
	var user User

	err := db.QueryRow("SELECT id, name, photo FROM users WHERE id = ?", userID).
		Scan(&user.ID, &user.Username, &user.Photo)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

func (db *appdbimpl) GetUserByID(userID string) (User, error) {
	query := `SELECT id, name, photo FROM users WHERE id = ?;`
	row := db.c.QueryRow(query, userID)

	var user User
	err := row.Scan(&user.ID, &user.Username, &user.Photo)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, sql.ErrNoRows
		}
		return User{}, err
	}

	return user, nil
}

// GetUserIDByUsername resolves the username to a user ID
// GetUserIDByUsername fetches the user ID by their username
func (db *appdbimpl) GetUserIDByUsername(username string) (string, error) {
	query := `SELECT id FROM users WHERE name = ?`
	var userID string
	err := db.c.QueryRow(query, username).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("user not found")
		}
		return "", err
	}
	return userID, nil
}
