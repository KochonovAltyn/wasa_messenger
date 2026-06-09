package api

import (
	"database/sql" // Import the database/sql package
	"errors"       // Import the errors package
	"net/http"
	"strings" // Import the strings package

	uuid "github.com/KochonovAltyn/wasa_messenger/service/uid"
	"github.com/julienschmidt/httprouter"
	logrus "github.com/KochonovAltyn/wasa_messenger/service/applog"
	"github.com/KochonovAltyn/wasa_messenger/service/api/reqcontext"
)

// httpRouterHandler is the signature for functions that accepts a reqcontext.RequestContext in addition to those
// required by the httprouter package.
type httpRouterHandler func(http.ResponseWriter, *http.Request, httprouter.Params, *reqcontext.RequestContext)

// wrap parses the request and adds a reqcontext.RequestContext instance related to the request.
// wrap parses the request and adds a reqcontext.RequestContext instance related to the request.
// wrap parses the request and adds a reqcontext.RequestContext instance related to the request.
func (rt *_router) wrap(fn httpRouterHandler) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// Log the request
		rt.baseLogger.Infof("Request received: method=%s, path=%s", r.Method, r.URL.Path)

		reqUUID, err := uuid.NewV4()
		if err != nil {
			rt.baseLogger.WithError(err).Error("can't generate a request UUID")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Skip authentication for specific endpoints (e.g., doLogin)
		if r.URL.Path == "/session" && r.Method == http.MethodPost {
			// Create a request-specific logger without user ID
			ctx := &reqcontext.RequestContext{
				ReqUUID: reqUUID,
				Logger: rt.baseLogger.WithFields(logrus.Fields{
					"reqid":     reqUUID.String(),
					"remote-ip": r.RemoteAddr,
				}),
			}

			// Call the next handler in chain
			fn(w, r, ps, ctx)
			return
		}

		// For all other endpoints, enforce authentication
		userID, err := rt.extractUserIDFromHeader(r)
		if err != nil {
			rt.baseLogger.WithError(err).Warn("authentication failed")
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// Validate the user ID (token) by checking if the user exists in the database
		_, err = rt.db.GetUserId(userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				rt.baseLogger.WithError(err).Warn("user not found")
				http.Error(w, "Invalid user ID (token)", http.StatusUnauthorized)
				return
			}
			rt.baseLogger.WithError(err).Error("error fetching user")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Create a request-specific logger with user ID
		ctx := &reqcontext.RequestContext{
			ReqUUID: reqUUID,
			UserID:  userID,
			Logger: rt.baseLogger.WithFields(logrus.Fields{
				"reqid":     reqUUID.String(),
				"remote-ip": r.RemoteAddr,
				"user":      userID,
			}),
		}

		// Call the next handler in
		fn(w, r, ps, ctx)
	}
}

// extractUserIDFromHeader extracts the user ID (token) from the Authorization header
func (rt *_router) extractUserIDFromHeader(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("unauthorized: missing authorization header")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("unauthorized: invalid authorization format")
	}

	return parts[1], nil // ✅ Return token as user ID
}
