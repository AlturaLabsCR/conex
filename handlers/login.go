package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"app/config"
	"app/internal/db"
	"app/templates"

	"golang.org/x/crypto/bcrypt"
)

type key struct {
	hashedEmail string
	otp         string
	expires     time.Time
}

var keys = map[string]key{}

var csrfProtection = map[int64]string{}

func (h *Handler) RegisterForm(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	header := templates.RegisterHeader(h.Translator(r))
	content := templates.RegisterSendEmail(h.Translator(r))

	templates.Base(h.Translator(r), header, content).Render(ctx, w)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	email := r.FormValue("email")
	if _, err := mail.ParseAddress(email); err != nil {
		templates.RegisterWarnInvalidEmail(h.Translator(r)).Render(ctx, w)
		h.Log().Debug("failed to parse email address", "email", email)
		return
	}

	queries := db.New(h.DB())

	if _, err := queries.GetUserByEmail(ctx, email); err == nil {
		templates.RegisterWarnInvalidEmail(h.Translator(r)).Render(ctx, w)
		h.Log().Debug("email exists", "email", email)
		return
	}

	token, err := h.issueOTP(h.Translator(r), email)
	if err != nil {
		h.Log().Error("error issuing otp", "error", err)
		templates.RegisterError(h.Translator(r)).Render(ctx, w)
		return
	}

	templates.RegisterConfirmEmail(h.Translator(r), token, email).Render(ctx, w)
}

