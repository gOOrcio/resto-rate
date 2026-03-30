package services

import "api/src/internal/models"

// missingRestaurantFields returns a map of field updates for a restaurant that is missing
// city, country, or photo_reference but the caller has values. Used to backfill old records.
func missingRestaurantFields(r *models.Restaurant, city, country, photoRef string) map[string]any {
	updates := map[string]any{}
	if r.City == "" && city != "" {
		updates["city"] = city
	}
	if r.Country == "" && country != "" {
		updates["country"] = country
	}
	if r.PhotoReference == "" && photoRef != "" {
		updates["photo_reference"] = photoRef
	}
	return updates
}

// Error message constants — shared across services to avoid S1192 duplicate literals.
const (
	errDatabaseNotInitialized = "database not initialized"
	errGooglePlacesIDRequired = "google_places_id is required"
	errUserNotFound           = "user not found"
	errReviewNotFound         = "review not found"
	errIDRequired             = "id is required"
	errRequestIDRequired      = "request_id is required"
	errFriendUserIDRequired   = "friend_user_id is required"
	errUsernameRequired       = "username is required"
	errInvalidUsername        = "invalid username"
)
