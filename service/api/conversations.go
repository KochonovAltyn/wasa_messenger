package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/KochonovAltyn/wasa_messenger/service/api/reqcontext"
)

func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, ps httprouter.Params, context *reqcontext.RequestContext) {
	// Log the request
	context.Logger.Info("Request received: method=%s, path=%s", r.Method, r.URL.Path)

	// Retrieve user_id from the path parameters
	userID := ps.ByName("id")
	if userID == "" {
		context.Logger.Error("User ID is required in the path")
		http.Error(w, "User ID is required in the path", http.StatusBadRequest)
		return
	}

	// Fetch conversations from the database
	conversations, err := rt.db.GetMyConversations_db(userID)
	if err != nil {
		context.Logger.WithError(err).Error("Error fetching conversations")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Respond with the list of conversations
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(conversations)
	if err != nil {
		context.Logger.WithError(err).Error("Error encoding response")
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}
func (rt *_router) sendMessageFirst(w http.ResponseWriter, r *http.Request, ps httprouter.Params, context *reqcontext.RequestContext) {
	// Log the request
	context.Logger.Info("Request received: method=%s, path=%s", r.Method, r.URL.Path)

	// Extract sender ID from the path
	senderID := ps.ByName("id")
	if senderID == "" {
		context.Logger.Error("Sender ID is required in the path")
		http.Error(w, "Sender ID is required in the path", http.StatusBadRequest)
		return
	}

	// Extract recipient username from the form data
	recipientUsername := r.FormValue("recipient_username")
	if recipientUsername == "" {
		recipientUsername = r.URL.Query().Get("recipient_username") // Fallback
	}
	if recipientUsername == "" {
		context.Logger.Error("Recipient username is required")
		http.Error(w, "Recipient username is required", http.StatusBadRequest)
		return
	}

	// Retrieve recipient user ID by username
	recipientID, err := rt.db.GetUserIDByUsername(recipientUsername)
	if err != nil {
		context.Logger.WithError(err).Error("Recipient not found")
		http.Error(w, "Recipient not found", http.StatusNotFound)
		return
	}

	// Validate sender
	sender, err := rt.db.GetUserId(senderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			context.Logger.WithError(err).Error("Sender not found")
			http.Error(w, "Sender not found", http.StatusNotFound)
			return
		}
		context.Logger.WithError(err).Error("Error fetching sender")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Validate recipient
	recipient, err := rt.db.GetUserId(recipientID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			context.Logger.WithError(err).Error("Recipient not found")
			http.Error(w, "Recipient not found", http.StatusNotFound)
			return
		}
		context.Logger.WithError(err).Error("Error fetching recipient")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check if a private conversation already exists
	exists, err := rt.db.ConversationExists(senderID, recipientID)
	if err != nil {
		context.Logger.WithError(err).Error("Error checking for existing conversation")
		http.Error(w, "Internal server error: failed to check conversation", http.StatusInternalServerError)
		return
	}
	if exists {
		context.Logger.Error("A private conversation already exists between these users")
		http.Error(w, "A private conversation already exists between these users", http.StatusConflict)
		return
	}

	// Create a new private conversation
	newConvo, err := rt.db.CreateConversation_db(false, "", "")
	if err != nil {
		context.Logger.WithError(err).Error("Error creating conversation")
		http.Error(w, "Error creating conversation", http.StatusInternalServerError)
		return
	}

	// Add sender to the conversation
	err = rt.db.AddUsersToConversation(sender.ID, newConvo.ID)
	if err != nil {
		context.Logger.WithError(err).Error("Error adding sender to conversation")
		http.Error(w, "Error adding sender to conversation", http.StatusInternalServerError)
		return
	}

	// Add recipient to the conversation
	err = rt.db.AddUsersToConversation(recipient.ID, newConvo.ID)
	if err != nil {
		context.Logger.WithError(err).Error("Error adding recipient to conversation")
		http.Error(w, "Error adding recipient to conversation", http.StatusInternalServerError)
		return
	}

	// Handle file uploads (photo or GIF). Any text sent along with the file
	// is kept as the caption, so one message can hold both image and text.
	file, header, err := r.FormFile("file")
	var contentType, content, caption string
	if err == nil { // File is uploaded
		defer file.Close()

		caption = r.FormValue("content")

		// Save the file and get path & type
		contentType, content, err = rt.db.SaveUploadedFile(file, header, senderID)
		if err != nil {
			context.Logger.WithError(err).Error("Failed to save uploaded file")
			http.Error(w, "Failed to save uploaded file", http.StatusInternalServerError)
			return
		}
	} else {
		// Parse request body for text messages
		contentType = r.FormValue("content_type")
		content = r.FormValue("content")

		if contentType == "" || content == "" {
			context.Logger.Error("Invalid input: content_type and content are required")
			http.Error(w, "Invalid input: content_type and content are required", http.StatusBadRequest)
			return
		}
	}

	// Send the first message
	err = rt.db.SendMessageWithMedia(newConvo.ID, sender.ID, contentType, content, caption)
	if err != nil {
		context.Logger.WithError(err).Error("Error sending message")
		http.Error(w, "Error sending message", http.StatusInternalServerError)
		return
	}

	// Respond with conversation ID
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Message sent successfully",
		"c_id":    newConvo.ID,
	})
	if err != nil {
		context.Logger.WithError(err).Error("Error encoding response")
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}


func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, context *reqcontext.RequestContext) {
	context.Logger.Info("Request received: method=%s, path=%s", r.Method, r.URL.Path)

	// Extract conversation ID
	conversationIDStr := ps.ByName("conversation_id")
	conversationID, err := strconv.Atoi(conversationIDStr)
	if err != nil || conversationID <= 0 {
		context.Logger.WithError(err).Error("Invalid conversation ID")
		http.Error(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}
	context.Logger.Infof("Extracted conversation_id: %d", conversationID)

	// Extract sender ID (the authenticated user)
	senderID := context.UserID
	if senderID == "" {
		context.Logger.Error("Sender ID is required")
		http.Error(w, "Sender ID is required", http.StatusUnauthorized)
		return
	}
	context.Logger.Infof("Extracted sender_id: %s", senderID)

	// Ensure the sender is a participant in the conversation
	isMember, err := rt.db.IsUserInConversation(senderID, conversationID)
	if err != nil {
		context.Logger.WithError(err).Error("Error checking if user is in conversation")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !isMember {
		context.Logger.Error("Sender is not part of the conversation")
		http.Error(w, "Sender is not part of the conversation", http.StatusForbidden)
		return
	} 

	// ----------------------------------------------------------------
	// OPTIONALLY parse "reply_to" from form data (if user is replying)
	// ----------------------------------------------------------------
	replyToStr := r.FormValue("reply_to")
	var replyTo *int
	if replyToStr != "" {
		val, convErr := strconv.Atoi(replyToStr)
		if convErr == nil {
			replyTo = &val
		} else {
			// If invalid, we can just log a warning or ignore
			context.Logger.WithError(convErr).Warnf("Ignoring invalid reply_to=%q", replyToStr)
		}
	}

	// Check if a file is uploaded (photo, GIF, etc.) or if it's just text.
	// A single message can carry BOTH an image and text: the typed text is
	// stored as the caption of the media message.
	var content, contentType, caption string
	file, header, fileErr := r.FormFile("file")

	if fileErr == nil { // A file is uploaded
		defer file.Close()

		// Keep any text typed together with the image as the caption.
		caption = r.FormValue("content")

		// Validate file extension
		fileExt := strings.ToLower(filepath.Ext(header.Filename))
		allowedExts := map[string]string{
			".jpg":  "photo",
			".jpeg": "photo",
			".png":  "photo",
			".gif":  "gif",
		}

		fileType, valid := allowedExts[fileExt]
		if !valid {
			context.Logger.Error("Invalid file type")
			http.Error(w, "Invalid file type", http.StatusBadRequest)
			return
		}

		// Ensure the upload directory exists
		uploadDir := "webui/public/uploads"
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			context.Logger.WithError(err).Error("Failed to create upload directory")
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			return
		}

		// Generate a unique filename
		fileName := senderID + "_" + strconv.Itoa(int(time.Now().Unix())) + fileExt
		filePath := filepath.Join(uploadDir, fileName)

		out, err := os.Create(filePath)
		if err != nil {
			context.Logger.WithError(err).Error("Failed to save file")
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			return
		}
		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			context.Logger.WithError(err).Error("Failed to write file")
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			return
		}

		// Set message content as the uploaded file path
		content = "/uploads/" + fileName
		contentType = fileType

	} else {
		// Handle text message (no file)
		content = r.FormValue("content")
		contentType = r.FormValue("content_type")

		if content == "" || contentType == "" {
			context.Logger.Error("Invalid input: content and content_type are required")
			http.Error(w, "Invalid input: content and content_type are required", http.StatusBadRequest)
			return
		}
	}

	// Fetch the username and photo for the response
	user, err := rt.db.GetUserByID(senderID)
	if err != nil {
		context.Logger.WithError(err).Error("Failed to fetch user")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// ---------------------------------------------------------------------
	// Save message to database, passing the optional replyTo reference
	// ---------------------------------------------------------------------
	err = rt.db.SendMessageWithType(conversationID, senderID, content, contentType, caption, replyTo)
	if err != nil {
		context.Logger.WithError(err).Error("Error saving message")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return JSON response with some data about the message
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]interface{}{
		"message":         "Message sent successfully",
		"content_type":    contentType,
		"content":         content,
		"caption":         caption,
		"sender_username": user.Username,
		"sender_photo":    user.Photo.String, // could be empty if no photo
	})
	if err != nil {
		context.Logger.WithError(err).Error("Error encoding response")
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

func (rt *_router) getConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, context *reqcontext.RequestContext) {
	// Extract conversation ID from the path parameters
	conversationID, err := strconv.Atoi(ps.ByName("c_id"))
	if err != nil || conversationID <= 0 {
		context.Logger.WithError(err).Error("Invalid conversation ID")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid conversation ID"})
		return
	}

	// Fetch the conversation details
	conversation, err := rt.db.GetConversationById(conversationID)
	if err != nil {
		context.Logger.WithError(err).Error("Failed to fetch conversation")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch conversation"})
		return
	}

	if conversation.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Conversation not found"})
		return
	}

	// The current user is opening this conversation, so mark all incoming
	// messages (sent by others) as read. This is what lights up the sender's
	// "read" checkmark. We do this before fetching so the returned messages
	// already carry the updated status.
	if err := rt.db.MarkMessagesAsRead(conversationID, context.UserID); err != nil {
		// Non-fatal: log and continue, the conversation can still be shown.
		context.Logger.WithError(err).Warn("failed to mark messages as read")
	}

	// Fetch all messages in the conversation
	messages, err := rt.db.GetMessagesByConversationId(conversationID)
	if err != nil {
		context.Logger.WithError(err).Error("Failed to fetch messages")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch messages"})
		return
	}

	// Ensure each message has a valid sender_photo, otherwise use default,
	// and attach the emoji reactions placed on each message.
	for i, message := range messages {
		if !message.SenderPhoto.Valid || message.SenderPhoto.String == "" {
			messages[i].SenderPhoto.String = "/default-profile.png" // Set to default profile image
		}

		reactions, err := rt.db.GetReactionsByMessageID(message.ID)
		if err != nil {
			context.Logger.WithError(err).Warn("failed to fetch reactions for message")
		} else {
			messages[i].Reactions = reactions
		}
	}

	// Prepare the response
	response := map[string]interface{}{
		"conversation": conversation,
		"messages":     messages,
	}

	// Respond with the conversation and messages
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, context *reqcontext.RequestContext) {
	context.Logger.Info("Request received: DELETE message")

	// Extract conversation ID and message ID
	conversationIDStr := ps.ByName("conversation_id")
	messageIDStr := ps.ByName("message_id")

	if conversationIDStr == "" || messageIDStr == "" {
		context.Logger.Error("Missing conversation_id or message_id in path")
		http.Error(w, "Invalid conversation ID or message ID", http.StatusBadRequest)
		return
	}

	conversationID, err := strconv.Atoi(conversationIDStr)
	if err != nil || conversationID <= 0 {
		context.Logger.WithError(err).Error("Invalid conversation ID")
		http.Error(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	messageID, err := strconv.Atoi(messageIDStr)
	if err != nil || messageID <= 0 {
		context.Logger.WithError(err).Error("Invalid message ID")
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	userID := context.UserID
	if userID == "" {
		context.Logger.Error("User not authenticated")
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// ✅ Ensure the message exists
	exists, err := rt.db.DoesMessageExist(messageID)
	if err != nil {
		context.Logger.WithError(err).Error("Error checking message existence")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !exists {
		context.Logger.Error("Message not found")
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	// ✅ Convert comments to normal messages before deleting
	err = rt.db.ConvertCommentsToMessages(messageID, conversationID)
	if err != nil {
		context.Logger.WithError(err).Error("Error converting comments to messages")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// ✅ Delete the message
	err = rt.db.DeleteMessage(messageID)
	if err != nil {
		context.Logger.WithError(err).Error("Error deleting message")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Message deleted successfully, comments converted to normal messages",
	})
}

func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, context *reqcontext.RequestContext) {
	// Log the request
	context.Logger.Info("Request received: method=%s, path=%s", r.Method, r.URL.Path)

	// Step 1: Extract source conversation ID and message ID.
	sourceConversationIDStr := ps.ByName("conversation_id")
	sourceConversationID, err := strconv.Atoi(sourceConversationIDStr)
	if err != nil || sourceConversationID <= 0 {
		context.Logger.WithError(err).Error("Invalid source conversation ID")
		http.Error(w, "Invalid source conversation ID", http.StatusBadRequest)
		return
	}
	messageIDStr := ps.ByName("message_id")
	messageID, err := strconv.Atoi(messageIDStr)
	if err != nil || messageID <= 0 {
		context.Logger.WithError(err).Error("Invalid message ID")
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	// Step 2: Extract target conversation identifier.
	// It can either be a numeric ID or the literal "new"
	targetConversationIDStr := ps.ByName("target_conversation_id")
	var targetConversationID int

	// Step 3: Get the forwarding user from context.
	userID := context.UserID
	if userID == "" {
		context.Logger.Error("User not authenticated")
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	if targetConversationIDStr == "new" {
		// Forward to a new conversation using a target name (which can be a group or a user).
		var input struct {
			TargetUsername string `json:"target_username"` // if not a group, this is a user name
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			context.Logger.WithError(err).Error("Failed to decode request body for forward message")
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}
		if input.TargetUsername == "" {
			context.Logger.Error("No target username provided")
			http.Error(w, "Target username required", http.StatusBadRequest)
			return
		}
		context.Logger.Info("Forward request target name:", input.TargetUsername)
		// First try to see if this target name corresponds to a group conversation.
		groupConv, err := rt.db.GetGroupByName(input.TargetUsername)
		if err != nil {
			context.Logger.WithError(err).Error("Error searching for group")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if groupConv.ID != 0 {
			// Found a group conversation. Verify membership.
			isMember, err := rt.db.IsUserInConversation(userID, groupConv.ID)
			if err != nil {
				context.Logger.WithError(err).Error("Error checking membership in target group")
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			if !isMember {
				context.Logger.Error("User is not part of the target group")
				http.Error(w, "User is not part of the target group", http.StatusForbidden)
				return
			}
			context.Logger.Info("Forwarding to group with conversation ID:", groupConv.ID)
			targetConversationID = groupConv.ID
		} else {
			// Otherwise, treat it as a target user (one-on-one conversation)
			context.Logger.Info("No group found. Looking up target user:", input.TargetUsername)
			targetUser, err := rt.db.GetUser(input.TargetUsername)
			if err != nil {
				context.Logger.WithError(err).Error("Target user not found")
				http.Error(w, "Target user not found", http.StatusNotFound)
				return
			}
			// Check if a one-on-one conversation between userID and targetUser.ID exists.
			conv, err := rt.db.GetConversationBetweenUsers(userID, targetUser.ID)
			if err != nil {
				context.Logger.WithError(err).Error("Error checking conversation between users")
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			if conv.ID == 0 {
				context.Logger.Info("No existing one-on-one conversation found. Creating new conversation...")
				conv, err = rt.db.CreateConversation_db(false, "", "")
				if err != nil {
					context.Logger.WithError(err).Error("Error creating new conversation")
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
				// Add both users to the conversation.
				if err := rt.db.AddUsersToConversation(userID, conv.ID); err != nil {
					context.Logger.WithError(err).Error("Error adding forwarding user to conversation")
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
				if err := rt.db.AddUsersToConversation(targetUser.ID, conv.ID); err != nil {
					context.Logger.WithError(err).Error("Error adding target user to conversation")
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
			}
			targetConversationID = conv.ID
		}
	} else {
		// Otherwise, parse the target conversation ID as a number.
		targetConversationID, err = strconv.Atoi(targetConversationIDStr)
		if err != nil || targetConversationID <= 0 {
			context.Logger.WithError(err).Error("Invalid target conversation ID")
			http.Error(w, "Invalid target conversation ID", http.StatusBadRequest)
			return
		}
		// If the target conversation is numeric, check if it is a group.
		isGroup, err := rt.db.IsConversationGroup(targetConversationID)
		if err != nil {
			context.Logger.WithError(err).Error("Error checking target conversation type")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if isGroup {
			groupName, err := rt.db.GetGroupNameById(targetConversationID)
			if err != nil {
				context.Logger.WithError(err).Error("Error retrieving group name")
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			context.Logger.Info("Forwarding to group:", groupName)
		}
	}

	// Step 4: Validate membership.
	// Check that the user is part of the source conversation.
	isMemberSource, err := rt.db.IsUserInConversation(userID, sourceConversationID)
	if err != nil {
		context.Logger.WithError(err).Error("Error checking membership in source conversation")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !isMemberSource {
		context.Logger.Error("User is not part of the source conversation")
		http.Error(w, "User is not part of the source conversation", http.StatusForbidden)
		return
	}
	// For target conversation:
	// If it is a group, check that the user is a member.
	isGroup, err := rt.db.IsConversationGroup(targetConversationID)
	if err != nil {
		context.Logger.WithError(err).Error("Error checking conversation type")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if isGroup {
		isMemberTarget, err := rt.db.IsUserInConversation(userID, targetConversationID)
		if err != nil {
			context.Logger.WithError(err).Error("Error checking membership in target group conversation")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if !isMemberTarget {
			context.Logger.Error("User is not part of the target group conversation")
			http.Error(w, "User is not part of the target group conversation", http.StatusForbidden)
			return
		}
	} else {
		// For one-on-one conversations, ensure that the forwarding user is a member.
		isMemberTarget, err := rt.db.IsUserInConversation(userID, targetConversationID)
		if err != nil {
			context.Logger.WithError(err).Error("Error checking membership in target conversation")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if !isMemberTarget {
			context.Logger.Error("User is not part of the target conversation")
			http.Error(w, "User is not part of the target conversation", http.StatusForbidden)
			return
		}
	}

	// Step 5: Retrieve the message content from the source conversation.
	messageContent, err := rt.db.GetMessageContent(messageID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			context.Logger.WithError(err).Error("Message not found")
			http.Error(w, "Message not found", http.StatusNotFound)
			return
		}
		context.Logger.WithError(err).Error("Error fetching message content")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Step 6: Forward the message.
	// The new message will be created with sender = userID (i.e. the forwarding user).
	err = rt.db.ForwardMessage(targetConversationID, userID, messageContent)
	if err != nil {
		context.Logger.WithError(err).Error("Error forwarding message")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Step 7: Respond with success.
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Message forwarded successfully",
	})
}

// createGroup handles the creation of a new group conversation
func (rt *_router) createGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, context *reqcontext.RequestContext) {
	creatorID := context.UserID // Get authenticated user (creator)
	if creatorID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Parse the incoming form data (10MB limit)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	groupName := r.FormValue("group_name")
	usernamesRaw := r.FormValue("usernames")

	// Parse usernames from JSON string
	var usernames []string
	if usernamesRaw != "" {
		err = json.Unmarshal([]byte(usernamesRaw), &usernames)
		if err != nil {
			http.Error(w, "Invalid usernames format", http.StatusBadRequest)
			return
		}
	}

	if groupName == "" || len(usernames) == 0 {
		http.Error(w, "Invalid input: group_name and usernames are required", http.StatusBadRequest)
		return
	}

	// Handle the uploaded photo
	var photoPath string
	photoFile, _, err := r.FormFile("photo") // Get the uploaded file
	if err == nil {
		// Save the photo to disk and get the file path
		_, filePath, saveErr := rt.db.SaveUploadedFile(photoFile, r.MultipartForm.File["photo"][0], creatorID)
		if saveErr != nil {
			http.Error(w, "Error saving photo", http.StatusInternalServerError)
			return
		}
		photoPath = filePath
	}

	// Step 1: Create a new group conversation
	newGroup, err := rt.db.CreateConversation_db(true, groupName, photoPath)
	if err != nil {
		http.Error(w, "Error creating group", http.StatusInternalServerError)
		return
	}

	// Step 2: Add the creator to the group
	err = rt.db.AddUsersToConversation(creatorID, newGroup.ID)
	if err != nil {
		http.Error(w, "Error adding creator to group", http.StatusInternalServerError)
		return
	}

	// Step 3: Fetch user IDs for given usernames and add them to the group
	for _, username := range usernames {
		user, err := rt.db.GetUser(username)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "User '"+username+"' not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Error fetching user "+username, http.StatusInternalServerError)
			return
		}

		err = rt.db.AddUsersToConversation(user.ID, newGroup.ID)
		if err != nil {
			http.Error(w, "Error adding user "+username+" to group", http.StatusInternalServerError)
			return
		}
	}

	// Step 4: Respond with success, group details, and members
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Group created successfully",
		"group_id":   newGroup.ID,
		"c_id":       newGroup.ID,
		"group_name": newGroup.Name,
		"photo":      newGroup.Photo.String, // This should now correctly include the photo path
		"members":    usernames,
	})
}

func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, context *reqcontext.RequestContext) {
	// Extract user making the request (must be part of the group)
	requesterID := context.UserID
	if requesterID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Extract group conversation ID (`c_id`) from the URL
	conversationIDStr := ps.ByName("c_id")
	conversationID, err := strconv.Atoi(conversationIDStr)
	if err != nil || conversationID <= 0 {
		http.Error(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	// Parse request body to get the list of usernames to add
	var input struct {
		Usernames []string `json:"usernames"` // List of usernames to add
	}
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil || len(input.Usernames) == 0 {
		http.Error(w, "Invalid input: at least one username is required", http.StatusBadRequest)
		return
	}

	// Check if the requester is part of the group
	isMember, err := rt.db.IsUserInConversation(requesterID, conversationID)
	if err != nil {
		http.Error(w, "Error checking membership", http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "You must be a member of the group to add others", http.StatusForbidden)
		return
	}

	// Retrieve user IDs for each provided username
	var addedUsers []string
	for _, username := range input.Usernames {
		user, err := rt.db.GetUser(username)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "User not found: "+username, http.StatusNotFound)
				return
			}
			http.Error(w, "Error retrieving user: "+username, http.StatusInternalServerError)
			return
		}

		// Check if the user is already in the group
		alreadyMember, err := rt.db.IsUserInConversation(user.ID, conversationID)
		if err != nil {
			http.Error(w, "Error checking user membership", http.StatusInternalServerError)
			return
		}
		if alreadyMember {
			continue // Skip users already in the group
		}

		// Add user to the group
		err = rt.db.AddUsersToConversation(user.ID, conversationID)
		if err != nil {
			http.Error(w, "Error adding user: "+username, http.StatusInternalServerError)
			return
		}

		// Add to the response list
		addedUsers = append(addedUsers, username)
	}

	// Respond with success and list of added users
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     "Users added to group successfully",
		"c_id":        conversationID,
		"added_users": addedUsers,
	})
}

func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx *reqcontext.RequestContext) {
	userID := ctx.UserID
	groupIDStr := ps.ByName("c_id") // Extract group (conversation) ID

	// Convert groupID to integer
	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil || groupID <= 0 {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	// ✅ Ensure this is a valid group conversation
	isGroup, err := rt.db.IsConversationGroup(groupID)
	if err != nil {
		http.Error(w, "Error checking group type", http.StatusInternalServerError)
		return
	}
	if !isGroup {
		http.Error(w, "This conversation is not a group", http.StatusBadRequest)
		return
	}

	// ✅ Check if the user is a member of the group
	isMember, err := rt.db.IsUserInConversation(userID, groupID)
	if err != nil {
		http.Error(w, "Error checking user membership", http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "User is not part of this group", http.StatusForbidden)
		return
	}

	// ✅ Remove the user from the group
	err = rt.db.RemoveUserFromGroup(userID, groupID)
	if err != nil {
		http.Error(w, "Error leaving the group", http.StatusInternalServerError)
		return
	}

	// ✅ Check if the group is now empty
	remainingMembers, err := rt.db.GetGroupMemberCount(groupID)
	if err != nil {
		http.Error(w, "Error checking remaining group members", http.StatusInternalServerError)
		return
	}

	// ✅ If no members are left, delete the group (optional)
	if remainingMembers == 0 {
		err = rt.db.DeleteGroup(groupID)
		if err != nil {
			http.Error(w, "Error deleting empty group", http.StatusInternalServerError)
			return
		}
	}

	// ✅ Respond with success
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Successfully left the group"})
}

func (rt *_router) setGroupName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, context *reqcontext.RequestContext) {
	userID := context.UserID // ✅ Get authenticated user
	if userID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// ✅ Extract group ID from URL
	groupIDStr := ps.ByName("c_id")
	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil || groupID <= 0 {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	// ✅ Parse request body to get new group name
	var input struct {
		NewName string `json:"new_name"`
	}
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil || input.NewName == "" {
		http.Error(w, "Invalid input: new_name is required", http.StatusBadRequest)
		return
	}

	// ✅ Check if the conversation is a group
	isGroup, err := rt.db.IsConversationGroup(groupID)
	if err != nil {
		http.Error(w, "Error checking group type", http.StatusInternalServerError)
		return
	}
	if !isGroup {
		http.Error(w, "This conversation is not a group", http.StatusBadRequest)
		return
	}

	// ✅ Check if the user is a member of the group
	isMember, err := rt.db.IsUserInConversation(userID, groupID)
	if err != nil {
		http.Error(w, "Error checking user membership", http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "User is not part of this group", http.StatusForbidden)
		return
	}

	// ✅ Check if the new group name is already taken
	nameExists, err := rt.db.GroupNameExists(input.NewName)
	if err != nil {
		http.Error(w, "Error checking group name availability", http.StatusInternalServerError)
		return
	}
	if nameExists {
		http.Error(w, "A group with this name already exists", http.StatusConflict)
		return
	}

	// ✅ Update the group name
	err = rt.db.UpdateGroupName(groupID, input.NewName)
	if err != nil {
		http.Error(w, "Error updating group name", http.StatusInternalServerError)
		return
	}

	// ✅ Respond with success
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message":  "Group name updated successfully",
		"group_id": groupIDStr,
		"new_name": input.NewName,
	})
}

func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx *reqcontext.RequestContext) {
	// Extract authenticated user ID
	userID := ctx.UserID
	if userID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Extract group (conversation) ID
	groupIDStr := ps.ByName("c_id")
	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil || groupID <= 0 {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	// Check if the conversation is a group
	isGroup, err := rt.db.IsConversationGroup(groupID)
	if err != nil {
		http.Error(w, "Error checking conversation type", http.StatusInternalServerError)
		return
	}
	if !isGroup {
		http.Error(w, "This conversation is not a group", http.StatusBadRequest)
		return
	}

	// Check if the user is a member of the group
	isMember, err := rt.db.IsUserInConversation(userID, groupID)
	if err != nil {
		http.Error(w, "Error checking user membership", http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "User is not part of this group", http.StatusForbidden)
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

	// Generate new filename using the group ID
	fileName := "group_" + strconv.Itoa(groupID) + fileExt
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

	// Generate the correct photo URL for the frontend. The file name is fixed
	// per group, so append a version parameter: the URL changes on every
	// update and browsers re-fetch the image instead of serving the cached
	// one — this is what makes the new photo show up instantly everywhere.
	photoURL := "/uploads/" + fileName + "?v=" + strconv.FormatInt(time.Now().Unix(), 10)

	// Update the group's profile photo in the database
	err = rt.db.UpdateGroupPhoto(groupID, photoURL)
	if err != nil {
		http.Error(w, "Failed to update group photo", http.StatusInternalServerError)
		return
	}

	// ✅ Respond with the updated group photo URL
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Group photo updated successfully",
		"photo":   photoURL,
	})
}

func (rt *_router) commentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, context *reqcontext.RequestContext) {
	// Extract authenticated user ID
	userID := context.UserID
	if userID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Extract conversation ID
	conversationIDStr := ps.ByName("conversation_id")
	conversationID, err := strconv.Atoi(conversationIDStr)
	if err != nil || conversationID <= 0 {
		http.Error(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	// Extract message ID (the message being commented on)
	messageIDStr := ps.ByName("message_id")
	messageID, err := strconv.Atoi(messageIDStr)
	if err != nil || messageID <= 0 {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	// Check if the user is a member of the conversation
	isMember, err := rt.db.IsUserInConversation(userID, conversationID)
	if err != nil {
		http.Error(w, "Error checking user membership", http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "User is not part of this conversation", http.StatusForbidden)
		return
	}

	// Handle file uploads (photo or GIF)
	file, header, err := r.FormFile("file")
	var contentType, content string
	if err == nil { // File is uploaded
		defer file.Close()

		// Validate file type
		fileExt := strings.ToLower(filepath.Ext(header.Filename))
		allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
		if !allowedExts[fileExt] {
			http.Error(w, "Invalid file type", http.StatusBadRequest)
			return
		}

		// Ensure uploads directory exists
		uploadDir := "webui/public/uploads"
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			context.Logger.WithError(err).Error("Failed to create upload directory")
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			return
		}

		// Generate unique filename
		fileName := userID + "_" + strconv.Itoa(messageID) + fileExt
		filePath := filepath.Join(uploadDir, fileName)

		// Save file
		out, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			return
		}
		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			return
		}

		// Set contentType and content for DB
		if fileExt == ".gif" {
			contentType = "gif"
		} else {
			contentType = "photo"
		}
		content = "/uploads/" + fileName // ✅ Store relative file path
	} else {
		// Parse request body for text or emoji comment
		var input struct {
			ContentType string `json:"content_type"` // "text", "emoji"
			Content     string `json:"content"`
		}
		err = json.NewDecoder(r.Body).Decode(&input)
		if err != nil || input.ContentType == "" || input.Content == "" {
			http.Error(w, "Invalid input: content_type and content are required", http.StatusBadRequest)
			return
		}
		contentType = input.ContentType
		content = input.Content
	}

	// ✅ Check if the commented message still exists
	exists, err := rt.db.DoesMessageExist(messageID)
	if err != nil {
		http.Error(w, "Error checking message existence", http.StatusInternalServerError)
		return
	}

	// ✅ If message is deleted, comment becomes a normal message
	if !exists {
		err = rt.db.SendMessageFull(conversationID, userID, content)
		if err != nil {
			http.Error(w, "Error sending message", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "Original message deleted. Comment added as normal message.",
		})
		return
	}

	// ✅ If message exists, add a comment linked to the message
	err = rt.db.CommentOnMessage(messageID, userID, contentType, content)
	if err != nil {
		http.Error(w, "Error commenting on message", http.StatusInternalServerError)
		return
	}

	// ✅ Respond with success
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message":      "Comment added successfully",
		"message_id":   strconv.Itoa(messageID),
		"content_type": contentType,
		"content":      content,
	})
}

func (rt *_router) uncommentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, context *reqcontext.RequestContext) {
	// Extract user ID
	userID := context.UserID
	if userID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Extract parameters from URL
	conversationIDStr := ps.ByName("conversation_id")
	messageIDStr := ps.ByName("message_id")
	commentIDStr := ps.ByName("comment_id")

	conversationID, err1 := strconv.Atoi(conversationIDStr)
	messageID, err2 := strconv.Atoi(messageIDStr)
	commentID, err3 := strconv.Atoi(commentIDStr)

	if err1 != nil || err2 != nil || err3 != nil || conversationID <= 0 || messageID <= 0 || commentID <= 0 {
		http.Error(w, "Invalid conversation, message, or comment ID", http.StatusBadRequest)
		return
	}

	// ✅ Check if the user is a member of the conversation
	isMember, err := rt.db.IsUserInConversation(userID, conversationID)
	if err != nil {
		http.Error(w, "Error checking user membership", http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "User is not part of this conversation", http.StatusForbidden)
		return
	}

	// ✅ Check if the user is the owner of the comment
	isOwner, err := rt.db.IsCommentOwner(userID, commentID)
	if err != nil {
		http.Error(w, "Error checking comment ownership", http.StatusInternalServerError)
		return
	}
	if !isOwner {
		http.Error(w, "User does not own this comment", http.StatusForbidden)
		return
	}

	// ✅ Delete the comment
	err = rt.db.DeleteComment(commentID)
	if err != nil {
		http.Error(w, "Error deleting comment", http.StatusInternalServerError)
		return
	}

	// ✅ Respond with success
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Comment deleted successfully",
	})
}

