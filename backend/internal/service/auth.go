package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"
	"time"
	"unicode"

	"github.com/arkhe-systems/senddock/internal/db"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	var hasUpper, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		case !unicode.IsLetter(r) && !unicode.IsDigit(r) && !unicode.IsSpace(r):
			hasSpecial = true
		}
	}
	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasDigit {
		return errors.New("password must contain at least one number")
	}
	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}
	return nil
}

type AuthTokens struct {
	AccessToken  string
	RefreshToken string
}
type AuthService struct {
	queries   *db.Queries
	jwtSecret []byte
}

func NewAuthService(queries *db.Queries, jwtSecret string) *AuthService {
	return &AuthService{
		queries:   queries,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, name string) (AuthTokens, error) {
	if err := ValidatePassword(password); err != nil {
		return AuthTokens{}, err
	}

	_, err := s.queries.GetUserByEmail(ctx, email)
	if err == nil {
		return AuthTokens{}, errors.New("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return AuthTokens{}, err
	}

	user, err := s.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		PasswordHash: sql.NullString{String: string(hash), Valid: true},
		Name:         name,
	})

	if err != nil {
		return AuthTokens{}, err
	}

	if err := s.bootstrapDefaultWorkspace(ctx, user.ID); err != nil {
		return AuthTokens{}, err
	}

	return s.generateTokens(ctx, user.ID)
}

func (s *AuthService) bootstrapDefaultWorkspace(ctx context.Context, userID uuid.UUID) error {
	ws, err := s.queries.CreateWorkspace(ctx, db.CreateWorkspaceParams{
		Name:      "My Workspace",
		CreatedBy: userID,
	})
	if err != nil {
		return err
	}
	_, err = s.queries.AddWorkspaceMember(ctx, db.AddWorkspaceMemberParams{
		WorkspaceID: ws.ID,
		UserID:      userID,
		Role:        WorkspaceRoleOwner,
		InvitedBy:   uuid.NullUUID{UUID: userID, Valid: true},
	})
	return err
}

func (s *AuthService) EnsureUser(ctx context.Context, email, name string) (string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if user, err := s.queries.GetUserByEmail(ctx, email); err == nil {
		return user.ID.String(), nil
	}
	user, err := s.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		PasswordHash: sql.NullString{Valid: false},
		Name:         name,
	})
	if err != nil {
		return "", err
	}
	if err := s.bootstrapDefaultWorkspace(ctx, user.ID); err != nil {
		return "", err
	}
	return user.ID.String(), nil
}

type LoginResult struct {
	Tokens         AuthTokens
	Requires2FA    bool
	TwoFactorToken string
}

func (s *AuthService) Login(ctx context.Context, email, password string) (LoginResult, error) {
	user, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return LoginResult{}, errors.New("invalid credentials")
	}

	if !user.PasswordHash.Valid {
		return LoginResult{}, errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(password))
	if err != nil {
		return LoginResult{}, errors.New("invalid credentials")
	}

	if user.TotpEnabled {
		twoFAToken, err := s.generateTwoFactorToken(user.ID)
		if err != nil {
			return LoginResult{}, err
		}
		return LoginResult{Requires2FA: true, TwoFactorToken: twoFAToken}, nil
	}

	tokens, err := s.generateTokens(ctx, user.ID)
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{Tokens: tokens}, nil
}

func (s *AuthService) VerifyTwoFactor(ctx context.Context, twoFAToken, code string) (AuthTokens, error) {
	userID, err := s.parseTwoFactorToken(twoFAToken)
	if err != nil {
		return AuthTokens{}, errors.New("two-factor session expired, please sign in again")
	}

	user, err := s.queries.GetUserById(ctx, userID)
	if err != nil || !user.TotpEnabled || user.TotpSecret.String == "" {
		return AuthTokens{}, errors.New("two-factor not configured for this account")
	}

	clean := strings.TrimSpace(code)
	if IsLikelyRecoveryCode(clean) {
		if err := s.consumeRecoveryCode(ctx, userID, clean); err != nil {
			return AuthTokens{}, ErrInvalidRecoveryCode
		}
	} else {
		if !ValidateTOTPCode(user.TotpSecret.String, clean) {
			return AuthTokens{}, ErrInvalidTOTPCode
		}
	}

	return s.generateTokens(ctx, userID)
}

func (s *AuthService) consumeRecoveryCode(ctx context.Context, userID uuid.UUID, raw string) error {
	clean := normalizeRecoveryCode(raw)
	rows, err := s.queries.ListUserRecoveryCodes(ctx, userID)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if bcrypt.CompareHashAndPassword([]byte(row.CodeHash), []byte(clean)) == nil {
			return s.queries.MarkRecoveryCodeUsed(ctx, row.ID)
		}
	}
	return ErrInvalidRecoveryCode
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (AuthTokens, error) {
	tokenHash := hashToken(refreshToken)

	stored, err := s.queries.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		return AuthTokens{}, errors.New("invalid refresh token")
	}

	err = s.queries.DeleteRefreshToken(ctx, stored.TokenHash)
	if err != nil {
		return AuthTokens{}, err
	}
	return s.generateTokens(ctx, stored.UserID)
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	tokenHash := hashToken(refreshToken)
	return s.queries.DeleteRefreshToken(ctx, tokenHash)
}

