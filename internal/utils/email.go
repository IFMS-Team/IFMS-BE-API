package utils

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
	"time"
)

func SendOTPEmail(toEmail, otp string) error {
	if os.Getenv("GMAIL_CLIENT_ID") != "" {
		return sendWithGmailAPI(toEmail, otp)
	}

	if os.Getenv("SMTP_HOST") != "" {
		return sendWithSMTP(toEmail, otp)
	}

	return fmt.Errorf("no email provider configured: set GMAIL_CLIENT_ID or SMTP_HOST")
}

func getGmailAccessToken() (string, error) {
	clientID := os.Getenv("GMAIL_CLIENT_ID")
	clientSecret := os.Getenv("GMAIL_CLIENT_SECRET")
	refreshToken := os.Getenv("GMAIL_REFRESH_TOKEN")

	data := url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"refresh_token": {refreshToken},
		"grant_type":    {"refresh_token"},
	}

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", data)
	if err != nil {
		return "", fmt.Errorf("gmail token request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("gmail token decode failed: %w", err)
	}

	if result.Error != "" {
		return "", fmt.Errorf("gmail oauth error: %s - %s", result.Error, result.ErrorDesc)
	}

	return result.AccessToken, nil
}

func sendWithGmailAPI(toEmail, otp string) error {
	accessToken, err := getGmailAccessToken()
	if err != nil {
		return err
	}

	from := os.Getenv("SMTP_FROM")
	if from == "" {
		from = os.Getenv("SMTP_USER")
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

	encodedMsg := base64.URLEncoding.EncodeToString([]byte(msg))

	payload, _ := json.Marshal(map[string]string{
		"raw": encodedMsg,
	})

	req, err := http.NewRequest("POST",
		"https://gmail.googleapis.com/gmail/v1/users/me/messages/send",
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return fmt.Errorf("gmail request failed: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("gmail send failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gmail error (%d): %s", resp.StatusCode, string(respBody))
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
