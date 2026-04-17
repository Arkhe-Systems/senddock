package handler

import "regexp"

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func isValidEmail(email string) bool {
	return len(email) <= 255 && emailRegex.MatchString(email)
}

func isValidPassword(password string) bool {
	return len(password) >= 8
}
