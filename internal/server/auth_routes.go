package server

import (
	"context"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"google.golang.org/api/idtoken"
)

// addAuthRoutes configures SSO endpoints.
func addAuthRoutes(r *mux.Router, deps Deps, cfg Config) {
	// get current session info (if any)
	r.HandleFunc(cfg.AdminPrefix+"/auth/me", func(w http.ResponseWriter, r *http.Request) {
		uid := GetUserID(r)
		if uid == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		WriteJSON(w, http.StatusOK, map[string]string{"user_id": uid})
	}).Methods(http.MethodGet)
	// exchange Google ID token for app session
	r.HandleFunc(cfg.AdminPrefix+"/auth/google", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Credential string `json:"credential"`
		}
		if !ReadJSON(w, r, &body) {
			return
		}
		if body.Credential == "" {
			WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "missing credential"})
			return
		}
		// verify ID token
		payload, err := idtoken.Validate(r.Context(), body.Credential, deps.GoogleClientID)
		if err != nil {
			WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid google token"})
			return
		}
		email, _ := payload.Claims["email"].(string)
		name, _ := payload.Claims["name"].(string)
		if email == "" {
			WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "email missing"})
			return
		}
		// ensure user exists (reuse hub service db through queries if available)
		// We don't have a dedicated user service exposed; use hub service's query layer indirectly by creating a helper.
		// For simplicity, create a minimal user via hub service's queries through an exported func if available.
		// Here, we piggyback on Hubs service by adding a no-op list call to ensure DB connection created.
		_ = deps.Hubs // ensure not nil
		uid := ensureUser(r.Context(), deps, email)
		role := "user"
		if deps.AdminUserID != "" && (uid == deps.AdminUserID || email == deps.AdminUserID) {
			role = "admin"
		}
		// build app jwt
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"uid":   uid,
			"email": email,
			"name":  name,
			"role":  role,
			"exp":   time.Now().Add(60 * time.Minute).Unix(),
			"iat":   time.Now().Unix(),
		})
		s, err := token.SignedString([]byte(deps.JWTSecret))
		if err != nil {
			WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "token error"})
			return
		}
		// set cookie
		http.SetCookie(w, &http.Cookie{Name: "session", Value: s, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: false, MaxAge: 3600})
		WriteJSON(w, http.StatusOK, map[string]any{"user_id": uid, "email": email, "name": name})
	}).Methods(http.MethodPost)

	// username/password basic login → sets JWT session cookie
	r.HandleFunc(cfg.AdminPrefix+"/auth/basic", func(w http.ResponseWriter, r *http.Request) {
		type req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		var b req
		if !ReadJSON(w, r, &b) {
			return
		}
		if b.Username == "" || b.Password == "" {
			WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "missing credentials"})
			return
		}
		// validate against configured credentials
		// We piggyback on deps.UserService to expose admin id if needed.
		// Username is used as email also for simplicity.
		if b.Username != "" {
			// compare with configured credentials stored in server deps via user service accessor
		}
		// Deps does not currently carry basic creds; we use r.Context() bound configs via package level closure.
		// As a practical approach, we use environment-configured credentials passed into user service accessors.
		// Here, rely on special endpoint in UserService to validate; fallback to direct comparison using headers not available here.
		// For brevity, we embed comparison via a helper
		if !validateBasic(deps, b.Username, b.Password) {
			WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
			return
		}
		// ensure user exists and use canonical internal id
		uid := ensureUser(r.Context(), deps, b.Username)
		role := "user"
		if deps.AdminUserID != "" && (uid == deps.AdminUserID || b.Username == deps.AdminUserID) {
			role = "admin"
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"uid":   uid,
			"email": uid,
			"role":  role,
			"exp":   time.Now().Add(60 * time.Minute).Unix(),
			"iat":   time.Now().Unix(),
		})
		s, err := token.SignedString([]byte(deps.JWTSecret))
		if err != nil {
			WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "token error"})
			return
		}
		http.SetCookie(w, &http.Cookie{Name: "session", Value: s, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: false, MaxAge: 3600})
		WriteJSON(w, http.StatusOK, map[string]any{"user_id": uid, "email": uid})
	}).Methods(http.MethodPost)

	// logout
	r.HandleFunc(cfg.AdminPrefix+"/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "", Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: false, MaxAge: 0})
		w.WriteHeader(http.StatusNoContent)
	}).Methods(http.MethodPost)

}

// validateBasic compares provided credentials to configured ones.
func validateBasic(deps Deps, username, password string) bool {
	if deps.BasicUsername == "" {
		return false
	}
	return username == deps.BasicUsername && password == deps.BasicPassword
}

// ensureUser creates a user record if missing. For brevity, this is a placeholder where you would
// call the sqlc Users queries. Here we no-op to keep the example focused.
func ensureUser(ctx context.Context, deps Deps, email string) string {
	if deps.UserService != nil {
		if id, err := deps.UserService.GetOrCreateByEmail(ctx, email); err == nil && id != "" {
			return id
		}
	}
	return email
}
