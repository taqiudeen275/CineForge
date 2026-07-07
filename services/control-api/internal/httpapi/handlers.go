package httpapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ats-tech/cineforge/services/control-api/internal/domain"
	"github.com/ats-tech/cineforge/services/control-api/internal/store"
	"github.com/go-chi/chi/v5"
	"go.temporal.io/sdk/client"
)

func (s *Server) getMe(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, currentUser(r))
}
func (s *Server) updateMe(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	version, err := parseVersion(r)
	if err != nil {
		problem(w, 428, "precondition_required", "A valid If-Match version is required")
		return
	}
	var req struct {
		DisplayName string  `json:"displayName"`
		AvatarURL   *string `json:"avatarUrl"`
		Locale      string  `json:"locale"`
		Timezone    string  `json:"timezone"`
	}
	if !decode(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.DisplayName) == "" || !validLocale(req.Locale) || req.Timezone == "" {
		problem(w, 400, "invalid_profile", "Display name, locale, and timezone are required")
		return
	}
	out, err := s.store.UpdateUser(r.Context(), u.ID, version, req.DisplayName, req.AvatarURL, req.Locale, req.Timezone)
	if err != nil {
		handleError(s, w, r, err)
		return
	}
	w.Header().Set("ETag", strconv.FormatInt(out.Version, 10))
	writeJSON(w, 200, out)
}
func (s *Server) listSessions(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	out, err := s.store.ListSessions(r.Context(), u.ID)
	if err != nil {
		handleError(s, w, r, err)
		return
	}
	writeJSON(w, 200, map[string]any{"items": out})
}
func (s *Server) revokeSession(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	if err := s.store.RevokeSession(r.Context(), u.ID, chi.URLParam(r, "sessionID")); err != nil {
		handleError(s, w, r, err)
		return
	}
	w.WriteHeader(204)
}
func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	if device, err := r.Cookie("cf_device"); err == nil {
		_ = s.store.RevokeSessionByToken(r.Context(), u.ID, device.Value)
	}
	clearSessionCookies(w, s.cfg)
	w.WriteHeader(204)
}
func (s *Server) logoutAll(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	if err := s.store.RevokeAllSessions(r.Context(), u.ID); err != nil {
		handleError(s, w, r, err)
		return
	}
	_ = s.auth.RevokeRefreshTokens(r.Context(), u.IdentityProviderUID)
	clearSessionCookies(w, s.cfg)
	w.WriteHeader(204)
}
func (s *Server) scheduleAccountDeletion(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	if err := s.store.ScheduleAccountDeletion(r.Context(), u.ID); err != nil {
		handleError(s, w, r, err)
		return
	}
	clearSessionCookies(w, s.cfg)
	w.WriteHeader(202)
}
func (s *Server) cancelAccountDeletion(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	if err := s.store.CancelAccountDeletion(r.Context(), u.ID); err != nil {
		handleError(s, w, r, err)
		return
	}
	w.WriteHeader(204)
}

