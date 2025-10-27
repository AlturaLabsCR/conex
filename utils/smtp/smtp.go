// Package smtp
package smtp

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"mime/multipart"
	"net/smtp"
	"net/textproto"
	"strconv"
	"strings"
)

type Auth struct {
	params AuthParams
}

type AuthParams struct {
	Host string
	Port string
	Pass string
}

type Attachment struct {
	Filename string
	Bytes    *bytes.Buffer
}

func Client(params AuthParams) *Auth {
	return &Auth{params}
}

func (s *Auth) tlsConfig() *tls.Config {
	return &tls.Config{ServerName: s.params.Host}
}

func (s *Auth) smtpClient() (*smtp.Client, error) {
	address := s.params.Host + ":" + s.params.Port
	portNum, err := strconv.Atoi(s.params.Port)
	if err != nil {
		return nil, fmt.Errorf("invalid port number: %w", err)
	}

	if portNum == 465 {
		conn, err := tls.Dial("tcp", address, s.tlsConfig())
		if err != nil {
			return nil, fmt.Errorf("failed to connect to SMTPS server: %w", err)
		}
		return smtp.NewClient(conn, s.params.Host)
	}

	client, err := smtp.Dial(address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SMTP server: %w", err)
	}

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(s.tlsConfig()); err != nil {
			client.Close()
			return nil, fmt.Errorf("failed to start TLS: %w", err)
		}
	} else {
		client.Close()
		return nil, fmt.Errorf("TLS not supported by the server")
	}

	return client, nil
}

func (s *Auth) SendText(
	from string,
	to []string,
	subject string,
	body string,
	attachments ...[]Attachment,
) error {
	return s.sendMessage(from, to, subject, body, "text/plain; charset=utf-8", attachments...)
}

func (s *Auth) SendHTML(
	from string,
	to []string,
	subject string,
	html string,
	attachments ...[]Attachment,
) error {
	return s.sendMessage(from, to, subject, html, "text/html; charset=utf-8", attachments...)
}

func (s *Auth) sendMessage(
	from string,
	to []string,
	subject string,
	body string,
	contentType string,
	attachments ...[]Attachment,
) error {
	var att []Attachment
	if len(attachments) > 0 && attachments[0] != nil {
		att = attachments[0]
	}

	message, err := s.buildMessage(from, to, subject, body, contentType, att)
	if err != nil {
		return fmt.Errorf("failed to create email message: %w", err)
	}

	client, err := s.smtpClient()
	if err != nil {
		return err
	}
	defer client.Close()

	auth := smtp.PlainAuth("", from, s.params.Pass, s.params.Host)
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	if err = client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}
	for _, recipient := range to {
		if err = client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to add recipient %s: %w", recipient, err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}
	if _, err = w.Write(message); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return client.Quit()
}

func (s *Auth) Validate(user string) error {
	client, err := s.smtpClient()
	if err != nil {
		return err
	}
	defer client.Close()

	auth := smtp.PlainAuth("", user, s.params.Pass, s.params.Host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	return nil
}

func (s *Auth) buildMessage(
	from string,
	to []string,
	subject string,
	body string,
	contentType string,
	attachments []Attachment,
) ([]byte, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	buf.WriteString(fmt.Sprintf("From: %s\r\n", from))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ",")))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n\r\n", writer.Boundary()))

	if err := s.write(writer, contentType, body); err != nil {
		return nil, fmt.Errorf("failed to write email body: %w", err)
	}

	for _, a := range attachments {
		if err := s.attach(writer, a); err != nil {
			return nil, fmt.Errorf("failed to attach file %s: %w", a.Filename, err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	return buf.Bytes(), nil
}

func (s *Auth) write(writer *multipart.Writer, contentType, content string) error {
	partHeader := make(textproto.MIMEHeader)
	partHeader.Set("Content-Type", contentType)
	part, err := writer.CreatePart(partHeader)
	if err != nil {
		return fmt.Errorf("failed to create part: %w", err)
	}

	_, err = part.Write([]byte(content))
	return err
}

func (s *Auth) attach(writer *multipart.Writer, a Attachment) error {
	header := make(textproto.MIMEHeader)
	header.Set("Content-Type", "application/octet-stream")
	header.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, a.Filename))

	part, err := writer.CreatePart(header)
	if err != nil {
		return fmt.Errorf("failed to create attachment part: %w", err)
	}

	_, err = part.Write(a.Bytes.Bytes())
	return err
}
