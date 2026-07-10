package httpapi

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/ats-tech/cineforge/services/control-api/internal/config"
	"github.com/ats-tech/cineforge/services/control-api/internal/domain"
	"github.com/ats-tech/cineforge/services/control-api/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
	"golang.org/x/text/language"
)

type Server struct {
	cfg      config.Config
	store    *store.Store
	auth     *auth.Client
	temporal client.Client
	log      *slog.Logger
	linkMu   sync.Mutex
	linkSent map[string]time.Time
}
type contextKey string

const userKey contextKey = "user"

type principal struct {
	User     domain.User
	AuthTime time.Time
	MFA      bool
}

func New(cfg config.Config, st *store.Store, authClient *auth.Client, temporalClient client.Client, logger *slog.Logger) *Server {
	return &Server{cfg: cfg, store: st, auth: authClient, temporal: temporalClient, log: logger, linkSent: map[string]time.Time{}}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer, s.securityHeaders)
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Route("/v1", func(r chi.Router) {
		r.Get("/auth/csrf", s.csrf)
		r.With(s.requireCSRF).Post("/auth/email-link", s.requestEmailLink)
		r.With(s.requireCSRF).Post("/auth/session", s.exchangeSession)
		r.Group(func(r chi.Router) {
			r.Use(s.authenticate)
			r.Get("/me", s.getMe)
			r.Get("/me/mfa", s.getMFAStatus)
			r.With(s.requireCSRF).Patch("/me", s.updateMe)
			r.Get("/auth/sessions", s.listSessions)
			r.With(s.requireCSRF).Delete("/auth/sessions/{sessionID}", s.revokeSession)
			r.With(s.requireCSRF).Post("/auth/logout", s.logout)
			r.With(s.requireCSRF).Post("/auth/logout-all", s.logoutAll)
			r.With(s.requireCSRF).Post("/auth/reauth", s.reauthenticate)
			r.With(s.requireCSRF, s.requireRecentAuth, s.requireIdempotency).Post("/me/deletion", s.scheduleAccountDeletion)
			r.With(s.requireCSRF).Post("/me/deletion/cancel", s.cancelAccountDeletion)
			r.Group(func(r chi.Router) {
				r.Use(s.requireVerifiedEmail)
				r.Get("/workspaces", s.listWorkspaces)
				r.With(s.requireCSRF, s.requireIdempotency).Post("/workspaces", s.createWorkspace)
				r.With(s.requireCSRF).Post("/workspaces/bootstrap", s.bootstrapWorkspace)
				r.With(s.requireCSRF).Patch("/workspaces/{workspaceID}", s.updateWorkspace)
				r.Get("/workspaces/{workspaceID}/members", s.listMembers)
				r.With(s.requireCSRF).Patch("/workspaces/{workspaceID}/members/{memberID}", s.updateMember)
				r.Get("/workspaces/{workspaceID}/invitations", s.listInvitations)
				r.With(s.requireCSRF, s.requireIdempotency).Post("/workspaces/{workspaceID}/invitations", s.createInvitation)
				r.With(s.requireCSRF).Delete("/workspaces/{workspaceID}/invitations/{invitationID}", s.revokeInvitation)
				r.With(s.requireCSRF).Post("/invitations/accept", s.acceptInvitation)
				r.With(s.requireCSRF, s.requireRecentAuth, s.requireMFA).Post("/workspaces/{workspaceID}/ownership", s.transferOwnership)
				r.With(s.requireCSRF, s.requireRecentAuth, s.requireMFA, s.requireIdempotency).Delete("/workspaces/{workspaceID}", s.trashWorkspace)
				r.With(s.requireCSRF).Post("/workspaces/{workspaceID}/restore", s.restoreWorkspace)
				r.Get("/workspaces/{workspaceID}/budget-policy", s.getBudget)
				r.With(s.requireCSRF).Put("/workspaces/{workspaceID}/budget-policy", s.updateBudget)
				r.Get("/workspaces/{workspaceID}/projects", s.listProjects)
				r.With(s.requireCSRF, s.requireIdempotency).Post("/workspaces/{workspaceID}/projects", s.createProject)
				r.Get("/projects/{projectID}", s.getProject)
				r.Get("/projects/{projectID}/versions", s.listProjectVersions)
				r.With(s.requireCSRF).Put("/projects/{projectID}", s.updateProject)
				r.With(s.requireCSRF).Post("/projects/{projectID}/archive", s.archiveProject)
				r.With(s.requireCSRF).Post("/projects/{projectID}/restore", s.restoreProject)
				r.With(s.requireCSRF, s.requireIdempotency).Delete("/projects/{projectID}", s.trashProject)
				r.With(s.requireCSRF, s.requireIdempotency).Post("/projects/{projectID}/duplicate", s.duplicateProject)
				r.Get("/project-templates", s.listProjectTemplates)
				r.With(s.requireCSRF, s.requireIdempotency).Post("/workspaces/{workspaceID}/project-templates", s.createProjectTemplate)
				r.With(s.requireCSRF).Put("/project-templates/{templateID}", s.updateProjectTemplate)
				r.Get("/workspaces/{workspaceID}/libraries", s.listLibraries)
				r.With(s.requireCSRF, s.requireIdempotency).Post("/workspaces/{workspaceID}/libraries", s.createLibrary)
				r.Get("/libraries/{libraryID}/entities", s.listEntities)
				r.With(s.requireCSRF, s.requireIdempotency).Post("/libraries/{libraryID}/entities", s.createEntity)
				r.With(s.requireCSRF).Put("/library-entities/{entityID}", s.updateEntity)
				r.With(s.requireCSRF).Post("/library-entities/{entityID}/status", s.setEntityStatus)
				r.With(s.requireCSRF).Put("/projects/{projectID}/library-links", s.linkEntity)
				r.With(s.requireCSRF, s.requireIdempotency).Post("/workspaces/{workspaceID}/media/uploads", s.createMediaUpload)
				r.With(s.requireCSRF).Put("/media/uploads/{assetID}/content", s.uploadMediaContent)
				r.With(s.requireCSRF).Post("/media/uploads/{assetID}/complete", s.completeMediaUpload)
				r.Get("/workspaces/{workspaceID}/audit-events", s.listAuditEvents)
			})
		})
	})
	return r
}

