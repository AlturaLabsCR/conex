package mailer

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tavocg/go-email/smtp"
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
	from := strings.TrimSpace(opts.From)
	address := strings.TrimSpace(opts.SMTPAddress)
	user := strings.TrimSpace(opts.SMTPUser)
	password := strings.TrimSpace(opts.SMTPPassword)

	if from == "" {
		return nil, errors.New("mailer from address is required")
	}
	if address == "" {
		return nil, errors.New("mailer smtp address is required")
	}
	if user == "" {
		return nil, errors.New("mailer smtp user is required")
	}
	if password == "" {
		return nil, errors.New("mailer smtp password is required")
	}

	return &Mailer{
		opts: Options{
			From:         from,
			SMTPAddress:  address,
			SMTPUser:     user,
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
	message := smtp.NewPlainMessage(
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

func (m *Mailer) newClient(ctx context.Context) (*smtp.Client, error) {
	smtpOpts := []smtp.Option{}
	if m.opts.SMTPStartTLS {
		smtpOpts = append(smtpOpts, smtp.WithStartTLS())
	}

	client, err := smtp.NewClient(
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
