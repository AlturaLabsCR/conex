package config

import "app/utils/smtp"

func InitSMTPAuth() smtp.AuthParams {
	return smtp.AuthParams{
		Host: ServerSMTPHost,
		Port: ServerSMTPPort,
		Pass: ServerSMTPPass,
	}
}
