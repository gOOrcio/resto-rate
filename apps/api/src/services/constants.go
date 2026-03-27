package services

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
