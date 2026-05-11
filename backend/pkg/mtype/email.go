package mtype

import (
	"regexp"
	"strings"
)

var emailRegexp = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

type Email string

func NewEmail(email string) (Email, bool) {
	if !emailRegexp.MatchString(email) {
		return "", false
	}
	email = strings.ToLower(email)

	return Email(email), true
}

func (e Email) String() string {
	return string(e)
}

func (e Email) IsValid() bool {
	return emailRegexp.MatchString(e.String())
}