func (s *Server) bootstrapWorkspace(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	out, err := s.store.BootstrapPersonalWorkspace(r.Context(), u)
	if err != nil {
		handleError(s, w, r, err)
		return
	}
	writeJSON(w, 200, out)
}
func (s *Server) listWorkspaces(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	out, err := s.store.ListWorkspaces(r.Context(), u.ID)
	if err != nil {
		handleError(s, w, r, err)
		return
	}
	writeJSON(w, 200, map[string]any{"items": out})
}
func (s *Server) createWorkspace(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	var req struct {
		Name string `json:"name"`
	}
	if !decode(w, r, &req) {
		return
	}
	if len(strings.TrimSpace(req.Name)) < 2 {
		problem(w, 400, "invalid_name", "Workspace name is too short")
		return
	}
	out, err := s.store.CreateTeamWorkspace(r.Context(), u.ID, req.Name)
	if err != nil {
		handleError(s, w, r, err)
		return
	}
	writeJSON(w, 201, out)
}
func (s *Server) listMembers(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	out, err := s.store.ListMembers(r.Context(), u.ID, chi.URLParam(r, "workspaceID"))
	if err != nil {
		handleError(s, w, r, err)
		return
	}
	writeJSON(w, 200, map[string]any{"items": out})
}
func (s *Server) updateMember(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	var req struct {
		Role               domain.Role `json:"role"`
		CanSpend           bool        `json:"canSpend"`
		MonthlyLimitMicros *int64      `json:"monthlyLimitMicros"`
		PerRunLimitMicros  *int64      `json:"perRunLimitMicros"`
		LibraryPublish     bool        `json:"libraryPublish"`
	}
	if !decode(w, r, &req) {
		return
	}
	err := s.store.UpdateMember(r.Context(), u.ID, chi.URLParam(r, "workspaceID"), chi.URLParam(r, "memberID"), req.Role, req.CanSpend, req.MonthlyLimitMicros, req.PerRunLimitMicros, req.LibraryPublish)
	if err != nil {
		handleError(s, w, r, err)
		return
	}
	w.WriteHeader(204)
}
func (s *Server) createInvitation(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	var req struct {
		Email string      `json:"email"`
		Role  domain.Role `json:"role"`
	}
	if !decode(w, r, &req) {
		return
	}
	token, _ := randomToken(32)
	id, err := s.store.CreateInvitation(r.Context(), u.ID, chi.URLParam(r, "workspaceID"), req.Email, req.Role, token, time.Now().Add(7*24*time.Hour))
	if err != nil {
		handleError(s, w, r, err)
		return
	}
	if err := s.sendInvitation(req.Email, token); err != nil {
		s.log.Warn("invitation email delivery failed", "invitationId", id, "error", err)
	}
	out := map[string]any{"id": id, "expiresInDays": 7}
	if s.cfg.Environment == "development" {
		out["developmentToken"] = token
	}
	writeJSON(w, 201, out)
}
func (s *Server) acceptInvitation(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	var req struct {
		Token string `json:"token"`
	}
	if !decode(w, r, &req) {
		return
	}
	out, err := s.store.AcceptInvitation(r.Context(), u, req.Token)
	if err != nil {
		handleError(s, w, r, err)
		return
	}
	writeJSON(w, 200, out)
}
func (s *Server) transferOwnership(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	var req struct {
		UserID string `json:"userId"`
	}
	if !decode(w, r, &req) {
		return
	}
	if err := s.store.TransferOwnership(r.Context(), u.ID, chi.URLParam(r, "workspaceID"), req.UserID); err != nil {
		handleError(s, w, r, err)
		return
	}
	w.WriteHeader(204)
}
func (s *Server) trashWorkspace(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	if err := s.store.TrashWorkspace(r.Context(), u.ID, chi.URLParam(r, "workspaceID")); err != nil {
		handleError(s, w, r, err)
		return
	}
	w.WriteHeader(202)
}
func (s *Server) restoreWorkspace(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	if err := s.store.RestoreWorkspace(r.Context(), u.ID, chi.URLParam(r, "workspaceID")); err != nil {
		handleError(s, w, r, err)
		return
	}
	w.WriteHeader(204)
}
func (s *Server) getBudget(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	out, err := s.store.GetBudget(r.Context(), u.ID, chi.URLParam(r, "workspaceID"))
	if err != nil {
		handleError(s, w, r, err)
		return
	}
	w.Header().Set("ETag", strconv.FormatInt(out.Version, 10))
	writeJSON(w, 200, out)
}
func (s *Server) updateBudget(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	expected, err := parseVersion(r)
	if err != nil {
		problem(w, 428, "precondition_required", "A valid If-Match version is required")
		return
	}
	var req domain.BudgetPolicy
	if !decode(w, r, &req) {
		return
	}
	if req.Currency != "USD" {
		problem(w, 400, "unsupported_currency", "Only USD is supported in this slice")
		return
	}
	out, err := s.store.UpdateBudget(r.Context(), u.ID, chi.URLParam(r, "workspaceID"), expected, req)
	if err != nil {
		handleError(s, w, r, err)
		return
	}
	w.Header().Set("ETag", strconv.FormatInt(out.Version, 10))
	writeJSON(w, 200, out)
}

