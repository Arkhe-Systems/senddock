package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/arkhe-systems/senddock/internal/response"
	"github.com/arkhe-systems/senddock/internal/service"
	"github.com/arkhe-systems/senddock/pkg/auth"
	"github.com/google/uuid"
)

type WorkspaceHandler struct {
	svc            *service.WorkspaceService
	projectService *service.ProjectService
	Audit          *service.AuditService
}

func NewWorkspaceHandler(svc *service.WorkspaceService, projectService *service.ProjectService) *WorkspaceHandler {
	return &WorkspaceHandler{svc: svc, projectService: projectService}
}

type workspaceResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedBy string `json:"created_by"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Role      string `json:"role,omitempty"`
}

func (h *WorkspaceHandler) List(w http.ResponseWriter, r *http.Request) {
	_, uid, ok := requireUser(w, r)
	if !ok {
		return
	}

	rows, err := h.svc.ListByUser(r.Context(), uid)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]workspaceResponse, 0, len(rows))
	for _, ws := range rows {
		out = append(out, workspaceResponse{
			ID:        ws.ID.String(),
			Name:      ws.Name,
			CreatedBy: ws.CreatedBy.String(),
			CreatedAt: ws.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: ws.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"workspaces": out})
}

type createWorkspaceRequest struct {
	Name string `json:"name"`
}

func (h *WorkspaceHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, uid, ok := requireUser(w, r)
	if !ok {
		return
	}

	var req createWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ws, err := h.svc.Create(r.Context(), uid, req.Name)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if h.Audit != nil {
		h.Audit.LogFromRequest(r, "", userID, "workspace.create", "workspace", ws.ID.String(), map[string]any{"name": ws.Name})
	}

	writeJSON(w, http.StatusCreated, workspaceResponse{
		ID:        ws.ID.String(),
		Name:      ws.Name,
		CreatedBy: ws.CreatedBy.String(),
		CreatedAt: ws.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: ws.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		Role:      service.WorkspaceRoleOwner,
	})
}

type renameWorkspaceRequest struct {
	Name string `json:"name"`
}

func (h *WorkspaceHandler) Rename(w http.ResponseWriter, r *http.Request) {
	userID, uid, ok := requireUser(w, r)
	if !ok {
		return
	}
	wid, ok := requireUUID(w, r, "id", "invalid workspace id")
	if !ok {
		return
	}

	var req renameWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ws, err := h.svc.Rename(r.Context(), wid, uid, req.Name)
	if err != nil {
		writeWorkspaceServiceError(w, err)
		return
	}

	if h.Audit != nil {
		h.Audit.LogFromRequest(r, "", userID, "workspace.rename", "workspace", ws.ID.String(), map[string]any{"name": ws.Name})
	}

	writeJSON(w, http.StatusOK, workspaceResponse{
		ID:        ws.ID.String(),
		Name:      ws.Name,
		CreatedBy: ws.CreatedBy.String(),
		CreatedAt: ws.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: ws.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *WorkspaceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, uid, ok := requireUser(w, r)
	if !ok {
		return
	}
	wid, ok := requireUUID(w, r, "id", "invalid workspace id")
	if !ok {
		return
	}

	if err := h.svc.Delete(r.Context(), wid, uid); err != nil {
		writeWorkspaceServiceError(w, err)
		return
	}

	if h.Audit != nil {
		h.Audit.LogFromRequest(r, "", userID, "workspace.delete", "workspace", wid.String(), nil)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *WorkspaceHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := requireUser(w, r)
	if !ok {
		return
	}
	wid, ok := requireUUID(w, r, "id", "invalid workspace id")
	if !ok {
		return
	}

	projects, err := h.projectService.ListByWorkspace(r.Context(), wid.String(), userID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response.FromProjects(projects))
}

func (h *WorkspaceHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	_, uid, ok := requireUser(w, r)
	if !ok {
		return
	}
	wid, ok := requireUUID(w, r, "id", "invalid workspace id")
	if !ok {
		return
	}

	members, err := h.svc.ListMembers(r.Context(), wid, uid)
	if err != nil {
		writeWorkspaceServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"members": members})
}

type addMemberRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (h *WorkspaceHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	userID, uid, ok := requireUser(w, r)
	if !ok {
		return
	}
	wid, ok := requireUUID(w, r, "id", "invalid workspace id")
	if !ok {
		return
	}

	var req addMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	member, err := h.svc.AddMember(r.Context(), wid, uid, req.Email, req.Role)
	if err != nil {
		writeWorkspaceServiceError(w, err)
		return
	}

	if h.Audit != nil {
		h.Audit.LogFromRequest(r, "", userID, "workspace.member_added", "workspace", wid.String(), map[string]any{"member_id": member.UserID.String(), "role": member.Role})
	}

	writeJSON(w, http.StatusCreated, member)
}

type createUserRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

func (h *WorkspaceHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	userID, uid, ok := requireUser(w, r)
	if !ok {
		return
	}
	wid, ok := requireUUID(w, r, "id", "invalid workspace id")
	if !ok {
		return
	}

	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	created, err := h.svc.CreateUserAndAddMember(r.Context(), wid, uid, req.Email, req.Name, req.Password, req.Role)
	if err != nil {
		writeWorkspaceServiceError(w, err)
		return
	}

	if h.Audit != nil {
		h.Audit.LogFromRequest(r, "", userID, "workspace.user_created", "workspace", wid.String(), map[string]any{"member_id": created.UserID.String(), "email": created.Email, "role": created.Role})
	}

	writeJSON(w, http.StatusCreated, created)
}

type updateMemberRequest struct {
	Role string `json:"role"`
}

func (h *WorkspaceHandler) UpdateMember(w http.ResponseWriter, r *http.Request) {
	userID, uid, ok := requireUser(w, r)
	if !ok {
		return
	}
	wid, ok := requireUUID(w, r, "id", "invalid workspace id")
	if !ok {
		return
	}
	target, ok := requireUUID(w, r, "userId", "invalid user id")
	if !ok {
		return
	}

	var req updateMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.svc.UpdateMemberRole(r.Context(), wid, uid, target, req.Role); err != nil {
		writeWorkspaceServiceError(w, err)
		return
	}

	if h.Audit != nil {
		h.Audit.LogFromRequest(r, "", userID, "workspace.member_role_changed", "workspace", wid.String(), map[string]any{"member_id": target.String(), "role": req.Role})
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *WorkspaceHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	userID, uid, ok := requireUser(w, r)
	if !ok {
		return
	}
	wid, ok := requireUUID(w, r, "id", "invalid workspace id")
	if !ok {
		return
	}
	target, ok := requireUUID(w, r, "userId", "invalid user id")
	if !ok {
		return
	}

	if err := h.svc.RemoveMember(r.Context(), wid, uid, target); err != nil {
		writeWorkspaceServiceError(w, err)
		return
	}

	if h.Audit != nil {
		h.Audit.LogFromRequest(r, "", userID, "workspace.member_removed", "workspace", wid.String(), map[string]any{"member_id": target.String()})
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeWorkspaceServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrWorkspaceNotFound):
		writeJSONError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrWorkspaceForbidden):
		writeJSONError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, service.ErrWorkspaceOwnerRequired):
		writeJSONError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, service.ErrWorkspaceHasProjects):
		writeJSONError(w, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrLastOwner):
		writeJSONError(w, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrUserNotFound):
		writeJSONError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrInvalidRole):
		writeJSONError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrEmailTaken):
		writeJSONError(w, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrWorkspaceMembersLicense):
		writeJSONError(w, http.StatusPaymentRequired, err.Error())
	default:
		writeJSONError(w, http.StatusInternalServerError, err.Error())
	}
}

func requireUser(w http.ResponseWriter, r *http.Request) (string, uuid.UUID, bool) {
	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "missing user")
		return "", uuid.Nil, false
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "invalid user")
		return "", uuid.Nil, false
	}
	return userID, uid, true
}

func requireUUID(w http.ResponseWriter, r *http.Request, name, msg string) (uuid.UUID, bool) {
	id, err := uuid.Parse(r.PathValue(name))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, msg)
		return uuid.Nil, false
	}
	return id, true
}