func (s *Server) requestEmailLink(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
		Next  string `json:"next"`
	}
	if !decode(w, r, &req) {
		return
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if _, err := mail.ParseAddress(email); err != nil {
		problem(w, http.StatusBadRequest, "invalid_email", "Enter a valid email address")
		return
	}
	next := safeNext(req.Next)
	key := ipPrefix(r.RemoteAddr) + "|" + email
	s.linkMu.Lock()
	last := s.linkSent[key]
	if time.Since(last) >= time.Minute {
		s.linkSent[key] = time.Now()
	}
	s.linkMu.Unlock()
	if !last.IsZero() && time.Since(last) < time.Minute {
		writeJSON(w, http.StatusAccepted, map[string]any{"sent": true, "retryAfterSeconds": 60})
		return
	}
	continueURL := s.cfg.WebOrigin + "/auth/email-link?next=" + url.QueryEscape(next)
	link, err := s.auth.EmailSignInLink(r.Context(), email, &auth.ActionCodeSettings{URL: continueURL, HandleCodeInApp: true})
	if err == nil {
		err = s.sendEmailSignIn(email, link)
	}
	if err != nil {
		s.log.Warn("email sign-in delivery failed", "error", err)
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"sent": true, "retryAfterSeconds": 60})
}

func safeNext(value string) string {
	if value == "/app" || strings.HasPrefix(value, "/app/") || strings.HasPrefix(value, "/invite?token=") {
		return value
	}
	return "/app"
}

func (s *Server) sendEmailSignIn(email, link string) error {
	from := s.cfg.MailFrom
	if strings.Contains(from, "<") {
		if start, end := strings.LastIndex(from, "<"), strings.LastIndex(from, ">"); start >= 0 && end > start {
			from = from[start+1 : end]
		}
	}
	message := []byte("From: " + s.cfg.MailFrom + "\r\nTo: " + email + "\r\nSubject: Your CineForge sign-in link\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n<div style=\"background:#09090d;color:#f6f3ef;padding:40px;font-family:Arial,sans-serif\"><h1 style=\"letter-spacing:.12em;font-size:18px\">CINEFORGE</h1><h2 style=\"font-size:28px\">Enter your studio</h2><p style=\"color:#aaa6b0;line-height:1.6\">Use this secure, single-use link to sign in. It expires automatically.</p><p><a style=\"display:inline-block;background:#7956ff;color:white;text-decoration:none;padding:14px 20px;border-radius:10px\" href=\"" + html.EscapeString(link) + "\">Continue to CineForge</a></p><p style=\"color:#77727e;font-size:12px\">If you did not request this link, you can ignore this email.</p></div>")
	return smtp.SendMail(s.cfg.MailSMTPAddr, nil, from, []string{email}, message)
}

