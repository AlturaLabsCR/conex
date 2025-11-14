package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"net/http"
	"net/mail"
	"strconv"
	"time"

	"github.com/mileusna/useragent"

	"app/config"
	"app/internal/db"
	"app/templates"

	"golang.org/x/crypto/bcrypt"
)

type ctxKey string

const ctxSessionKey ctxKey = "session"

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
	content := templates.Register(h.Translator(r))

	templates.Base(h.Translator(r), header, content, true).Render(ctx, w)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tr := h.Translator(r)

	email := r.FormValue("email")
	if _, err := mail.ParseAddress(email); err != nil {
		templates.Notice(
			templates.RegisterNoticeID,
			templates.NoticeWarn,
			tr("warn"),
			tr("invalid_email"),
		).Render(ctx, w)
		h.Log().Debug("failed to parse email address", "email", email)
		return
	}

	queries := db.New(h.DB())

	if _, err := queries.GetUserByEmail(ctx, email); err == nil {
		templates.Notice(
			templates.RegisterNoticeID,
			templates.NoticeWarn,
			tr("warn"),
			tr("existent_email"),
		).Render(ctx, w)
		h.Log().Debug("email exists", "email", email)
		return
	}

	token, err := h.issueOTP(h.Translator(r), email)
	if err != nil {
		h.Log().Error("error issuing otp", "error", err)
		templates.Notice(
			templates.RegisterNoticeID,
			templates.NoticeWarn,
			tr("warn"),
			tr("invalid_otp"),
		).Render(ctx, w)
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

	tr := h.Translator(r)

	email := r.FormValue("email")
	if _, err := mail.ParseAddress(email); err != nil {
		h.Log().Debug("failed to parse email address", "email", email)
		templates.Notice(
			templates.RegisterNoticeID,
			templates.NoticeWarn,
			tr("warn"),
			tr("invalid_email"),
		).Render(ctx, w)
		return
	}

	otp := r.FormValue("otp")

	if err := h.verifyOTP(key, email, token, otp); err != nil {
		templates.Notice(
			templates.RegisterNoticeID,
			templates.NoticeWarn,
			tr("warn"),
			tr("invalid_otp"),
		).Render(ctx, w)
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
		templates.Notice(
			templates.RegisterNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("register_error"),
		).Render(ctx, w)
		h.Log().Error("error registering user", "error", err)
		return
	}

	h.Log().Info("new user registration")

	if err := h.loginClient(w, r, email); err != nil {
		h.Log().Debug("failed to login user", "error", err)
	}

	templates.Redirect(config.Endpoints[config.DashboardPath]).Render(ctx, w)
}

func (h *Handler) LoginForm(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	_, err := h.verifyClient(w, r, false)
	if err == nil {
		http.Redirect(w, r, config.Endpoints[config.DashboardPath], http.StatusSeeOther)
		return
	}

	header := templates.LoginHeader(h.Translator(r))
	content := templates.Login(h.Translator(r))

	templates.Base(h.Translator(r), header, content, true).Render(ctx, w)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tr := h.Translator(r)

	email := r.FormValue("email")
	if _, err := mail.ParseAddress(email); err != nil {
		templates.Notice(
			templates.LoginNoticeID,
			templates.NoticeWarn,
			tr("warn"),
			tr("invalid_email"),
		).Render(ctx, w)
		h.Log().Debug("failed to parse email address", "email", email)
		return
	}

	queries := db.New(h.DB())

	if _, err := queries.GetUserByEmail(ctx, email); err != nil {
		templates.Notice(
			templates.LoginNoticeID,
			templates.NoticeWarn,
			tr("warn"),
			tr("nonexistent_email"),
		).Render(ctx, w)
		h.Log().Debug("email does not exist", "email", email)
		return
	}

	token, err := h.issueOTP(h.Translator(r), email)
	if err != nil {
		h.Log().Error("error issuing otp", "error", err)
		templates.Notice(
			templates.LoginNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("login_error"),
		).Render(ctx, w)
		return
	}

	templates.LoginConfirmEmail(h.Translator(r), token, email).Render(ctx, w)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	session, ok := ctx.Value(ctxSessionKey).(db.Session)
	if !ok {
		h.Log().Error("error retrieving session from ctx")
		return
	}

	h.Sessions.JWTTerminate(w, r)

	queries := db.New(h.DB())

	queries.DeleteSession(ctx, session.SessionID)

	http.SetCookie(w, &http.Cookie{
		Name:     "csrf",
		Value:    "",
		Path:     config.Endpoints[config.RootPath],
		Expires:  time.Unix(0, 0),
		HttpOnly: false,
		Secure:   r.TLS != nil,
	})

	http.Redirect(w, r, config.Endpoints[config.LoginPath], http.StatusSeeOther)
}

