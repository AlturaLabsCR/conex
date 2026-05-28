package mailer

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	emailmailer "github.com/tavocg/go-email/mailer"
	"github.com/tavocg/go-email/mailer/backends/mailgun"
)

type Mailer struct {
	from   string
	client emailmailer.Mailer
}

type Options struct {
	From           string
	MailgunDomain  string
	MailgunAPIKey  string
	MailgunAPIBase string
}

func NewMailer(opts Options) (*Mailer, error) {
	from := strings.TrimSpace(opts.From)
	domain := strings.TrimSpace(opts.MailgunDomain)
	apiKey := strings.TrimSpace(opts.MailgunAPIKey)
	apiBase := strings.TrimSpace(opts.MailgunAPIBase)

	if from == "" {
		return nil, errors.New("mailer from address is required")
	}
	if domain == "" {
		return nil, errors.New("mailer mailgun domain is required")
	}
	if apiKey == "" {
		return nil, errors.New("mailer mailgun api key is required")
	}

	client, err := mailgun.NewClient(domain, apiKey, mailgun.WithAPIBase(apiBase))
	if err != nil {
		return nil, fmt.Errorf("create mailgun client: %w", err)
	}

	return &Mailer{from: from, client: client}, nil
}

func (m *Mailer) SendOTP(ctx context.Context, L func(string, ...any) string, recipient string, otp int64, expiresAt int64) error {
	code := fmt.Sprintf("%06d", otp)
	expiresAtTime := time.Unix(expiresAt, 0).UTC()
	message := emailmailer.NewPlainMessage(
		m.from,
		[]string{recipient},
		L("mail.otp.subject"),
		L("mail.otp.body", map[string]any{
			"Code":      code,
			"ExpiresAt": expiresAtTime.Format(time.RFC1123),
		}),
	)

	if err := m.client.Send(ctx, message); err != nil {
		return fmt.Errorf("send otp email: %w", err)
	}

	return nil
}
