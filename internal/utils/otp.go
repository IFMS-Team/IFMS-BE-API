package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
)

var (
	hasLower   = regexp.MustCompile(`[a-z]`)
	hasUpper   = regexp.MustCompile(`[A-Z]`)
	hasDigit   = regexp.MustCompile(`[0-9]`)
	hasSpecial = regexp.MustCompile(`[@$!%*?&]`)
)

func GenerateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func ValidatePasswordStrength(password string) (bool, string) {
	if len(password) < 8 {
		return false, "Password must be at least 8 characters long"
	}
	if len(password) > 64 {
		return false, "Password must be at most 64 characters long"
	}
	if !hasLower.MatchString(password) {
		return false, "Password must contain at least one lowercase letter"
	}
	if !hasUpper.MatchString(password) {
		return false, "Password must contain at least one uppercase letter"
	}
	if !hasDigit.MatchString(password) {
		return false, "Password must contain at least one number"
	}
	if !hasSpecial.MatchString(password) {
		return false, "Password must contain at least one special character (@$!%*?&)"
	}
	return true, ""
}