func (rt *_router) getComments(w http.ResponseWriter, r *http.Request, ps httprouter.Params, context *reqcontext.RequestContext) {
	messageIDStr := ps.ByName("message_id")
	messageID, err := strconv.Atoi(messageIDStr)
	if err != nil || messageID <= 0 {
		http.Error(w, "Invalid message id", http.StatusBadRequest)
		return
	}

	comments, err := rt.db.GetCommentsByMessageID(messageID)
	if err != nil {
		context.Logger.WithError(err).Error("Error retrieving comments")
		// Optionally, write out the error in plain text to help debugging (remove in production)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(comments); err != nil {
		context.Logger.WithError(err).Error("Error encoding comments response")
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (rt *_router) searchUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx *reqcontext.RequestContext) {
	// Get the query parameter "username"
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Missing username query parameter", http.StatusBadRequest)
		return
	}

	// Look up the user in the database by username.
	user, err := rt.db.GetUser(username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "No such user found", http.StatusNotFound)
			return
		}
		ctx.Logger.WithError(err).Error("Error searching for user")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check if a one-on-one conversation already exists between the current user and the searched user.
	convo, err := rt.db.GetConversationBetweenUsers(ctx.UserID, user.ID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Error checking conversation")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Prepare the response payload.
	response := map[string]interface{}{
		"user": user,
	}
	if convo.ID != 0 {
		response["conversation_id"] = convo.ID
	} else {
		response["conversation_id"] = nil
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		ctx.Logger.WithError(err).Error("Error encoding search response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