func (s *Server) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) csrf(w http.ResponseWriter, _ *http.Request) {
	token, _ := randomToken(32)
	http.SetCookie(w, &http.Cookie{Name: "cf_csrf", Value: token, Path: "/", Secure: s.cfg.SessionCookieSecure, HttpOnly: false, SameSite: http.SameSiteLaxMode, MaxAge: 3600})
	writeJSON(w, http.StatusOK, map[string]string{"csrfToken": token})
}
func (s *Server) requireCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("cf_csrf")
		header := r.Header.Get("X-CSRF-Token")
		if err != nil || header == "" || len(cookie.Value) != len(header) || subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(header)) != 1 {
			problem(w, http.StatusForbidden, "csrf_failed", "CSRF validation failed")
			return
		}
		next.ServeHTTP(w, r)
	})
}

type bufferedResponse struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func (w *bufferedResponse) Header() http.Header { return w.header }
func (w *bufferedResponse) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
	}
}
func (w *bufferedResponse) Write(value []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.body.Write(value)
}

func (s *Server) requireIdempotency(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("Idempotency-Key")
		if len(key) < 16 || len(key) > 128 {
			problem(w, http.StatusBadRequest, "idempotency_key_required", "A 16 to 128 character Idempotency-Key is required")
			return
		}
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
		if err != nil {
			problem(w, http.StatusRequestEntityTooLarge, "request_too_large", "The request body is too large")
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))
		hash := fmt.Sprintf("%x", sha256.Sum256(append([]byte(r.Method+" "+r.URL.Path+"\n"), body...)))
		u := currentUser(r)
		workspaceID := chi.URLParam(r, "workspaceID")
		route := r.Method + " " + r.URL.Path
		replay, err := s.store.BeginIdempotency(r.Context(), workspaceID, u.ID, key, route, hash)
		if err != nil {
			if errors.Is(err, store.ErrIdempotencyMismatch) || errors.Is(err, store.ErrIdempotencyInProgress) {
				problem(w, http.StatusConflict, "idempotency_conflict", err.Error())
				return
			}
			internal(s, w, r, err)
			return
		}
		if replay.Exists {
			w.Header().Set("Idempotency-Replayed", "true")
			if len(replay.Body) > 0 {
				w.Header().Set("Content-Type", "application/json")
			}
			w.WriteHeader(replay.Status)
			_, _ = w.Write(replay.Body)
			return
		}
		buffer := &bufferedResponse{header: make(http.Header)}
		next.ServeHTTP(buffer, r)
		if buffer.status == 0 {
			buffer.status = http.StatusOK
		}
		if buffer.status >= 500 {
			s.store.AbandonIdempotency(r.Context(), workspaceID, u.ID, key, route)
		} else if err := s.store.CompleteIdempotency(r.Context(), workspaceID, u.ID, key, route, buffer.status, buffer.body.Bytes()); err != nil {
			s.store.AbandonIdempotency(r.Context(), workspaceID, u.ID, key, route)
			internal(s, w, r, err)
			return
		}
		for name, values := range buffer.header {
			for _, value := range values {
				w.Header().Add(name, value)
			}
		}
		w.WriteHeader(buffer.status)
		_, _ = w.Write(buffer.body.Bytes())
	})
}