func (h *Handler) RegisterConfirm(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	token := r.FormValue("token")
	key, ok := keys[token]
	if !ok {
		h.Log().Debug("invalid token", "token", token)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	if _, err := mail.ParseAddress(email); err != nil {
		h.Log().Debug("failed to parse email address", "email", email)
		templates.LoginWarnInvalidEmail(h.Translator(r)).Render(ctx, w)
		return
	}

	otp := r.FormValue("otp")

	if err := h.verifyOTP(key, email, token, otp); err != nil {
		templates.RegisterErrorInvalidOTP(h.Translator(r), token, email).Render(ctx, w)
		h.Log().Debug("failed to verify otp", "error", err)
		return
	}

	now := time.Now().Unix()
	queries := db.New(h.DB())
	if _, err := queries.InsertUser(ctx, db.InsertUserParams{
		UserEmail:        email,
		UserName:         "",
		UserCreatedUnix:  now,
		UserModifiedUnix: now,
		UserDeleted:      0,
	}); err != nil {
		templates.RegisterError(h.Translator(r)).Render(ctx, w)
		h.Log().Error("error registering user", "error", err)
		return
	}

	h.Log().Info("new user registration")

	if err := h.loginClient(w, r, email); err != nil {
		h.Log().Debug("failed to login user", "error", err)
	}

	templates.Redirect(config.DashboardPath).Render(ctx, w)
}

func (h *Handler) LoginForm(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := h.verifyClient(w, r)
	if err == nil {
		http.Redirect(w, r, config.DashboardPath, http.StatusSeeOther)
		return
	}

	header := templates.LoginHeader(h.Translator(r))
	content := templates.Login(h.Translator(r))

	templates.Base(h.Translator(r), header, content).Render(ctx, w)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	email := r.FormValue("email")
	if _, err := mail.ParseAddress(email); err != nil {
		templates.LoginWarnInvalidEmail(h.Translator(r)).Render(ctx, w)
		h.Log().Debug("failed to parse email address", "email", email)
		return
	}

	queries := db.New(h.DB())

	if _, err := queries.GetUserByEmail(ctx, email); err != nil {
		templates.LoginWarnInvalidEmail(h.Translator(r)).Render(ctx, w)
		h.Log().Debug("email does not exist", "email", email)
		return
	}

	token, err := h.issueOTP(h.Translator(r), email)
	if err != nil {
		h.Log().Error("error issuing otp", "error", err)
		templates.LoginError(h.Translator(r)).Render(ctx, w)
		return
	}

	templates.LoginConfirmEmail(h.Translator(r), token, email).Render(ctx, w)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	h.Sessions.JWTTerminate(w, r)

	http.SetCookie(w, &http.Cookie{
		Name:     "csrf",
		Value:    "",
		Path:     config.RootPrefix,
		Expires:  time.Unix(0, 0),
		HttpOnly: false,
		Secure:   r.TLS != nil,
	})

	http.Redirect(w, r, config.LoginPath, http.StatusSeeOther)
}

func (h *Handler) LoginConfirm(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	token := r.FormValue("token")
	key, ok := keys[token]
	if !ok {
		h.Log().Debug("invalid token", "token", token)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	if _, err := mail.ParseAddress(email); err != nil {
		h.Log().Debug("failed to parse email address", "email", email)
		templates.LoginWarnInvalidEmail(h.Translator(r)).Render(ctx, w)
		return
	}

	otp := r.FormValue("otp")

	if err := h.verifyOTP(key, email, token, otp); err != nil {
		templates.LoginErrorInvalidOTP(h.Translator(r), token, email).Render(ctx, w)
		h.Log().Debug("failed to verify otp", "error", err)
		return
	}

	if err := h.loginClient(w, r, email); err != nil {
		h.Log().Debug("failed to login user", "error", err)
	}

	templates.Redirect(config.DashboardPath).Render(ctx, w)
}

func (h *Handler) loginClient(w http.ResponseWriter, r *http.Request, email string) error {
	queries := db.New(h.DB())

	user, err := queries.GetUserByEmail(r.Context(), email)
	if err != nil {
		h.Log().Debug("failed to get user by email", "error", err)
		return err
	}

	now := time.Now().Unix()

	sessionID, err := queries.InsertSession(r.Context(), db.InsertSessionParams{
		SessionUser:        user.UserID,
		SessionOs:          getClientOS(r),
		SessionCreatedUnix: now,
	})
	if err != nil {
		h.Log().Debug("error inserting session", "error", err)
		return err
	}

	_, err = h.Sessions.JWTSet(w, r, db.Session{
		SessionID:          sessionID,
		SessionUser:        user.UserID,
		SessionOs:          getClientOS(r),
		SessionCreatedUnix: now,
	})
	if err != nil {
		h.Log().Debug("error setting jwt", "error", err)
		return err
	}

	return nil
}

func (h *Handler) verifyClient(w http.ResponseWriter, r *http.Request) error {
	session, expired, err := h.Sessions.JWTValidate(r)
	if err != nil && !expired {
		h.Log().Debug("error validating session", "error", err)
		return err
	}

	now := time.Now().Unix()

	if expired {
		_, err = h.Sessions.JWTSet(w, r, db.Session{
			SessionID:          session.SessionID,
			SessionUser:        session.SessionUser,
			SessionOs:          session.SessionOs,
			SessionCreatedUnix: now,
		})
		if err != nil {
			h.Log().Debug("error setting jwt", "error", err)
			return err
		}
	}

	csrfToken, err := randStr()
	if err != nil {
		h.Log().Error("error generating csrf token", "error", err)
		return err
	}

	csrfProtection[session.SessionID] = csrfToken

	http.SetCookie(w, &http.Cookie{
		Name:     "csrf",
		Value:    csrfToken,
		Path:     config.RootPrefix,
		Expires:  time.Now().Add(8 * time.Hour),
		HttpOnly: false,
		Secure:   r.TLS != nil,
	})

	return nil
}

func (h *Handler) issueOTP(tr func(string) string, email string) (string, error) {
	hashedEmailBytes, err := bcrypt.GenerateFromPassword([]byte(email), bcrypt.DefaultCost)
	hashedEmail := string(hashedEmailBytes)
	if err != nil {
		return "", err
	}

	otp, err := otp()
	if err != nil {
		return "", err
	}

	key := key{
		hashedEmail: hashedEmail,
		otp:         otp,
		expires:     time.Now().Add(5 * time.Minute),
	}

	token, err := randStr()
	if err != nil {
		return "", err
	}

	keys[token] = key

	h.Log().Debug("otp issued", "otp", otp)

	subject := tr("otp_code_email_subject")
	body := tr("otp_code_email_body") + " " + otp

	if h.Prod() {
		h.SMTPClient().SendText(
			config.ServerSMTPUser,
			[]string{email},
			subject,
			body,
		)
	} else {
		h.Log().Debug(
			"sent otp email",
			"from", config.ServerSMTPUser,
			"to", email,
			"subject", subject,
			"body", body,
		)
	}

	return token, nil
}

func (h *Handler) verifyOTP(key key, email, token, otp string) error {
	if time.Now().After(key.expires) {
		delete(keys, token)
		h.Log().Debug("token expired", "token", token)
		return fmt.Errorf("key expired")
	}

	if key.otp != otp {
		h.Log().Debug("invalid otp", "expected", key.otp, "got", otp)
		return fmt.Errorf("invalid otp")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(key.hashedEmail), []byte(email)); err != nil {
		return err
	}

	delete(keys, token)

	return nil
}

func randStr() (string, error) {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func getClientOS(r *http.Request) string {
	userAgent := r.UserAgent()

	switch {
	case strings.Contains(userAgent, "Windows"):
		return "Windows"
	case strings.Contains(userAgent, "Mac OS") || strings.Contains(userAgent, "Macintosh"):
		return "macOS"
	case strings.Contains(userAgent, "Linux"):
		return "Linux"
	case strings.Contains(userAgent, "Android"):
		return "Android"
	case strings.Contains(userAgent, "iPhone") || strings.Contains(userAgent, "iPad") || strings.Contains(userAgent, "iOS"):
		return "iOS"
	default:
		return "Unknown"
	}
}

func otp() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%06d", n.Int64()), nil
}
