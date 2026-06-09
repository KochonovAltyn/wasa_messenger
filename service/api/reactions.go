package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"

	"github.com/KochonovAltyn/wasa_messenger/service/api/reqcontext"
)

// setReaction handles PUT /conversations/:conversation_id/messages/:message_id/reaction
// It sets (or replaces) the authenticated user's emoji reaction on a message.
func (rt *_router) setReaction(w http.ResponseWriter, r *http.Request, ps httprouter.Params, context *reqcontext.RequestContext) {
	userID := context.UserID
	if userID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// PUT route uses the :c_id wildcard (see api-handler.go note).
	conversationID, err := strconv.Atoi(ps.ByName("c_id"))
	if err != nil || conversationID <= 0 {
		http.Error(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	messageID, err := strconv.Atoi(ps.ByName("message_id"))
	if err != nil || messageID <= 0 {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	// Only members of the conversation may react.
	isMember, err := rt.db.IsUserInConversation(userID, conversationID)
	if err != nil {
		context.Logger.WithError(err).Error("error checking membership")
		http.Error(w, "Error checking user membership", http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "User is not part of this conversation", http.StatusForbidden)
		return
	}

	// Parse the emoji from the request body.
	var input struct {
		Emoji string `json:"emoji"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Emoji == "" {
		http.Error(w, "Invalid input: emoji is required", http.StatusBadRequest)
		return
	}

	if err := rt.db.SetReaction(messageID, userID, input.Emoji); err != nil {
		context.Logger.WithError(err).Error("failed to set reaction")
		http.Error(w, "Failed to set reaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Reaction set successfully"})
}

// removeReaction handles DELETE /conversations/:conversation_id/messages/:message_id/reaction
// It removes the authenticated user's reaction from a message.
func (rt *_router) removeReaction(w http.ResponseWriter, r *http.Request, ps httprouter.Params, context *reqcontext.RequestContext) {
	userID := context.UserID
	if userID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	conversationID, err := strconv.Atoi(ps.ByName("conversation_id"))
	if err != nil || conversationID <= 0 {
		http.Error(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	messageID, err := strconv.Atoi(ps.ByName("message_id"))
	if err != nil || messageID <= 0 {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	isMember, err := rt.db.IsUserInConversation(userID, conversationID)
	if err != nil {
		context.Logger.WithError(err).Error("error checking membership")
		http.Error(w, "Error checking user membership", http.StatusInternalServerError)
		return
	}
	if !isMember {
		http.Error(w, "User is not part of this conversation", http.StatusForbidden)
		return
	}

	if err := rt.db.RemoveReaction(messageID, userID); err != nil {
		context.Logger.WithError(err).Error("failed to remove reaction")
		http.Error(w, "Failed to remove reaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Reaction removed successfully"})
}
