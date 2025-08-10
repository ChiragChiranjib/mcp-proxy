package server

import (
	"net/http"
	"time"

	ck "github.com/ChiragChiranjib/mcp-proxy/internal/contextkey"
	m "github.com/ChiragChiranjib/mcp-proxy/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"google.golang.org/api/idtoken"
)

// addAuthRoutes configures SSO endpoints.
func addAuthRoutes(r *mux.Router, deps Deps, cfg Config) {
	// get current session info (if any)
	r.HandleFunc(cfg.AdminPrefix+"/auth/me",
		func(w http.ResponseWriter, r *http.Request) {
			deps.Logger.Info("AUTH_ME_INIT",
				"method", r.Method,
				"path", r.URL.Path,
				"user_id", ck.GetUserIDFromContext(r.Context()),
			)

			uid := ck.GetUserIDFromContext(r.Context())
			if uid == "" {
				deps.Logger.Info("AUTH_ME_UNAUTHORIZED")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			deps.Logger.Info("AUTH_ME_SUCCESS", "user_id", uid)
			u, err := deps.UserService.FindUserByID(r.Context(), uid)
			if err != nil {
				deps.Logger.Info("AUTH_USER_NOT_FOUND", "user_id", uid)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			WriteJSON(w, http.StatusOK,
				map[string]string{"user_id": uid, "email": u.Username})
		}).Methods(http.MethodGet)
	// exchange Google ID token for app session
	r.HandleFunc(cfg.AdminPrefix+"/auth/google",
		func(w http.ResponseWriter, r *http.Request) {
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
				deps.Logger.Error("AUTH_GOOGLE_READ_BODY_ERROR")
				return
			}
			if body.Credential == "" {
				deps.Logger.Error("AUTH_GOOGLE_MISSING_CREDENTIAL")
				WriteJSON(w, http.StatusBadRequest,
					map[string]string{"error": "missing credential"})
				return
			}
			// verify ID token
			clientID := ""
			if deps.AppConfig != nil {
				clientID = deps.AppConfig.Google.ClientID
			}
			payload, err := idtoken.Validate(r.Context(), body.Credential, clientID)
			if err != nil {
				deps.Logger.Error("AUTH_GOOGLE_VALIDATE_ERROR", "error", err)
				WriteJSON(w, http.StatusUnauthorized,
					map[string]string{"error": "invalid google token"})
				return
			}
			email, _ := payload.Claims["email"].(string)
			name, _ := payload.Claims["name"].(string)
			if email == "" {
				deps.Logger.Error("AUTH_GOOGLE_EMAIL_MISSING")
				WriteJSON(w, http.StatusUnauthorized,
					map[string]string{"error": "email missing"})
				return
			}

			user, err := deps.UserService.FetchOrCreateByUsername(r.Context(), email)
			if err != nil {
				deps.Logger.Error("AUTH_GOOGLE_TOKEN_ERROR", "error", err)
				WriteJSON(w, http.StatusInternalServerError,
					map[string]string{"error": "authorization failed."})
				return
			}

			// build JWT Token
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"email":    email,
				"uid":      user.ID,
				"username": user.Username,
				"role":     user.Role,
				"auth":     "sso",
				"exp":      time.Now().Add(120 * time.Minute).Unix(),
				"iat":      time.Now().Unix(),
			})

			s, err := token.SignedString([]byte(deps.AppConfig.Security.JWTSecret))
			if err != nil {
				deps.Logger.Error("AUTH_GOOGLE_TOKEN_ERROR", "error", err)
				WriteJSON(w, http.StatusInternalServerError,
					map[string]string{"error": "token error"})
				return
			}

			// set cookie
			http.SetCookie(w, &http.Cookie{
				Name:     "session",
				Value:    s,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				Secure:   false,
				MaxAge:   3600,
			})

			deps.Logger.Info("AUTH_GOOGLE_SUCCESS", "user_id", user.ID)
			WriteJSON(w, http.StatusOK,
				map[string]any{"user_id": user.ID, "email": email, "name": name})
		}).Methods(http.MethodPost)

	// username/password basic login → sets JWT session cookie
	r.HandleFunc(cfg.AdminPrefix+"/auth/basic",
		func(w http.ResponseWriter, r *http.Request) {
			deps.Logger.Info("AUTH_BASIC_INIT", "method", r.Method, "path", r.URL.Path)

			var req struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}
			if !ReadJSON(w, r, &req) {
				deps.Logger.Error("AUTH_BASIC_READ_BODY_ERROR")
				return
			}
			if req.Username == "" || req.Password == "" {
				deps.Logger.Error("AUTH_BASIC_MISSING_CREDENTIALS")
				WriteJSON(w, http.StatusBadRequest,
					map[string]string{"error": "missing credentials"})
				return
			}

			if req.Username != deps.AppConfig.Security.BasicUsername ||
				req.Password != deps.AppConfig.Security.BasicPassword {
				deps.Logger.Error("AUTH_BASIC_INVALID_CREDENTIALS", "req", req)
				WriteJSON(w, http.StatusUnauthorized,
					map[string]string{"error": "invalid credentials"})
				return
			}

			userEntity, err := deps.UserService.FindUserByUserName(
				r.Context(), req.Username)
			if err != nil || userEntity.Role != string(m.RoleAdmin) {
				WriteJSON(w, http.StatusUnauthorized,
					map[string]string{"error": "invalid credentials"})
				http.Error(w, "unauthorized user", http.StatusUnauthorized)
				return
			}

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"email":    userEntity.Username,
				"uid":      userEntity.ID,
				"username": userEntity.Username,
				"role":     userEntity.Role,
				"auth":     "basic",
				"exp":      time.Now().Add(120 * time.Minute).Unix(),
				"iat":      time.Now().Unix(),
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
				WriteJSON(w, http.StatusInternalServerError,
					map[string]string{"error": "token error"})
				return
			}
			http.SetCookie(w, &http.Cookie{
				Name:     "session",
				Value:    s,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				Secure:   false,
				MaxAge:   3600,
			})
			deps.Logger.Info("AUTH_BASIC_SUCCESS", "user_id", userEntity.ID)
			WriteJSON(w, http.StatusOK,
				map[string]any{"user_id": userEntity.ID, "username": userEntity.Username})
		}).Methods(http.MethodPost)

	// logout
	r.HandleFunc(cfg.AdminPrefix+"/auth/logout",
		func(w http.ResponseWriter, r *http.Request) {
			deps.Logger.Info("AUTH_LOGOUT_INIT",
				"user_id", ck.GetUserIDFromContext(r.Context()))
			http.SetCookie(w, &http.Cookie{
				Name:     "session",
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				Secure:   false,
				MaxAge:   0,
			})

			w.WriteHeader(http.StatusNoContent)
			deps.Logger.Info("AUTH_LOGOUT_SUCCESS",
				"user_id", ck.GetUserIDFromContext(r.Context()))
		}).Methods(http.MethodPost)
}
