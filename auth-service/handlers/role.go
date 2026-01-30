package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"auth-service/models"
)

type RoleHandler struct {
	roleModel *models.RoleModel
}

func NewRoleHandler(roleModel *models.RoleModel) *RoleHandler {
	return &RoleHandler{roleModel: roleModel}
}

// GetUserRoles returns all roles for a specific user
func (h *RoleHandler) GetUserRoles(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	roles, err := h.roleModel.GetUserRoles(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to get user roles", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id": userID,
		"roles":   roles,
	})
}

// AssignRoleToUser assigns a role to a user
func (h *RoleHandler) AssignRoleToUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID int64 `json:"user_id"`
		RoleID int64 `json:"role_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == 0 || req.RoleID == 0 {
		http.Error(w, "user_id and role_id are required", http.StatusBadRequest)
		return
	}

	err := h.roleModel.AssignRoleToUser(r.Context(), req.UserID, req.RoleID)
	if err != nil {
		http.Error(w, "failed to assign role", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "role assigned successfully",
	})
}

// RemoveRoleFromUser removes a role from a user
func (h *RoleHandler) RemoveRoleFromUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID int64 `json:"user_id"`
		RoleID int64 `json:"role_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == 0 || req.RoleID == 0 {
		http.Error(w, "user_id and role_id are required", http.StatusBadRequest)
		return
	}

	err := h.roleModel.RemoveRoleFromUser(r.Context(), req.UserID, req.RoleID)
	if err != nil {
		http.Error(w, "failed to remove role", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "role removed successfully",
	})
}

// CheckUserRole checks if a user has a specific role
func (h *RoleHandler) CheckUserRole(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	roleName := r.URL.Query().Get("role")

	if userIDStr == "" || roleName == "" {
		http.Error(w, "user_id and role are required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	hasRole, err := h.roleModel.HasRole(r.Context(), userID, roleName)
	if err != nil {
		http.Error(w, "failed to check role", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":  userID,
		"role":     roleName,
		"has_role": hasRole,
	})
}