func (h *Handler) LogoutAskedSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	session, ok := ctx.Value(ctxSessionKey).(db.Session)
	if !ok {
		h.Log().Error("error retrieving session from ctx")
		return
	}

	queries := db.New(h.DB())

	sessions, err := queries.GetSessionsByUser(ctx, session.SessionUser)
	if err != nil {
		h.Log().Error("error retrieving sessions by user")
		return
	}

	askedSessionStr := r.PathValue("sessionID")
	askedSession, err := strconv.ParseInt(askedSessionStr, 10, 64)
	if err != nil {
		h.Log().Error("error retrieving asked session from url", "error", err, "session", askedSessionStr)
		return
	}

	var index int
	for id, s := range sessions {
		if s.SessionID == askedSession {
			queries.DeleteSession(ctx, askedSession)
			index = id
			h.Log().Debug("session deleted", "sessionID", askedSession)
			break
		}
	}

	sessions = append(sessions[:index], sessions[index+1:]...)

	if len(sessions) == 0 || askedSession == session.SessionID {
		templates.Redirect(config.Endpoints[config.LoginPath]).Render(ctx, w)
		return
	}

	templates.SessionsTable(h.Translator(r), session, sessions).Render(ctx, w)
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

	tr := h.Translator(r)

	email := r.FormValue("email")
	if _, err := mail.ParseAddress(email); err != nil {
		h.Log().Debug("failed to parse email address", "email", email)
		templates.Notice(
			templates.LoginNoticeID,
			templates.NoticeWarn,
			tr("warn"),
			tr("invalid_email"),
		).Render(ctx, w)
		return
	}

	otp := r.FormValue("otp")

	if err := h.verifyOTP(key, email, token, otp); err != nil {
		templates.Notice(
			templates.LoginNoticeID,
			templates.NoticeWarn,
			tr("warn"),
			tr("invalid_otp"),
		).Render(ctx, w)
		h.Log().Debug("failed to verify otp", "error", err)
		return
	}

	if err := h.loginClient(w, r, email); err != nil {
		h.Log().Debug("failed to login user", "error", err)
	}

	templates.Redirect(config.Endpoints[config.DashboardPath]).Render(ctx, w)
}

func (h *Handler) loginClient(w http.ResponseWriter, r *http.Request, email string) error {
	queries := db.New(h.DB())

	user, err := queries.GetUserByEmail(r.Context(), email)
	if err != nil {
		h.Log().Debug("failed to get user by email", "error", err)
		return err
	}

	ua := useragent.Parse(r.UserAgent())
	device := ua.OS + ", " + ua.Name

	now := time.Now().Unix()

	sessionID, err := queries.InsertSession(r.Context(), db.InsertSessionParams{
		SessionUser:          user.UserID,
		SessionDevice:        device,
		SessionLastLoginUnix: now,
	})
	if err != nil {
		h.Log().Debug("error inserting session", "error", err)
		return err
	}

	_, err = h.Sessions.JWTSet(w, r, db.Session{
		SessionID:            sessionID,
		SessionUser:          user.UserID,
		SessionDevice:        device,
		SessionLastLoginUnix: now,
	})
	if err != nil {
		h.Log().Debug("error setting jwt", "error", err)
		return err
	}

	return nil
}

func (h *Handler) verifyClient(w http.ResponseWriter, r *http.Request, enforceCSRF bool) (db.Session, error) {
	ctx := r.Context()

	session, expired, err := h.Sessions.JWTValidate(r)
	if err != nil {
		templates.Redirect(config.Endpoints[config.LoginPath])
		return db.Session{}, err
	}

	queries := db.New(h.DB())

	exists, err := queries.SessionExists(ctx, session.SessionID)
	if err != nil {
		templates.Redirect(config.Endpoints[config.LoginPath])
		return db.Session{}, err
	}

	if exists != 1 {
		templates.Redirect(config.Endpoints[config.LoginPath])
		return db.Session{}, fmt.Errorf("session does not exist")
	}

	exists, err = queries.UserExists(ctx, session.SessionUser)
	if err != nil {
		templates.Redirect(config.Endpoints[config.LoginPath])
		return db.Session{}, err
	}

	if exists != 1 {
		templates.Redirect(config.Endpoints[config.LoginPath])
		return db.Session{}, fmt.Errorf("user does not exist")
	}

	if enforceCSRF {
		csrfClaim := r.Header.Get(config.CSRFHeaderName)
		oldCSRF, ok := csrfProtection[session.SessionID]
		if !ok || csrfClaim != oldCSRF {
			return db.Session{}, fmt.Errorf("invalid CSRF token")
		}
	}

	newCSRFToken, err := randStr()
	if err != nil {
		return db.Session{}, err
	}

	now := time.Now().Unix()

	if expired {
		_, err = h.Sessions.JWTSet(w, r, db.Session{
			SessionID:            session.SessionID,
			SessionUser:          session.SessionUser,
			SessionDevice:        session.SessionDevice,
			SessionLastLoginUnix: now,
		})
		if err != nil {
			return db.Session{}, err
		}
	}

	csrfProtection[session.SessionID] = newCSRFToken

	http.SetCookie(w, &http.Cookie{
		Name:     "csrf",
		Value:    newCSRFToken,
		Path:     config.Endpoints[config.RootPath],
		Expires:  time.Now().Add(time.Hour),
		HttpOnly: false,
		Secure:   r.TLS != nil,
	})

	queries.UpdateSession(ctx, db.UpdateSessionParams{
		SessionID:            session.SessionID,
		SessionLastLoginUnix: now,
	})

	return session, nil
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

func otp() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%06d", n.Int64()), nil
}

func (h *Handler) AuthenticationMiddleware(enforceCSRF bool, requiredPlan int64, redirect string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			session, err := h.verifyClient(w, r, enforceCSRF)
			if err != nil {
				h.Log().Debug("error validating client", "error", err)

				if redirect != "" {
					http.Redirect(w, r, redirect, http.StatusSeeOther)
				}

				return
			}

			ctx = context.WithValue(ctx, ctxSessionKey, session)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
