package auth

type contextKey string

const (
	UserIDKey    contextKey = "userID"
	ProjectIDKey contextKey = "projectID"
)
