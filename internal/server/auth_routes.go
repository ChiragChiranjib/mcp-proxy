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
		if deps.Logger != nil {
			deps.Logger.Info("AUTH_ME_INIT",
				"method", r.Method,
				"path", r.URL.Path,
				"user_id", GetUserID(r),
			)
		}
		uid := GetUserID(r)
		if uid == "" {
			if deps.Logger != nil {
				deps.Logger.Info("AUTH_ME_UNAUTHORIZED")
			}
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if deps.Logger != nil {
			deps.Logger.Info("AUTH_ME_OK", "user_id", uid)
		}
		WriteJSON(w, http.StatusOK, map[string]string{"user_id": uid})
	}).Methods(http.MethodGet)
	// exchange Google ID token for app session
	r.HandleFunc(cfg.AdminPrefix+"/auth/google", func(w http.ResponseWriter, r *http.Request) {
		if deps.Logger != nil {
			deps.Logger.Info("AUTH_GOOGLE_INIT",
				"method", r.Method,
				"path", r.URL.Path,
			)
		}
		var body struct {
			Credential string `json:"credential"`
		}
		if !ReadJSON(w, r, &body) {
			if deps.Logger != nil {
				deps.Logger.Error("AUTH_GOOGLE_READ_BODY_ERROR")
			}
			return
		}
		if body.Credential == "" {
			if deps.Logger != nil {
				deps.Logger.Error("AUTH_GOOGLE_MISSING_CREDENTIAL")
			}
			WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "missing credential"})
			return
		}
		// verify ID token
		clientID := ""
		if deps.AppConfig != nil {
			clientID = deps.AppConfig.Google.ClientID
		}
		payload, err := idtoken.Validate(r.Context(), body.Credential, clientID)
		if err != nil {
			if deps.Logger != nil {
				deps.Logger.Error("AUTH_GOOGLE_VALIDATE_ERROR", "error", err)
			}
			WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid google token"})
			return
		}
		email, _ := payload.Claims["email"].(string)
		name, _ := payload.Claims["name"].(string)
		if email == "" {
			if deps.Logger != nil {
				deps.Logger.Error("AUTH_GOOGLE_EMAIL_MISSING")
			}
			WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "email missing"})
			return
		}

		uid := ensureUser(r.Context(), deps, email)
		role := "user"
		adminID := ""
		if deps.AppConfig != nil {
			adminID = deps.AppConfig.Security.AdminUserID
		}
		if adminID != "" && (uid == adminID || email == adminID) {
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
		secret := ""
		if deps.AppConfig != nil {
			secret = deps.AppConfig.Security.JWTSecret
		}
		s, err := token.SignedString([]byte(secret))
		if err != nil {
			if deps.Logger != nil {
				deps.Logger.Error("AUTH_GOOGLE_TOKEN_ERROR", "error", err)
			}
			WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "token error"})
			return
		}
		// set cookie
		http.SetCookie(w, &http.Cookie{Name: "session", Value: s, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: false, MaxAge: 3600})
		if deps.Logger != nil {
			deps.Logger.Info("AUTH_GOOGLE_OK", "user_id", uid)
		}
		WriteJSON(w, http.StatusOK, map[string]any{"user_id": uid, "email": email, "name": name})
	}).Methods(http.MethodPost)

	// username/password basic login → sets JWT session cookie
	r.HandleFunc(cfg.AdminPrefix+"/auth/basic", func(w http.ResponseWriter, r *http.Request) {
		if deps.Logger != nil {
			deps.Logger.Info("AUTH_BASIC_INIT",
				"method", r.Method,
				"path", r.URL.Path,
			)
		}
		type req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		var b req
		if !ReadJSON(w, r, &b) {
			if deps.Logger != nil {
				deps.Logger.Error("AUTH_BASIC_READ_BODY_ERROR")
			}
			return
		}
		if b.Username == "" || b.Password == "" {
			if deps.Logger != nil {
				deps.Logger.Error("AUTH_BASIC_MISSING_CREDENTIALS")
			}
			WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "missing credentials"})
			return
		}

		if b.Username != "" {
			// compare with configured credentials stored in server deps via user service accessor
		}

		if !validateBasic(deps, b.Username, b.Password) {
			if deps.Logger != nil {
				deps.Logger.Error("AUTH_BASIC_INVALID_CREDENTIALS")
			}
			WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
			return
		}
		// ensure user exists and use canonical internal id
		uid := ensureUser(r.Context(), deps, b.Username)
		role := "user"
		adminID := ""
		if deps.AppConfig != nil {
			adminID = deps.AppConfig.Security.AdminUserID
		}
		if adminID != "" && (uid == adminID || b.Username == adminID) {
			role = "admin"
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"uid":   uid,
			"email": uid,
			"role":  role,
			"exp":   time.Now().Add(60 * time.Minute).Unix(),
			"iat":   time.Now().Unix(),
		})
		secret := ""
		if deps.AppConfig != nil {
			secret = deps.AppConfig.Security.JWTSecret
		}
		s, err := token.SignedString([]byte(secret))
		if err != nil {
			if deps.Logger != nil {
				deps.Logger.Error("AUTH_BASIC_TOKEN_ERROR", "error", err)
			}
			WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "token error"})
			return
		}
		http.SetCookie(w, &http.Cookie{Name: "session", Value: s, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: false, MaxAge: 3600})
		if deps.Logger != nil {
			deps.Logger.Info("AUTH_BASIC_OK", "user_id", uid)
		}
		WriteJSON(w, http.StatusOK, map[string]any{"user_id": uid, "email": uid})
	}).Methods(http.MethodPost)

	// logout
	r.HandleFunc(cfg.AdminPrefix+"/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		if deps.Logger != nil {
			deps.Logger.Info("AUTH_LOGOUT_INIT", "user_id", GetUserID(r))
		}
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "", Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: false, MaxAge: 0})
		w.WriteHeader(http.StatusNoContent)
		if deps.Logger != nil {
			deps.Logger.Info("AUTH_LOGOUT_OK")
		}
	}).Methods(http.MethodPost)

}

// validateBasic compares provided credentials to configured ones.
func validateBasic(deps Deps, username, password string) bool {
	if deps.AppConfig == nil || deps.AppConfig.Security.BasicUsername == "" {
		return false
	}
	return username == deps.AppConfig.Security.BasicUsername && password == deps.AppConfig.Security.BasicPassword
}

// ensureUser creates a user record if missing. For brevity, this is a placeholder where you would
// call the repo Users methods. Here we no-op to keep the example focused.
func ensureUser(ctx context.Context, deps Deps, email string) string {
	if deps.UserService != nil {
		if id, err := deps.UserService.GetOrCreateByEmail(ctx, email); err == nil && id != "" {
			return id
		}
	}
	return email
}
