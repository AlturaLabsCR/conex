package i18n

func init() {
	Locales["en"] = en
}

var en = map[string]string{
	"root.greeting":    "Hello, world!",
	"mail.otp.subject": "Your Conex verification code",
	"mail.otp.body":    "Your Conex verification code is {{.Code}}.\n\nIt expires at {{.ExpiresAt}}.\n",
	"plans.free.name":  "Free",
}
