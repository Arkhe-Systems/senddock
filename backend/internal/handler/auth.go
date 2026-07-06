package handler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/arkhe-systems/senddock/internal/service"
)

type DeviceGate interface {
	CheckLogin(w http.ResponseWriter, r *http.Request, userID, email string) (bool, error)
}

// LoginLimiter throttles brute-force attempts per account (independent of the
// per-IP rate limiter, which can be evaded by spoofing X-Forwarded-For).
type LoginLimiter interface {
	Count(ctx context.Context, key string) int64
	Increment(ctx context.Context, key string, ttl time.Duration) (int64, error)
	Delete(ctx context.Context, keys ...string)
}

const (
	maxLoginAttempts = 10
	max2FAAttempts   = 10
	lockoutWindow    = 15 * time.Minute
)

type AuthHandler struct {
	authService *service.AuthService
	deviceGate  DeviceGate
	limiter     LoginLimiter
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) SetDeviceGate(g DeviceGate) {
	h.deviceGate = g
}

func (h *AuthHandler) SetLoginLimiter(l LoginLimiter) {
	h.limiter = l
}

func (h *AuthHandler) locked(ctx context.Context, key string, max int64) bool {
	return h.limiter != nil && h.limiter.Count(ctx, key) >= max
}

func (h *AuthHandler) recordFailure(ctx context.Context, key string) {
	if h.limiter != nil {
		h.limiter.Increment(ctx, key, lockoutWindow)
	}
}

func (h *AuthHandler) clearFailures(ctx context.Context, key string) {
	if h.limiter != nil {
		h.limiter.Delete(ctx, key)
	}
}

func lockoutKey(prefix, id string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(id))))
	return prefix + hex.EncodeToString(sum[:8])
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type errorResponse struct {
	Error string `json:"error"`
}

var secureCookies = os.Getenv("FRONTEND_URL") != "" && strings.HasPrefix(os.Getenv("FRONTEND_URL"), "https")

func setAuthCookies(w http.ResponseWriter, tokens service.AuthTokens) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    tokens.AccessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   secureCookies,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   900,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		Path:     "/api/v1/auth",
		HttpOnly: true,
		Secure:   secureCookies,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   7200,
	})
}

func SetAuthCookies(w http.ResponseWriter, tokens service.AuthTokens) {
	setAuthCookies(w, tokens)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "invalid request body"})
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "email, password and name are required"})
		return
	}

	if !isValidEmail(req.Email) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "invalid email format"})
		return
	}

	if !isValidPassword(req.Password) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "password must be at least 8 characters"})
		return
	}

	tokens, err := h.authService.Register(r.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(errorResponse{Error: "email already registered"})
		return
	}

	setAuthCookies(w, tokens)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "registered"})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "invalid request body"})
		return
	}

	if req.Email == "" || req.Password == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "email and password are required"})
		return
	}

	failKey := lockoutKey("login_fail:", req.Email)
	if h.locked(r.Context(), failKey, maxLoginAttempts) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(errorResponse{Error: "too many failed attempts, try again later"})
		return
	}

	result, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if errors.Is(err, service.ErrEmailNotVerified) {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "email not verified", "code": "email_not_verified"})
			return
		}
		h.recordFailure(r.Context(), failKey)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	h.clearFailures(r.Context(), failKey)

	w.Header().Set("Content-Type", "application/json")
	if result.Requires2FA {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"requires_2fa":     true,
			"two_factor_token": result.TwoFactorToken,
		})
		return
	}

	if h.deviceGate != nil {
		proceed, err := h.deviceGate.CheckLogin(w, r, result.UserID, req.Email)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(errorResponse{Error: "could not verify device"})
			return
		}
		if !proceed {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]bool{"requires_device_confirmation": true})
			return
		}
	}

	setAuthCookies(w, result.Tokens)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "logged in"})
}

type twoFactorRequest struct {
	TwoFactorToken string `json:"two_factor_token"`
	Code           string `json:"code"`
}

func (h *AuthHandler) VerifyTwoFactor(w http.ResponseWriter, r *http.Request) {
	var req twoFactorRequest
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "invalid request body"})
		return
	}
	if req.TwoFactorToken == "" || req.Code == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "two_factor_token and code are required"})
		return
	}

	failKey := lockoutKey("2fa_fail:", req.TwoFactorToken)
	if h.locked(r.Context(), failKey, max2FAAttempts) {
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(errorResponse{Error: "too many failed attempts, sign in again"})
		return
	}

	tokens, err := h.authService.VerifyTwoFactor(r.Context(), req.TwoFactorToken, req.Code)
	if err != nil {
		h.recordFailure(r.Context(), failKey)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	h.clearFailures(r.Context(), failKey)
	setAuthCookies(w, tokens)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "logged in"})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(errorResponse{Error: "no refresh token"})
		return
	}

	tokens, err := h.authService.Refresh(r.Context(), cookie.Value)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(errorResponse{Error: "invalid refresh token"})
		return
	}

	setAuthCookies(w, tokens)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "refreshed"})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(errorResponse{Error: "no refresh token"})
		return
	}

	err = h.authService.Logout(r.Context(), cookie.Value)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{Error: "Internal server error"})
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "access_token", Path: "/", MaxAge: -1})
	http.SetCookie(w, &http.Cookie{Name: "refresh_token", Path: "/api/v1/auth", MaxAge: -1})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "logged out"})
}