func (s *Server) exchangeSession(w http.ResponseWriter, r *http.Request) {
	var req struct{ IDToken, DeviceLabel string }
	if !decode(w, r, &req) {
		return
	}
	token, err := s.auth.VerifyIDToken(r.Context(), req.IDToken)
	if err != nil {
		problem(w, http.StatusUnauthorized, "invalid_identity_token", "The identity token is invalid")
		return
	}
	authTime := claimInt64(token.Claims, "auth_time")
	if authTime == 0 || time.Since(time.Unix(authTime, 0)) > 5*time.Minute {
		problem(w, http.StatusUnauthorized, "recent_auth_required", "Sign in again before creating a session")
		return
	}
	email, _ := token.Claims["email"].(string)
	verified, _ := token.Claims["email_verified"].(bool)
	name, _ := token.Claims["name"].(string)
	var avatar *string
	if v, ok := token.Claims["picture"].(string); ok && v != "" {
		avatar = &v
	}
	user, err := s.store.UpsertUser(r.Context(), store.Identity{UID: token.UID, Email: email, EmailVerified: verified, DisplayName: name, AvatarURL: avatar})
	if err != nil {
		internal(s, w, r, err)
		return
	}
	expires := time.Duration(s.cfg.SessionDurationHours) * time.Hour
	sessionCookie, err := s.auth.SessionCookie(r.Context(), req.IDToken, expires)
	if err != nil {
		internal(s, w, r, err)
		return
	}
	deviceToken, _ := randomToken(32)
	device := strings.TrimSpace(req.DeviceLabel)
	if device == "" {
		device = "Web browser"
	}
	sessionID, err := s.store.CreateSession(r.Context(), user.ID, deviceToken, device, r.UserAgent(), ipPrefix(r.RemoteAddr), time.Now().Add(expires))
	if err != nil {
		internal(s, w, r, err)
		return
	}
	setSessionCookies(w, s.cfg, sessionCookie, deviceToken, int(expires.Seconds()))
	writeJSON(w, http.StatusCreated, map[string]any{"user": user, "sessionId": sessionID, "requiresVerification": !verified})
}

func (s *Server) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := r.Cookie("cf_session")
		if err != nil {
			problem(w, http.StatusUnauthorized, "authentication_required", "Sign in to continue")
			return
		}
		device, err := r.Cookie("cf_device")
		if err != nil {
			problem(w, http.StatusUnauthorized, "authentication_required", "The device session is missing")
			return
		}
		token, err := s.auth.VerifySessionCookieAndCheckRevoked(r.Context(), session.Value)
		if err != nil {
			clearSessionCookies(w, s.cfg)
			problem(w, http.StatusUnauthorized, "session_expired", "The session is invalid or expired")
			return
		}
		user, err := s.store.GetUserByIdentityUID(r.Context(), token.UID)
		if err != nil || s.store.ValidateSession(r.Context(), user.ID, device.Value) != nil {
			clearSessionCookies(w, s.cfg)
			problem(w, http.StatusUnauthorized, "session_revoked", "The session has been revoked")
			return
		}
		if user.DeletionScheduledAt != nil && r.URL.Path != "/v1/me" && r.URL.Path != "/v1/me/deletion/cancel" {
			problem(w, http.StatusLocked, "account_deletion_pending", "Cancel account deletion before using CineForge")
			return
		}
		authTime := time.Unix(claimInt64(token.Claims, "auth_time"), 0)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey, principal{User: user, AuthTime: authTime, MFA: tokenHasMFA(token.Claims)})))
	})
}

func currentUser(r *http.Request) domain.User { return r.Context().Value(userKey).(principal).User }
func (s *Server) requireVerifiedEmail(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !currentUser(r).EmailVerified {
			problem(w, http.StatusForbidden, "email_verification_required", "Verify your email before continuing")
			return
		}
		next.ServeHTTP(w, r)
	})
}
func (s *Server) requireRecentAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, ok := r.Context().Value(userKey).(principal)
		if !ok || p.AuthTime.IsZero() || time.Since(p.AuthTime) > 5*time.Minute {
			problem(w, http.StatusUnauthorized, "recent_auth_required", "Sign in again before performing this action")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) requireMFA(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		p:=r.Context().Value(userKey).(principal);record,err:=s.auth.GetUser(r.Context(),p.User.IdentityProviderUID)
		if err!=nil{internal(s,w,r,err);return}
		if record.MultiFactor==nil||len(record.MultiFactor.EnrolledFactors)==0{problem(w,http.StatusForbidden,"mfa_enrollment_required","Set up an authenticator before this high-risk action");return}
		if !p.MFA{problem(w,http.StatusUnauthorized,"mfa_challenge_required","Complete multi-factor authentication to continue");return}
		next.ServeHTTP(w,r)
	})
}

func tokenHasMFA(claims map[string]any) bool { if fb,ok:=claims["firebase"].(map[string]any);ok { if v,ok:=fb["sign_in_second_factor"].(string);ok&&v!=""{return true} }; return false }