func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	user, err := s.queries.GetUserById(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}

	if !user.PasswordHash.Valid {
		return errors.New("password change not available for this account")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(currentPassword)); err != nil {
		return errors.New("current password is incorrect")
	}

	if err := ValidatePassword(newPassword); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.queries.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:           userID,
		PasswordHash: sql.NullString{String: string(hash), Valid: true},
	})
}

func (s *AuthService) generateTokens(ctx context.Context, userID uuid.UUID) (AuthTokens, error) {
	accessToken, err := s.generateAccessToken(userID)
	if err != nil {
		return AuthTokens{}, err
	}

	refreshToken, err := generateRandomToken()
	if err != nil {
		return AuthTokens{}, err
	}

	tokenHash := hashToken(refreshToken)
	expiredAt := time.Now().Add(2 * time.Hour)

	_, err = s.queries.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiredAt,
	})
	if err != nil {
		return AuthTokens{}, err
	}

	return AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) generateAccessToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"exp": time.Now().Add(15 * time.Minute).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *AuthService) IssueTokens(ctx context.Context, userID uuid.UUID) (AuthTokens, error) {
	return s.generateTokens(ctx, userID)
}

func generateRandomToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (s *AuthService) generateTwoFactorToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID.String(),
		"step": "2fa_pending",
		"exp":  time.Now().Add(5 * time.Minute).Unix(),
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *AuthService) parseTwoFactorToken(raw string) (uuid.UUID, error) {
	token, err := jwt.Parse(raw, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, errors.New("invalid claims")
	}
	step, _ := claims["step"].(string)
	if step != "2fa_pending" {
		return uuid.Nil, errors.New("wrong token type")
	}
	sub, _ := claims["sub"].(string)
	return uuid.Parse(sub)
}

func (s *AuthService) SetupTOTP(ctx context.Context, userID uuid.UUID) (TOTPSetup, error) {
	user, err := s.queries.GetUserById(ctx, userID)
	if err != nil {
		return TOTPSetup{}, errors.New("user not found")
	}
	if user.TotpEnabled {
		return TOTPSetup{}, errors.New("2fa already enabled — disable it first to regenerate")
	}

	setup, err := GenerateTOTPSetup(user.Email)
	if err != nil {
		return TOTPSetup{}, err
	}

	if err := s.queries.SetUserTotpSecret(ctx, db.SetUserTotpSecretParams{
		ID:         userID,
		TotpSecret: sql.NullString{String: setup.Secret, Valid: true},
	}); err != nil {
		return TOTPSetup{}, err
	}

	if err := s.replaceRecoveryCodes(ctx, userID, setup.RecoveryCodes); err != nil {
		return TOTPSetup{}, err
	}
	return setup, nil
}

func (s *AuthService) VerifyTOTPSetup(ctx context.Context, userID uuid.UUID, code string) error {
	user, err := s.queries.GetUserById(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}
	if !user.TotpSecret.Valid || user.TotpSecret.String == "" {
		return errors.New("2fa setup not started — call setup first")
	}
	if user.TotpEnabled {
		return errors.New("2fa already enabled")
	}
	if !ValidateTOTPCode(user.TotpSecret.String, code) {
		return ErrInvalidTOTPCode
	}
	return s.queries.EnableUserTotp(ctx, userID)
}

func (s *AuthService) DisableTOTP(ctx context.Context, userID uuid.UUID, password, code string) error {
	user, err := s.queries.GetUserById(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}
	if !user.TotpEnabled {
		return errors.New("2fa is not enabled")
	}
	if !user.PasswordHash.Valid {
		return errors.New("cannot disable 2fa on a passwordless account")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(password)); err != nil {
		return errors.New("current password is incorrect")
	}

	clean := strings.TrimSpace(code)
	if IsLikelyRecoveryCode(clean) {
		if err := s.consumeRecoveryCode(ctx, userID, clean); err != nil {
			return ErrInvalidRecoveryCode
		}
	} else {
		if !ValidateTOTPCode(user.TotpSecret.String, clean) {
			return ErrInvalidTOTPCode
		}
	}

	if err := s.queries.DisableUserTotp(ctx, userID); err != nil {
		return err
	}
	return s.queries.DeleteUserRecoveryCodes(ctx, userID)
}

func (s *AuthService) RegenerateRecoveryCodes(ctx context.Context, userID uuid.UUID, code string) ([]string, error) {
	user, err := s.queries.GetUserById(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if !user.TotpEnabled || user.TotpSecret.String == "" {
		return nil, errors.New("2fa is not enabled")
	}
	if !ValidateTOTPCode(user.TotpSecret.String, strings.TrimSpace(code)) {
		return nil, ErrInvalidTOTPCode
	}

	codes, err := generateRecoveryCodes(recoveryCodeN)
	if err != nil {
		return nil, err
	}
	if err := s.replaceRecoveryCodes(ctx, userID, codes); err != nil {
		return nil, err
	}
	return codes, nil
}

func (s *AuthService) replaceRecoveryCodes(ctx context.Context, userID uuid.UUID, codes []string) error {
	if err := s.queries.DeleteUserRecoveryCodes(ctx, userID); err != nil {
		return err
	}
	for _, raw := range codes {
		hash, err := bcrypt.GenerateFromPassword([]byte(normalizeRecoveryCode(raw)), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		if err := s.queries.CreateRecoveryCode(ctx, db.CreateRecoveryCodeParams{
			UserID:   userID,
			CodeHash: string(hash),
		}); err != nil {
			return err
		}
	}
	return nil
}
