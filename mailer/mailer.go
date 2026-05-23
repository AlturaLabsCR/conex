package mailer

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	gosmtp "github.com/tavocg/go-email/smtp"
)

type Mailer struct {
	opts Options
}

type Options struct {
	From         string
	SMTPAddress  string
	SMTPUser     string
	SMTPPassword string
	SMTPStartTLS bool
}

func NewMailer(opts Options) (*Mailer, error) {
	address := strings.TrimSpace(opts.SMTPAddress)
	if address == "" {
		return nil, errors.New("mailer smtp address is required")
	}

	return &Mailer{
		opts: Options{
			From:         strings.TrimSpace(opts.From),
			SMTPAddress:  address,
			SMTPUser:     strings.TrimSpace(opts.SMTPUser),
			SMTPPassword: opts.SMTPPassword,
			SMTPStartTLS: opts.SMTPStartTLS,
		},
	}, nil
}

func (m *Mailer) SendOTP(ctx context.Context, L func(string, ...any) string, recipient string, otp int64, expiresAt int64) error {
	client, err := m.newClient(ctx)
	if err != nil {
		return err
	}

	code := fmt.Sprintf("%06d", otp)
	expiresAtTime := time.Unix(expiresAt, 0).UTC()
	message := gosmtp.NewPlainMessage(
		m.opts.From,
		[]string{recipient},
		L("mail.otp.subject"),
		L("mail.otp.body", map[string]any{
			"Code":      code,
			"ExpiresAt": expiresAtTime.Format(time.RFC1123),
		}),
	)

	if err := client.Send(ctx, message); err != nil {
		return fmt.Errorf("send otp email: %w", err)
	}

	return nil
}

func (m *Mailer) newClient(ctx context.Context) (*gosmtp.Client, error) {
	smtpOpts := []gosmtp.Option{}
	if m.opts.SMTPStartTLS {
		smtpOpts = append(smtpOpts, gosmtp.WithStartTLS())
	}

	client, err := gosmtp.NewClient(
		ctx,
		m.opts.SMTPAddress,
		m.opts.SMTPUser,
		m.opts.SMTPPassword,
		smtpOpts...,
	)
	if err != nil {
		return nil, fmt.Errorf("create smtp client: %w", err)
	}

	return client, nil
}
