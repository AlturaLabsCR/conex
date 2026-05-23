package i18n

func init() {
	Locales["es"] = es
}

var es = map[string]string{
	"root.greeting":    "Hola, mundo!",
	"mail.otp.subject": "Tu código de verificación de Conex",
	"mail.otp.body":    "Tu código de verificación de Conex es {{.Code}}.\n\nVence a las {{.ExpiresAt}}.\n",
}