func defaultProject(in domain.ProjectSettings, u domain.User) domain.ProjectSettings {
	if in.AspectWidth == 0 {
		in.AspectWidth = 16
		in.AspectHeight = 9
	}
	if in.FrameRateNumerator == 0 {
		in.FrameRateNumerator = 24
		in.FrameRateDenominator = 1
	}
	if in.AudioLanguage == "" {
		in.AudioLanguage = u.Locale
		if !validLocale(in.AudioLanguage) {
			in.AudioLanguage = "en"
		}
	}
	if in.Rating == "" {
		in.Rating = "moderate"
	}
	if in.QualityPolicy == "" {
		in.QualityPolicy = "balanced"
	}
	return in
}
func validateProject(p domain.ProjectSettings) bool {
	validType := p.ProjectType == "single" || p.ProjectType == "series"
	formats := map[string]bool{"short_film": true, "feature": true, "episodic": true, "music_video": true, "advertisement": true, "trailer": true, "other": true}
	ratings := map[string]bool{"family": true, "moderate": true, "mature": true}
	quality := map[string]bool{"draft": true, "balanced": true, "final": true}
	return strings.TrimSpace(p.Name) != "" && validType && formats[p.ProductionFormat] && p.AspectWidth > 0 && p.AspectHeight > 0 && p.FrameRateNumerator > 0 && p.FrameRateDenominator > 0 && validLocale(p.AudioLanguage) && ratings[p.Rating] && quality[p.QualityPolicy] && len(p.StyleDirection) <= 2000
}
func (s *Server) listProjects(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	out, err := s.store.ListProjects(r.Context(), u.ID, chi.URLParam(r, "workspaceID"), r.URL.Query().Get("status"), r.URL.Query().Get("q"))
	if err != nil {
		handleError(s, w, r, err)
		return
	}
	writeJSON(w, 200, map[string]any{"items": out})
}
func (s *Server) createProject(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	var req domain.ProjectSettings
	if !decode(w, r, &req) {
		return
	}
	req = defaultProject(req, u)
	if !validateProject(req) {
		problem(w, 400, "invalid_project", "Project settings are invalid")
		return
	}
	out, err := s.store.CreateProject(r.Context(), u.ID, chi.URLParam(r, "workspaceID"), req)
	if err != nil {
		handleError(s, w, r, err)
		return
	}
	w.Header().Set("ETag", strconv.FormatInt(out.CurrentVersion, 10))
	writeJSON(w, 201, out)
}
func (s *Server) getProject(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	out, err := s.store.GetProject(r.Context(), u.ID, chi.URLParam(r, "projectID"))
	if err != nil {
		handleError(s, w, r, err)
		return
	}
	w.Header().Set("ETag", strconv.FormatInt(out.CurrentVersion, 10))
	writeJSON(w, 200, out)
}

