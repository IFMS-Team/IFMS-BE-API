package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendOTPEmail(toEmail, otp string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASSWORD")
	from := os.Getenv("SMTP_FROM")

	if from == "" {
		from = user
	}

	subject := "IFMS - Password Reset OTP"
	body := fmt.Sprintf(
		"Your OTP code for password reset is: %s\n\nThis code will expire in 5 minutes.\nIf you did not request this, please ignore this email.",
		otp,
	)

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\n%s",
		from, toEmail, subject, body,
	)

	auth := smtp.PlainAuth("", user, password, host)
	addr := fmt.Sprintf("%s:%s", host, port)

	return smtp.SendMail(addr, auth, from, []string{toEmail}, []byte(msg))
}
