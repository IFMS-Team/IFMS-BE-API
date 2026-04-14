package utils

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/smtp"
	"os"
	"time"
)

type resendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Text    string   `json:"text"`
}

func SendOTPEmail(toEmail, otp string) error {
	apiKey := os.Getenv("RESEND_API_KEY")

	if apiKey != "" {
		return sendWithResend(apiKey, toEmail, otp)
	}
	return sendWithSMTP(toEmail, otp)
}

func sendWithResend(apiKey, toEmail, otp string) error {
	from := os.Getenv("SMTP_FROM")
	if from == "" {
		from = "IFMS <onboarding@resend.dev>"
	}

	body := fmt.Sprintf(
		"Your OTP code for password reset is: %s\n\nThis code will expire in 5 minutes.\nIf you did not request this, please ignore this email.",
		otp,
	)

	payload := resendRequest{
		From:    from,
		To:      []string{toEmail},
		Subject: "IFMS - Password Reset OTP",
		Text:    body,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("email marshal failed: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("email request failed: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("email send failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend error (%d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func sendWithSMTP(toEmail, otp string) error {
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

	addr := fmt.Sprintf("%s:%s", host, port)

	if port == "465" {
		return sendWithSSL(addr, host, user, password, from, toEmail, msg)
	}
	return sendWithSTARTTLS(addr, host, user, password, from, toEmail, msg)
}

func sendWithSSL(addr, host, user, password, from, to, msg string) error {
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{ServerName: host})
	if err != nil {
		return fmt.Errorf("smtp ssl connect failed: %w", err)
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("smtp client failed: %w", err)
	}
	defer client.Close()

	auth := smtp.PlainAuth("", user, password, host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth failed: %w", err)
	}

	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte(msg)); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	return client.Quit()
}

func sendWithSTARTTLS(addr, host, user, password, from, to, msg string) error {
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("smtp connect failed: %w", err)
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("smtp client failed: %w", err)
	}
	defer client.Close()

	if err := client.StartTLS(&tls.Config{ServerName: host}); err != nil {
		return fmt.Errorf("smtp starttls failed: %w", err)
	}

	auth := smtp.PlainAuth("", user, password, host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth failed: %w", err)
	}

	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte(msg)); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	return client.Quit()
}
