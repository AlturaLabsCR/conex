package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/mail"
	"strconv"
	"sync"
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

var (
	keys           sync.Map // map[string]key
	csrfProtection sync.Map // map[int64]string
)

func (h *Handler) RegisterForm(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	header := templates.RegisterHeader(h.Translator(r))
	content := templates.Register(h.Translator(r))

	tr := h.Translator(r)

	head := templates.SiteHead{
		Title:       config.AppTitle + " | " + tr("register"),
		Description: "",
	}

	templates.Base(h.Translator(r), header, content, &head, true).Render(ctx, w)
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

	if len(email) > 63 {
		templates.Notice(
			templates.RegisterNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("invalid_email"),
		).Render(ctx, w)
		return
	}

	if _, err := h.Queries().GetUserByEmail(ctx, email); err == nil {
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
			tr("error"),
			tr("try_later"),
		).Render(ctx, w)
		return
	}

	if err := templates.RegisterConfirmEmail(h.Translator(r), token, email).Render(ctx, w); err != nil {
		h.Log().Error("error rendering template", "error", err)
		return
	}
}

func (h *Handler) RegisterConfirm(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	token := r.FormValue("token")

	val, ok := keys.Load(token)
	if !ok {
		h.Log().Debug("invalid token", "token", token)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	key := val.(key)

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

	tx, err := h.DB().Begin(ctx)
	if err != nil {
		h.Log().Error("error starting tx", "error", err)
		templates.Notice(
			templates.UpdateSettingsNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("try_later"),
		).Render(ctx, w)
		return
	}
	defer tx.Rollback(ctx)

	qtx := h.Queries().WithTx(tx)

	now := time.Now().Unix()

	user, err := qtx.InsertUser(ctx, db.InsertUserParams{
		UserEmail:        email,
		UserCreatedUnix:  now,
		UserModifiedUnix: now,
		UserDeleted:      0,
	})
	if err != nil {
		templates.Notice(
			templates.RegisterNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("register_error"),
		).Render(ctx, w)
		h.Log().Error("error registering user", "error", err)
		return
	}

	if _, err = qtx.InsertPlan(ctx, db.InsertPlanParams{
		UserPlanUser:         user,
		UserPlanCreatedUnix:  now,
		UserPlanModifiedUnix: now,
		UserPlanDueUnix:      0,
		UserPlanActive:       0,
	}); err != nil {
		h.Log().Error("error inserting plan", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.Log().Info("new user registration")

	if err := h.loginClient(w, r, email, qtx); err != nil {
		h.Log().Debug("failed to login user", "error", err)
	}

	if err := tx.Commit(ctx); err != nil {
		h.Log().Debug("failed to register user", "error", err)
		templates.Notice(
			templates.RegisterNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("try_later"),
		).Render(ctx, w)
		return
	}

	templates.Redirect(config.Endpoints[config.DashboardPath]).Render(ctx, w)
}

func (h *Handler) ChangeEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tr := h.Translator(r)

	oldEmail := r.PathValue("email")

	if oldEmail == "" {
		h.Log().Error("error invalid email", "old_email", oldEmail)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	h.Log().Debug("site slug is not empty valid")

	session, ok := ctx.Value(ctxSessionKey).(db.Session)
	if !ok {
		h.Log().Debug("error retrieving session from ctx")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	h.Log().Debug("session id valid")

	user, err := h.Queries().GetUserByID(ctx, session.SessionUser)
	if err != nil {
		h.Log().Error("error querying user by id", "error", err)
		templates.Notice(
			templates.ChangeEmailNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("try_later"),
		).Render(ctx, w)
		return
	}

	if user.UserEmail != oldEmail {
		h.Log().Error("tried to change account email without having account ownership", "error", err)
		templates.Notice(
			templates.ChangeEmailNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("try_later"),
		).Render(ctx, w)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.Log().Error("error reading update email body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req struct {
		Email string `json:"email"`
	}

	if err := json.Unmarshal(body, &req); err != nil {
		h.Log().Error("error invalid update email body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if _, err := mail.ParseAddress(req.Email); err != nil {
		templates.Notice(
			templates.ChangeEmailNoticeID,
			templates.NoticeWarn,
			tr("warn"),
			tr("invalid_email"),
		).Render(ctx, w)
		h.Log().Debug("failed to parse email address")
		return
	}

	if len(req.Email) > 63 {
		templates.Notice(
			templates.ChangeEmailNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("invalid_email"),
		).Render(ctx, w)
		return
	}

	if _, err := h.Queries().GetUserByEmail(ctx, req.Email); err == nil {
		templates.Notice(
			templates.ChangeEmailNoticeID,
			templates.NoticeWarn,
			tr("warn"),
			tr("account_change_email_already_exists"),
		).Render(ctx, w)
		h.Log().Debug("failed to parse email address")
		return
	}

	token, err := h.issueOTP(tr, req.Email)
	if err != nil {
		h.Log().Error("error issuing otp", "error", err)
		templates.Notice(
			templates.RegisterNoticeID,
			templates.NoticeWarn,
			tr("error"),
			tr("try_later"),
		).Render(ctx, w)
		return
	}

	if err := templates.ChangeEmailConfirm(tr, token, oldEmail).Render(ctx, w); err != nil {
		h.Log().Error("error rendering template", "error", err)
		return
	}
}

func (h *Handler) DeleteSite(w http.ResponseWriter, r *http.Request) {
	h.Log().Debug("endpoint hit", "pattern", r.Pattern)

	ctx := r.Context()
	tr := h.Translator(r)

	slug := r.PathValue("site")

	if slug == "" {
		h.Log().Error("error invalid site", "slug", slug)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	h.Log().Debug("site slug is not empty valid")

	session, ok := ctx.Value(ctxSessionKey).(db.Session)
	if !ok {
		h.Log().Debug("error retrieving session from ctx")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	h.Log().Debug("session id valid")

	site, err := h.Queries().GetSiteBySlug(ctx, slug)
	if err != nil {
		h.Log().Error("error querying site by slug", "error", err)
		templates.Notice(
			templates.EditorDeleteSiteNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("try_later"),
		).Render(ctx, w)
		return
	}

	if session.SessionUser != site.SiteUser {
		h.Log().Error("error tried to delete not owned site")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if err := h.Queries().DeleteSite(ctx, site.SiteID); err != nil {
		h.Log().Error("error deleting site", "error", err)
		templates.Notice(
			templates.AccountDeleteNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("try_later"),
		).Render(ctx, w)
		return
	}

	if err := templates.Redirect(config.Endpoints[config.DashboardPath]).Render(ctx, w); err != nil {
		h.Log().Error("error rendering template", "error", err)
		return
	}
}

func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tr := h.Translator(r)

	email := r.PathValue("email")

	if email == "" {
		h.Log().Error("error invalid email", "email", email)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	h.Log().Debug("site slug is not empty valid")

	session, ok := ctx.Value(ctxSessionKey).(db.Session)
	if !ok {
		h.Log().Debug("error retrieving session from ctx")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	h.Log().Debug("session id valid")

	user, err := h.Queries().GetUserByID(ctx, session.SessionUser)
	if err != nil {
		h.Log().Error("error querying user by id", "error", err)
		templates.Notice(
			templates.AccountDeleteNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("try_later"),
		).Render(ctx, w)
		return
	}

	if user.UserEmail != email {
		h.Log().Error("tried to delete account without having ownership", "error", err)
		templates.Notice(
			templates.AccountDeleteNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("try_later"),
		).Render(ctx, w)
		return
	}

	sites, err := h.Queries().GetSitesWithMetricsByUserID(ctx, session.SessionUser)
	if err != nil {
		h.Log().Error("error querying sites by account", "error", err)
		templates.Notice(
			templates.AccountDeleteNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("try_later"),
		).Render(ctx, w)
		return
	}

	tx, err := h.DB().Begin(ctx)
	if err != nil {
		h.Log().Error("error starting tx", "error", err)
		templates.Notice(
			templates.NewSiteNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("try_later"),
		).Render(ctx, w)
		return
	}
	defer tx.Rollback(ctx)

	qtx := h.Queries().WithTx(tx)

	for _, site := range sites {
		if err := qtx.DeleteSite(ctx, site.SiteID); err != nil {
			h.Log().Error("error deleting site", "error", err)
			templates.Notice(
				templates.AccountDeleteNoticeID,
				templates.NoticeError,
				tr("error"),
				tr("try_later"),
			).Render(ctx, w)
			return
		}
	}

	now := time.Now().Unix()

	if err := qtx.DeleteUser(ctx, db.DeleteUserParams{
		UserModifiedUnix: now,
		UserID:           session.SessionUser,
	}); err != nil {
		h.Log().Error("error deleting user", "error", err)
		templates.Notice(
			templates.AccountDeleteNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("try_later"),
		).Render(ctx, w)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		h.Log().Error("error deleting account", "error", err)
		templates.Notice(
			templates.AccountDeleteNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("try_later"),
		).Render(ctx, w)
		return
	}

	h.Logout(w, r)

	if err := templates.Redirect(config.Endpoints[config.RootPath]).Render(ctx, w); err != nil {
		h.Log().Error("error rendering template", "error", err)
		return
	}
}

func (h *Handler) ChangeEmailConfirm(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tr := h.Translator(r)

	oldEmail := r.PathValue("email")

	if oldEmail == "" {
		h.Log().Error("error invalid email", "old_email", oldEmail)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	h.Log().Debug("site slug is not empty valid")

	session, ok := ctx.Value(ctxSessionKey).(db.Session)
	if !ok {
		h.Log().Debug("error retrieving session from ctx")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	h.Log().Debug("session id valid")

	tx, err := h.DB().Begin(ctx)
	if err != nil {
		h.Log().Error("error starting tx", "error", err)
		templates.Notice(
			templates.NewSiteNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("try_later"),
		).Render(ctx, w)
		return
	}
	defer tx.Rollback(ctx)

	user, err := h.Queries().GetUserByID(ctx, session.SessionUser)
	if err != nil {
		h.Log().Error("error querying user by id", "error", err)
		templates.Notice(
			templates.ChangeEmailNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("try_later"),
		).Render(ctx, w)
		return
	}

	if user.UserEmail != oldEmail {
		h.Log().Error("tried to change account email without having account ownership", "error", err)
		templates.Notice(
			templates.ChangeEmailNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("try_later"),
		).Render(ctx, w)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.Log().Error("error reading update email confirm body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req struct {
		Email string `json:"email"`
		OTP   string `json:"otp"`
		Token string `json:"token"`
	}

	if err := json.Unmarshal(body, &req); err != nil {
		h.Log().Error("error invalid update email confirm body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	val, ok := keys.Load(req.Token)
	if !ok {
		h.Log().Debug("invalid token", "token", req.Token)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	key := val.(key)

	if err := h.verifyOTP(key, req.Email, req.Token, req.OTP); err != nil {
		templates.Notice(
			templates.ChangeEmailNoticeID,
			templates.NoticeWarn,
			tr("warn"),
			tr("invalid_otp"),
		).Render(ctx, w)
		h.Log().Debug("failed to verify otp", "error", err)
		return
	}

	if err := h.Queries().UpdateUser(ctx, db.UpdateUserParams{
		UserEmail:        req.Email,
		UserModifiedUnix: time.Now().Unix(),
		UserID:           session.SessionUser,
	}); err != nil {
		h.Log().Error("error updating user", "error", err)
		templates.Notice(
			templates.ChangeEmailNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("try_later"),
		).Render(ctx, w)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		h.Log().Error("error updating user", "error", err)
		templates.Notice(
			templates.ChangeEmailNoticeID,
			templates.NoticeError,
			tr("error"),
			tr("try_later"),
		).Render(ctx, w)
		return
	}

	if err := templates.AccountHeaderEmail(req.Email).Render(ctx, w); err != nil {
		h.Log().Error("error rendering template", "error", err)
		return
	}

	if err := templates.ChangeEmailForm(tr, req.Email).Render(ctx, w); err != nil {
		h.Log().Error("error rendering template", "error", err)
		return
	}

	if err := templates.Notice(
		templates.ChangeEmailNoticeID,
		templates.NoticeInfo,
		tr("info"),
		tr("account_change_email_success"),
	).Render(ctx, w); err != nil {
		h.Log().Error("error rendering template", "error", err)
		return
	}
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

	tr := h.Translator(r)

	head := templates.SiteHead{
		Title:       config.AppTitle + " | " + tr("log_in"),
		Description: "",
	}

	templates.Base(h.Translator(r), header, content, &head, true).Render(ctx, w)
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
		h.Log().Debug("failed to parse email address")
		return
	}

	if _, err := h.Queries().GetUserByEmail(ctx, email); err != nil {
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

	h.Queries().DeleteSession(ctx, session.SessionID)

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

	sessions, err := h.Queries().GetSessionsByUser(ctx, session.SessionUser)
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
			h.Queries().DeleteSession(ctx, askedSession)
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

	val, ok := keys.Load(token)
	if !ok {
		h.Log().Debug("invalid token", "token", token)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	key := val.(key)

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

	if err := h.loginClient(w, r, email, h.Queries()); err != nil {
		h.Log().Debug("failed to login user", "error", err)
	}

	templates.Redirect(config.Endpoints[config.DashboardPath]).Render(ctx, w)
}

func (h *Handler) loginClient(w http.ResponseWriter, r *http.Request, email string, queries *db.Queries) error {
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

	exists, err := h.Queries().SessionExists(ctx, session.SessionID)
	if err != nil {
		templates.Redirect(config.Endpoints[config.LoginPath])
		return db.Session{}, err
	}

	if !exists {
		templates.Redirect(config.Endpoints[config.LoginPath])
		return db.Session{}, fmt.Errorf("session does not exist")
	}

	exists, err = h.Queries().UserExists(ctx, session.SessionUser)
	if err != nil {
		templates.Redirect(config.Endpoints[config.LoginPath])
		return db.Session{}, err
	}

	if !exists {
		templates.Redirect(config.Endpoints[config.LoginPath])
		return db.Session{}, fmt.Errorf("user does not exist")
	}

	if enforceCSRF {
		csrfClaim := r.Header.Get(config.CSRFHeaderName)

		val, ok := csrfProtection.Load(session.SessionID)
		if !ok {
			return db.Session{}, fmt.Errorf("invalid CSRF token")
		}
		oldCSRF := val.(string)

		if csrfClaim != oldCSRF {
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

	csrfProtection.Store(session.SessionID, newCSRFToken)

	http.SetCookie(w, &http.Cookie{
		Name:     "csrf",
		Value:    newCSRFToken,
		Path:     config.Endpoints[config.RootPath],
		Expires:  time.Now().Add(time.Hour),
		HttpOnly: false,
		Secure:   r.TLS != nil,
	})

	h.Queries().UpdateSession(ctx, db.UpdateSessionParams{
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

	keys.Store(token, key)

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
		keys.Delete(token)
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

	keys.Delete(token)

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
