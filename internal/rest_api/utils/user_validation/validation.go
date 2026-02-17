package user_validation

import (
	"errors"
	"net/mail"
	"regexp"
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_.]{3,30}$`)

func ValidatePassword(p string) error {
	if len(p) < 8 {
		return errors.New("password too short")
	}
	if len(p) > 72 {
		return errors.New("password too long")
	}
	return nil
}

func ValidateEmail(email string) error {
	if len(email) > 254 {
		return errors.New("email too long")
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		return errors.New("invalid email format")
	}

	return nil
}

func ValidateUsername(username string) error {
	if !usernameRegex.MatchString(username) {
		return errors.New("invalid username format")
	}

	return nil
}
