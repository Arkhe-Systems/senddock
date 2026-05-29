package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
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

func (s *AuthService) Login(ctx context.Context, email, password string) (AuthTokens, error) {
	user, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return AuthTokens{}, errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(password))
	if err != nil {
		return AuthTokens{}, errors.New("invalid credentials")
	}

	token, err := s.generateTokens(ctx, user.ID)
	if err != nil {
		return AuthTokens{}, err
	}

	return token, nil
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
