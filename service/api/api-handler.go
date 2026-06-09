package api

import (
	"net/http"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {
	// Register routes
	rt.router.GET("/liveness", rt.liveness)         // done
	rt.router.POST("/session", rt.wrap(rt.doLogin)) // done
	rt.router.ServeFiles("/uploads/*filepath", http.Dir("webui/public/uploads"))

	// rt.router.POST("/logout", rt.wrap(rt.logout))
	rt.router.PUT("/users/me/username", rt.wrap(rt.setMyUserName))                                      // done
	rt.router.PUT("/users/me/photo", rt.wrap(rt.setMyPhoto))                                            // done
	rt.router.GET("/users/:id/conversations", rt.wrap(rt.getMyConversations))                           // done
	rt.router.POST("/users/:id/conversations/first-message", rt.wrap(rt.sendMessageFirst))              // done
	rt.router.POST("/conversations/:conversation_id/messages", rt.wrap(rt.sendMessage))                 // done
	rt.router.GET("/conversations/:c_id", rt.wrap(rt.getConversation))                                  // done
	rt.router.DELETE("/conversations/:conversation_id/messages/:message_id", rt.wrap(rt.deleteMessage)) // done
	rt.router.POST("/conversations/:conversation_id/messages/:message_id/forward/:target_conversation_id", rt.wrap(rt.forwardMessage))
	rt.router.POST("/groups", rt.wrap(rt.createGroup))                               // done
	rt.router.POST("/groups/:c_id/members", rt.wrap(rt.addToGroup))                  // done
	rt.router.DELETE("/groups/:c_id/leave", rt.wrap(rt.leaveGroup))                  // done
	rt.router.PUT("/groups/:c_id/name", rt.wrap(rt.setGroupName))                    // done
	rt.router.PUT("/conversations/:c_id/set-group-photo", rt.wrap(rt.setGroupPhoto)) // done
	rt.router.POST("/conversations/:conversation_id/messages/:message_id/comments", rt.wrap(rt.commentMessage))
	rt.router.DELETE("/conversations/:conversation_id/messages/:message_id/comments/:comment_id", rt.wrap(rt.uncommentMessage))
	rt.router.GET("/users/:id", rt.wrap(rt.getUser)) // ✅ Add this route
	rt.router.GET("/messages/:message_id/comments", rt.wrap(rt.getComments))
	rt.router.GET("/search/users", rt.wrap(rt.searchUser))

	// Emoji reactions on messages
	// NOTE: httprouter stores its routing tree per HTTP method, and each method
	// has already fixed the wildcard name on the /conversations/ prefix:
	// PUT uses :c_id (set-group-photo), DELETE uses :conversation_id (deleteMessage).
	// We must match those names per method to avoid a wildcard-conflict panic.
	rt.router.PUT("/conversations/:c_id/messages/:message_id/reaction", rt.wrap(rt.setReaction))
	rt.router.DELETE("/conversations/:conversation_id/messages/:message_id/reaction", rt.wrap(rt.removeReaction))

	// rt.router.POST("/conversations/:c_id/messages", rt.wrap(rt.sendMessage))// Send message to an existing conversation
	// rt.router.GET("/users/:id/conversations/:c_id", rt.getConversation)
	// :conversation

	return rt.router
}

// use context in every(not dologin) api
// after I iterate over rows or use them  in database I need to
// if err = rows.Err(); err != nil {
// }

// маршруты тут написаны и зависимо от метода выбирает пути 