func (s *Server) reauthenticate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDToken string `json:"idToken"`
	}
	if !decode(w, r, &req) {
		return
	}
	token, err := s.auth.VerifyIDToken(r.Context(), req.IDToken)
	if err != nil || token.UID != currentUser(r).IdentityProviderUID || time.Since(time.Unix(claimInt64(token.Claims, "auth_time"), 0)) > 5*time.Minute {
		problem(w, http.StatusUnauthorized, "recent_auth_required", "Complete sign-in again to continue")
		return
	}
	expires := time.Duration(s.cfg.SessionDurationHours) * time.Hour
	cookie, err := s.auth.SessionCookie(r.Context(), req.IDToken, expires)
	if err != nil {
		internal(s, w, r, err)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "cf_session", Value: cookie, Path: "/", HttpOnly: true, Secure: s.cfg.SessionCookieSecure, SameSite: http.SameSiteLaxMode, MaxAge: int(expires.Seconds())})
	writeJSON(w, http.StatusOK, map[string]bool{"reauthenticated": true})
}

func (s *Server) sendInvitation(email, token string) error {
	from := s.cfg.MailFrom
	if strings.Contains(from, "<") {
		if start, end := strings.LastIndex(from, "<"), strings.LastIndex(from, ">"); start >= 0 && end > start {
			from = from[start+1 : end]
		}
	}
	message := []byte("From: " + s.cfg.MailFrom + "\r\nTo: " + email + "\r\nSubject: CineForge workspace invitation\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\nAccept your invitation at " + s.cfg.WebOrigin + "/invite?token=" + url.QueryEscape(token) + "\r\nThis invitation expires in seven days.\r\n")
	return smtp.SendMail(s.cfg.MailSMTPAddr, nil, from, []string{email}, message)
}
func setSessionCookies(w http.ResponseWriter, cfg config.Config, session, device string, maxAge int) {
	for _, c := range []*http.Cookie{{Name: "cf_session", Value: session, Path: "/", HttpOnly: true, Secure: cfg.SessionCookieSecure, SameSite: http.SameSiteLaxMode, MaxAge: maxAge}, {Name: "cf_device", Value: device, Path: "/", HttpOnly: true, Secure: cfg.SessionCookieSecure, SameSite: http.SameSiteLaxMode, MaxAge: maxAge}} {
		http.SetCookie(w, c)
	}
}
func clearSessionCookies(w http.ResponseWriter, cfg config.Config) {
	setSessionCookies(w, cfg, "", "", -1)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		problem(w, http.StatusBadRequest, "invalid_request", err.Error())
		return false
	}
	return true
}
func problem(w http.ResponseWriter, status int, code, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	writeJSON(w, status, map[string]any{"type": "https://cineforge.local/problems/" + code, "title": http.StatusText(status), "status": status, "code": code, "detail": detail})
}
func internal(s *Server, w http.ResponseWriter, r *http.Request, err error) {
	s.log.Error("request failed", "error", err, "requestId", middleware.GetReqID(r.Context()))
	problem(w, http.StatusInternalServerError, "internal_error", "The request could not be completed")
}
func handleError(s *Server, w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		problem(w, http.StatusNotFound, "not_found", "The requested resource was not found")
	case errors.Is(err, store.ErrForbidden):
		problem(w, http.StatusForbidden, "forbidden", "You do not have permission for this action")
	case errors.Is(err, store.ErrConflict):
		problem(w, http.StatusPreconditionFailed, "version_conflict", "The resource changed; reload and try again")
	case errors.Is(err, store.ErrOwnerBlocked):
		problem(w, http.StatusConflict, "ownership_required", err.Error())
	default:
		internal(s, w, r, err)
	}
}
func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
func claimInt64(claims map[string]any, key string) int64 {
	switch v := claims[key].(type) {
	case int64:
		return v
	case float64:
		return int64(v)
	case json.Number:
		i, _ := v.Int64()
		return i
	}
	return 0
}
func ipPrefix(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return ""
	}
	if v := ip.To4(); v != nil {
		return fmt.Sprintf("%d.%d.%d.0/24", v[0], v[1], v[2])
	}
	return ip.Mask(net.CIDRMask(48, 128)).String() + "/48"
}
func parseVersion(r *http.Request) (int64, error) {
	v := strings.Trim(r.Header.Get("If-Match"), `"`)
	return strconv.ParseInt(v, 10, 64)
}
func validLocale(v string) bool { _, err := language.Parse(v); return err == nil }
func newID() string             { return uuid.Must(uuid.NewV7()).String() }

var _ = newID