func (s *Server) listProjectVersions(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	out, err := s.store.ListProjectVersions(r.Context(), u.ID, chi.URLParam(r, "projectID"))
	if err != nil {
		handleError(s, w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (s *Server) updateProject(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	v, err := parseVersion(r)
	if err != nil {
		problem(w, 428, "precondition_required", "A valid If-Match version is required")
		return
	}
	var req domain.ProjectSettings
	if !decode(w, r, &req) {
		return
	}
	if !validateProject(req) {
		problem(w, 400, "invalid_project", "Project settings are invalid")
		return
	}
	out, err := s.store.UpdateProject(r.Context(), u.ID, chi.URLParam(r, "projectID"), v, req)
	if err != nil {
		handleError(s, w, r, err)
		return
	}
	w.Header().Set("ETag", strconv.FormatInt(out.CurrentVersion, 10))
	writeJSON(w, 200, out)
}
func (s *Server) archiveProject(w http.ResponseWriter, r *http.Request) {
	s.projectStatus(w, r, "archived")
}
func (s *Server) restoreProject(w http.ResponseWriter, r *http.Request) {
	s.projectStatus(w, r, "active")
}
func (s *Server) trashProject(w http.ResponseWriter, r *http.Request) {
	s.projectStatus(w, r, "trashed")
}
func (s *Server) projectStatus(w http.ResponseWriter, r *http.Request, status string) {
	u := currentUser(r)
	out, err := s.store.SetProjectStatus(r.Context(), u.ID, chi.URLParam(r, "projectID"), status)
	if err != nil {
		handleError(s, w, r, err)
		return
	}
	writeJSON(w, 200, out)
}
func (s *Server) duplicateProject(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	var req struct {
		Name string `json:"name"`
	}
	if !decode(w, r, &req) {
		return
	}
	out, err := s.store.DuplicateProject(r.Context(), u.ID, chi.URLParam(r, "projectID"), req.Name)
	if err != nil {
		handleError(s, w, r, err)
		return
	}
	writeJSON(w, 201, out)
}

func (s *Server) listProjectTemplates(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	out, err := s.store.ListProjectTemplates(r.Context(), u.ID, r.URL.Query().Get("workspaceId"))
	if err != nil {
		handleError(s, w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (s *Server) createProjectTemplate(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	var req struct {
		ProjectID   string `json:"projectId"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if !decode(w, r, &req) {
		return
	}
	out, err := s.store.SaveProjectTemplate(r.Context(), u.ID, chi.URLParam(r, "workspaceID"), req.ProjectID, "", req.Name, req.Description, 0)
	if err != nil {
		handleError(s, w, r, err)
		return
	}
	w.Header().Set("ETag", strconv.FormatInt(out.CurrentVersion, 10))
	writeJSON(w, http.StatusCreated, out)
}

func (s *Server) updateProjectTemplate(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	expected, err := parseVersion(r)
	if err != nil {
		problem(w, 428, "precondition_required", "A valid If-Match version is required")
		return
	}
	var req struct {
		WorkspaceID string `json:"workspaceId"`
		ProjectID   string `json:"projectId"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if !decode(w, r, &req) {
		return
	}
	out, err := s.store.SaveProjectTemplate(r.Context(), u.ID, req.WorkspaceID, req.ProjectID, chi.URLParam(r, "templateID"), req.Name, req.Description, expected)
	if err != nil {
		handleError(s, w, r, err)
		return
	}
	w.Header().Set("ETag", strconv.FormatInt(out.CurrentVersion, 10))
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) listAuditEvents(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	out, err := s.store.ListAuditEvents(r.Context(), u.ID, chi.URLParam(r, "workspaceID"))
	if err != nil {
		handleError(s, w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (s *Server) listLibraries(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	out, err := s.store.ListLibraries(r.Context(), u.ID, chi.URLParam(r, "workspaceID"))
	if err != nil {
		handleError(s, w, r, err)
		return
	}
	writeJSON(w, 200, map[string]any{"items": out})
}
func (s *Server) createLibrary(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	var req struct{ Name, Type string }
	if !decode(w, r, &req) {
		return
	}
	out, err := s.store.CreateLibrary(r.Context(), u.ID, chi.URLParam(r, "workspaceID"), req.Name, req.Type)
	if err != nil {
		handleError(s, w, r, err)
		return
	}
	writeJSON(w, 201, out)
}
func entityInput(w http.ResponseWriter, r *http.Request) (store.EntityInput, bool) {
	var req store.EntityInput
	if !decode(w, r, &req) {
		return req, false
	}
	if req.Name == "" || req.Type == "" {
		problem(w, 400, "invalid_entity", "Name and type are required")
		return req, false
	}
	return req, true
}
func (s *Server) listEntities(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	out, err := s.store.ListLibraryEntities(r.Context(), u.ID, chi.URLParam(r, "libraryID"), r.URL.Query().Get("q"), r.URL.Query().Get("type"), r.URL.Query().Get("status"))
	if err != nil {
		handleError(s, w, r, err)
		return
	}
	writeJSON(w, 200, map[string]any{"items": out})
}
func (s *Server) createEntity(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	req, ok := entityInput(w, r)
	if !ok {
		return
	}
	out, err := s.store.CreateLibraryEntity(r.Context(), u.ID, chi.URLParam(r, "libraryID"), req)
	if err != nil {
		handleError(s, w, r, err)
		return
	}
	writeJSON(w, 201, out)
}
func (s *Server) updateEntity(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	req, ok := entityInput(w, r)
	if !ok {
		return
	}
	out, err := s.store.UpdateLibraryEntity(r.Context(), u.ID, chi.URLParam(r, "entityID"), req)
	if err != nil {
		handleError(s, w, r, err)
		return
	}
	writeJSON(w, 200, out)
}
func (s *Server) setEntityStatus(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	var req struct {
		Status string `json:"status"`
	}
	if !decode(w, r, &req) {
		return
	}
	out, err := s.store.SetLibraryEntityStatus(r.Context(), u.ID, chi.URLParam(r, "entityID"), req.Status)
	if err != nil {
		handleError(s, w, r, err)
		return
	}
	writeJSON(w, 200, out)
}
func (s *Server) linkEntity(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	var req struct {
		EntityVersionID string `json:"entityVersionId"`
	}
	if !decode(w, r, &req) {
		return
	}
	if err := s.store.LinkEntityToProject(r.Context(), u.ID, chi.URLParam(r, "projectID"), req.EntityVersionID); err != nil {
		handleError(s, w, r, err)
		return
	}
	w.WriteHeader(204)
}
func (s *Server) createMediaUpload(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	var req store.UploadInput
	if !decode(w, r, &req) {
		return
	}
	out, err := s.store.CreateMediaUpload(r.Context(), u.ID, chi.URLParam(r, "workspaceID"), req)
	if err != nil {
		handleError(s, w, r, err)
		return
	}
	writeJSON(w, 201, map[string]any{"asset": out, "uploadUrl": "/v1/media/uploads/" + out.ID + "/content", "expiresInSeconds": 600})
}

func (s *Server) uploadMediaContent(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	target, err := s.store.MediaUploadTarget(r.Context(), u.ID, chi.URLParam(r, "assetID"))
	if err != nil {
		handleError(s, w, r, err)
		return
	}
	if r.ContentLength < 0 || r.ContentLength != target.SizeBytes || r.Header.Get("Content-Type") != target.ContentType {
		problem(w, http.StatusBadRequest, "upload_mismatch", "Uploaded content does not match the authorized size and type")
		return
	}
	endpoint := strings.TrimRight(s.cfg.GCSUploadEndpoint, "/") + "/b/" + url.PathEscape(s.cfg.GCSQuarantineBucket) + "/o?uploadType=media&name=" + url.QueryEscape(target.Object)
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, endpoint, io.LimitReader(r.Body, target.SizeBytes+1))
	if err != nil {
		internal(s, w, r, err)
		return
	}
	req.Header.Set("Content-Type", target.ContentType)
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		internal(s, w, r, fmt.Errorf("write quarantine object: %w", err))
		return
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		internal(s, w, r, fmt.Errorf("quarantine storage returned %s", response.Status))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (s *Server) completeMediaUpload(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	assetID := chi.URLParam(r, "assetID")
	if err := s.store.CompleteMediaUpload(r.Context(), u.ID, assetID); err != nil {
		handleError(s, w, r, err)
		return
	}
	_, err := s.temporal.ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{ID: "media-ingest-" + assetID, TaskQueue: "cineforge-media"}, "IngestMedia", assetID)
	if err != nil {
		handleError(s, w, r, err)
		return
	}
	w.WriteHeader(202)
}
