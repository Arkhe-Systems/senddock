package handler

import (
	"encoding/json"
	"net/http"

	"github.com/arkhe-systems/senddock/pkg/auth"
	"github.com/arkhe-systems/senddock/internal/response"
	"github.com/arkhe-systems/senddock/internal/service"
)

type ProjectHandler struct {
	projectService *service.ProjectService
	Audit          *service.AuditService
}

func NewProjectHandler(projectService *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{projectService: projectService}
}

type createProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	WorkspaceID string `json:"workspace_id"`
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)

	var req createProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "invalid request body"})
		return
	}

	if req.Name == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "name is required"})
		return
	}

	if req.WorkspaceID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "workspace_id is required"})
		return
	}

	project, err := h.projectService.Create(r.Context(), userID, req.WorkspaceID, req.Name, req.Description)
	if err != nil {
		status := http.StatusInternalServerError
		if err == service.ErrWorkspaceForbidden {
			status = http.StatusForbidden
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	if h.Audit != nil {
		h.Audit.LogFromRequest(r, project.ID.String(), userID, "project.create", "project", project.ID.String(), map[string]any{"name": project.Name})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response.FromProject(project))
}

type updateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)
	projectID := r.PathValue("id")

	var req updateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "invalid request body"})
		return
	}

	if req.Name == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "name is required"})
		return
	}

	project, err := h.projectService.Update(r.Context(), projectID, userID, req.Name, req.Description)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	if h.Audit != nil {
		h.Audit.LogFromRequest(r, projectID, userID, "project.update", "project", projectID, map[string]any{"name": project.Name})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response.FromProject(project))
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)

	projects, err := h.projectService.ListByUser(r.Context(), userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response.FromProjects(projects))
}

func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)
	projectID := r.PathValue("id")

	project, err := h.projectService.GetByID(r.Context(), projectID, userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response.FromProject(project))
}

type updateSMTPRequest struct {
	SmtpHost     string `json:"smtp_host"`
	SmtpPort     int32  `json:"smtp_port"`
	SmtpUser     string `json:"smtp_user"`
	SmtpPassword string `json:"smtp_password"`
	FromName     string `json:"from_name"`
	FromEmail    string `json:"from_email"`
}

type updateUnsubscribeTemplateRequest struct {
	TemplateID string `json:"template_id"`
}

func (h *ProjectHandler) UpdateUnsubscribeTemplate(w http.ResponseWriter, r *http.Request) {
	projectID, userID, ok := requireCap(w, r, h.projectService, service.CapProjectSettings)
	if !ok {
		return
	}

	var req updateUnsubscribeTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "invalid request body"})
		return
	}

	project, err := h.projectService.SetUnsubscribeTemplate(r.Context(), projectID, userID, req.TemplateID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	if h.Audit != nil {
		h.Audit.LogFromRequest(r, projectID, userID, "project.unsubscribe_template_updated", "project", projectID, map[string]any{"template_id": req.TemplateID})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response.FromProject(project))
}

func (h *ProjectHandler) UpdateSMTP(w http.ResponseWriter, r *http.Request) {
	projectID, userID, ok := requireCap(w, r, h.projectService, service.CapProjectSettings)
	if !ok {
		return
	}

	var req updateSMTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "invalid request body"})
		return
	}

	if req.SmtpHost == "" || req.SmtpPort == 0 || req.SmtpUser == "" || req.SmtpPassword == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "smtp_host, smtp_port, smtp_user, and smtp_password are required"})
		return
	}

	project, err := h.projectService.UpdateSMTP(r.Context(), projectID, userID, req.SmtpHost, req.SmtpPort, req.SmtpUser, req.SmtpPassword, req.FromName, req.FromEmail)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	if h.Audit != nil {
		h.Audit.LogFromRequest(r, projectID, userID, "smtp.update", "project", projectID, map[string]any{"smtp_host": req.SmtpHost, "smtp_user": req.SmtpUser, "from_email": req.FromEmail})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response.FromProject(project))
}

type updateBounceIMAPRequest struct {
	Host     string `json:"host"`
	Port     int32  `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Folder   string `json:"folder"`
	Enabled  bool   `json:"enabled"`
}

func (h *ProjectHandler) GetBounceIMAP(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)
	projectID := r.PathValue("id")
	project, err := h.projectService.GetByID(r.Context(), projectID, userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}
	out := map[string]any{
		"enabled": project.BounceImapEnabled,
		"folder":  project.BounceImapFolder,
	}
	if project.BounceImapHost.Valid {
		out["host"] = project.BounceImapHost.String
	}
	if project.BounceImapPort.Valid {
		out["port"] = project.BounceImapPort.Int32
	}
	if project.BounceImapUser.Valid {
		out["user"] = project.BounceImapUser.String
	}
	out["password_set"] = project.BounceImapPasswordEncrypted.Valid && project.BounceImapPasswordEncrypted.String != ""
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func (h *ProjectHandler) UpdateBounceIMAP(w http.ResponseWriter, r *http.Request) {
	projectID, userID, ok := requireCap(w, r, h.projectService, service.CapProjectSettings)
	if !ok {
		return
	}

	var req updateBounceIMAPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "invalid request body"})
		return
	}
	if req.Folder == "" {
		req.Folder = "INBOX"
	}
	if req.Port == 0 {
		req.Port = 993
	}
	if req.Enabled && (req.Host == "" || req.User == "") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "host and user are required to enable IMAP"})
		return
	}
	project, err := h.projectService.UpdateBounceIMAP(r.Context(), projectID, userID, req.Host, req.Port, req.User, req.Password, req.Folder, req.Enabled)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}
	if h.Audit != nil {
		h.Audit.LogFromRequest(r, projectID, userID, "bounce_imap.update", "project", projectID, map[string]any{"host": req.Host, "user": req.User, "enabled": req.Enabled})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"enabled": project.BounceImapEnabled})
}

func (h *ProjectHandler) GetBounceWebhook(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)
	projectID := r.PathValue("id")
	project, err := h.projectService.GetByID(r.Context(), projectID, userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"project_id":   project.ID.String(),
		"bounce_token": project.BounceToken.String(),
		"path":         "/webhooks/bounces/" + project.ID.String() + "?token=" + project.BounceToken.String(),
	})
}

func (h *ProjectHandler) RotateBounceToken(w http.ResponseWriter, r *http.Request) {
	projectID, userID, ok := requireCap(w, r, h.projectService, service.CapProjectSettings)
	if !ok {
		return
	}
	project, err := h.projectService.RotateBounceToken(r.Context(), projectID, userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}
	if h.Audit != nil {
		h.Audit.LogFromRequest(r, projectID, userID, "bounce_token.rotate", "project", projectID, nil)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"project_id":   project.ID.String(),
		"bounce_token": project.BounceToken.String(),
		"path":         "/webhooks/bounces/" + project.ID.String() + "?token=" + project.BounceToken.String(),
	})
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	projectID, userID, ok := requireCap(w, r, h.projectService, service.CapProjectSettings)
	if !ok {
		return
	}

	err := h.projectService.Delete(r.Context(), projectID, userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResponse{Error: "project not found"})
		return
	}

	if h.Audit != nil {
		h.Audit.LogFromRequest(r, projectID, userID, "project.delete", "project", projectID, nil)
	}

	w.WriteHeader(http.StatusNoContent)
}
