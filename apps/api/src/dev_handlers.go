package main

import (
	"api/src/internal/models"
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/valkey-io/valkey-go"
	"gorm.io/gorm"
)

// devLoginHandler handles POST /dev/login.
// It finds or creates a dev user by email (default: dev@restorate.local),
// creates a Valkey session, and sets the session_token cookie.
// The caller should follow up with GetCurrentUser to hydrate the auth state.
func devLoginHandler(db *gorm.DB, kv valkey.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		email := r.URL.Query().Get("email")
		if email == "" {
			email = "dev@restorate.local"
		}

		ctx := context.Background()

		var user models.User
		err := db.WithContext(ctx).Where("email = ?", email).First(&user).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user = models.User{
				Email: models.StringPtr(email),
				Name:  "Dev User",
			}
			if createErr := db.WithContext(ctx).Create(&user).Error; createErr != nil {
				http.Error(w, "failed to create user", http.StatusInternalServerError)
				return
			}
		} else if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}

		token := uuid.New().String()
		setCmd := kv.B().Set().Key("session:" + token).Value(user.ID).Ex(24 * time.Hour).Build()
		if kv.Do(ctx, setCmd).Error() != nil {
			http.Error(w, "cache error", http.StatusInternalServerError)
			return
		}

		// SameSite=None is not set; dev never needs Secure
		cookie := "session_token=" + token + "; HttpOnly; Path=/; Max-Age=" + strconv.Itoa(86400) + "; SameSite=Lax"
		w.Header().Set("Set-Cookie", cookie)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	})
}
