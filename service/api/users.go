package api

import (
	"encoding/json"
	"errors"

	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/KochonovAltyn/wasa_messenger/service/api/reqcontext"

	// "github.com/sirupsen/logrus"
	"database/sql"

	"io"
	"os"
	"path/filepath"
	"strings"
)

func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params, context *reqcontext.RequestContext) {
	// Parse input
	var input struct {
		Username string `json:"username"`
	}
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil || input.Username == "" {
		http.Error(w, "Invalid input: username is required", http.StatusBadRequest)
		return
	}

	// Check if the user exists
	user, err := rt.db.GetUser(input.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// User doesn't exist, create them
			user, err = rt.db.CreateUser(input.Username)
			if err != nil {
				rt.baseLogger.WithError(err).Error("Failed to create user")
				http.Error(w, "Internal server error: failed to create user", http.StatusInternalServerError)
				return
			}
		} else {
			// Log unexpected errors
			rt.baseLogger.WithError(err).Error("Unexpected error fetching user")
			http.Error(w, "Internal server error: unexpected error", http.StatusInternalServerError)
			return
		}
	}

	// Respond with the user data and user ID (token)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"user":  user,
		"token": user.ID, // The user ID is the token
	})
}

// func (rt *_router) logout(w http.ResponseWriter, r *http.Request, ps httprouter.Params, context reqcontext.RequestContext) {
// 	// Parse the request body to extract the user ID (UUID)
// 	var user User
// 	err := json.NewDecoder(r.Body).Decode(&user)
// 	if err != nil || user.ID == "" {
// 		http.Error(w, "Invalid request body or missing user ID", http.StatusBadRequest)
// 		return
// 	}

// 	// Validate if the user exists in the database
// 	_, err = rt.db.GetUserId(user.ID)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			http.Error(w, "User not found", http.StatusNotFound)
// 			return
// 		}
// 		http.Error(w, "Failed to validate user", http.StatusInternalServerError)
// 		return
// 	}

// 	// Perform any additional logout logic here (if needed)

// 	// Send a success response
// 	w.WriteHeader(http.StatusOK)
// 	w.Header().Set("Content-Type", "text/plain")
// 	_, _ = w.Write([]byte("Logout successful"))
// }

// -----------------

func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, context *reqcontext.RequestContext) {
	userID, err := rt.extractUserIDFromHeader(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var input struct {
		NewName string `json:"newname"`
	}

	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil || input.NewName == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid input: newname is required"})
		return
	}

	// Try to find any user that already has the requested new username
	existingUser, err := rt.db.GetUser(input.NewName)
	if err != nil {
		// If the DB says "no rows," that simply means no user has that name—so it's safe to proceed.
		// Otherwise, treat as an error and stop.
		if !errors.Is(err, sql.ErrNoRows) {
			context.Logger.WithError(err).Error("Error checking existing user")
			http.Error(w, "Error checking existing user", http.StatusInternalServerError)
			return
		}
	}

	// If a user *does* exist with that name, check whether it’s a *different* user.
	if existingUser.Username != "" && existingUser.ID != userID {
		// Another user already has this name => conflict.
		// Return a JSON body so the frontend can show a clear error message.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "The username '" + input.NewName + "' is already taken by another user. Please choose a different one.",
		})
		return
	}

	// Otherwise, proceed to update
	err = rt.db.UpdateUserName(userID, input.NewName)
	if err != nil {
		http.Error(w, "Failed to update username", http.StatusInternalServerError)
		return
	}

	// Success
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Username updated successfully"})
}
func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx *reqcontext.RequestContext) {
	userID := ctx.UserID
	if userID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Parse the uploaded file
	file, header, err := r.FormFile("photo")
	if err != nil {
		http.Error(w, "Invalid file upload", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate the file extension
	fileExt := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
	if !allowedExts[fileExt] {
		http.Error(w, "Invalid file type", http.StatusBadRequest)
		return
	}

	// Ensure upload directory exists
	uploadDir := "webui/public/uploads"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		ctx.Logger.WithError(err).Error("Failed to create upload directory")
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Generate the new filename using the user ID
	fileName := userID + fileExt
	filePath := filepath.Join(uploadDir, fileName)

	// Save the file
	out, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to save image", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, "Failed to save image", http.StatusInternalServerError)
		return
	}

	// Generate the correct photo URL for the frontend
	photoURL := "/uploads/" + fileName

	// Update the user's profile photo in the database
	err = rt.db.UpdateUserPhoto(userID, photoURL)
	if err != nil {
		http.Error(w, "Failed to update profile picture", http.StatusInternalServerError)
		return
	}

	// Send response with the updated photo URL
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Profile picture updated successfully", "photo": photoURL})
}

func (rt *_router) getUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params, context *reqcontext.RequestContext) {
	userID := ps.ByName("id")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	user, err := rt.db.GetUserId(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error fetching user", http.StatusInternalServerError)
		return
	}

	// ✅ Ensure `photo` is properly extracted as a string
	var photoURL string
	if user.Photo.Valid {
		photoURL = user.Photo.String
	} else {
		photoURL = "" // Default if no photo is set
	}

	response := map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"photo":    photoURL, // ✅ Now a plain string
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		context.Logger.WithError(err).Error("Error encoding response")
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}

}

// This endpoint resolves the username to user ID
// func (rt *_router) getUserIDByUsername(w http.ResponseWriter, r *http.Request, ps httprouter.Params, context *reqcontext.RequestContext) {
// 	username := ps.ByName("username")
// 	if username == "" {
// 		context.Logger.Error("Username is required")
// 		http.Error(w, "Username is required", http.StatusBadRequest)
// 		return
// 	}

// 	// Query the database to get the user ID by username
// 	userID, err := rt.db.GetUserIDByUsername(username)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			context.Logger.WithError(err).Error("User not found")
// 			http.Error(w, "User not found", http.StatusNotFound)
// 			return
// 		}
// 		context.Logger.WithError(err).Error("Error fetching user ID")
// 		http.Error(w, "Internal server error", http.StatusInternalServerError)
// 		return
// 	}

// 	// Respond with the user ID
// 	w.WriteHeader(http.StatusOK)
// 	err = json.NewEncoder(w).Encode(map[string]string{"user_id": userID})
// 	if err != nil {
// 		context.Logger.WithError(err).Error("Error encoding response")
// 		http.Error(w, "Error encoding response", http.StatusInternalServerError)
// 	}
// }
