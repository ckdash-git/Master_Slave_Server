package service

import (
	"errors"
	"time"

	"github.com/cachatto/master-slave-server/internal/config"
	"github.com/cachatto/master-slave-server/internal/models"
	"github.com/cachatto/master-slave-server/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Common errors returned by AuthService.
var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrInvalidTokenType   = errors.New("invalid token type")
)

// TokenPair holds an access token and a refresh token.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// UserProfile is the public profile returned by verify.
type UserProfile struct {
	ID    uuid.UUID   `json:"id"`
	Email string      `json:"email"`
	Apps  []uuid.UUID `json:"authorized_apps"`
}

// JWTClaims are the custom claims embedded in each token.
type JWTClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Type   string    `json:"type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// AuthService handles login, token verification, and token refresh.
type AuthService struct {
	userRepo *repository.UserRepository
	appRepo  *repository.AppRepository
	cfg      *config.Config
}

// NewAuthService creates a new AuthService.
func NewAuthService(userRepo *repository.UserRepository, appRepo *repository.AppRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		appRepo:  appRepo,
		cfg:      cfg,
	}
}

// Login validates credentials and returns a token pair.
func (s *AuthService) Login(email, password string) (*TokenPair, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.generateTokenPair(user)
}

// VerifyToken parses an access token and returns the user profile with permitted apps.
func (s *AuthService) VerifyToken(tokenString string) (*UserProfile, error) {
	claims, err := s.parseToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.Type != "access" {
		return nil, ErrInvalidTokenType
	}

	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	apps, err := s.appRepo.GetPermittedApps(user.ID)
	if err != nil {
		return nil, err
	}

	appIDs := make([]uuid.UUID, len(apps))
	for i, app := range apps {
		appIDs[i] = app.ID
	}

	return &UserProfile{
		ID:    user.ID,
		Email: user.Email,
		Apps:  appIDs,
	}, nil
}

// RefreshToken validates a refresh token and returns a new token pair.
func (s *AuthService) RefreshToken(refreshToken string) (*TokenPair, error) {
	claims, err := s.parseToken(refreshToken)
	if err != nil {
		return nil, err
	}

	if claims.Type != "refresh" {
		return nil, ErrInvalidTokenType
	}

	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return s.generateTokenPair(user)
}

// GenerateTokenPairForUser creates a token pair for a given user (used by OTC service).
func (s *AuthService) GenerateTokenPairForUser(userID uuid.UUID) (*TokenPair, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return s.generateTokenPair(user)
}

// ParseAccessToken parses and validates an access token, returning the claims.
func (s *AuthService) ParseAccessToken(tokenString string) (*JWTClaims, error) {
	claims, err := s.parseToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.Type != "access" {
		return nil, ErrInvalidTokenType
	}
	return claims, nil
}

// generateTokenPair creates both access and refresh tokens for a user.
func (s *AuthService) generateTokenPair(user *models.User) (*TokenPair, error) {
	now := time.Now()

	// Access token
	accessClaims := JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.JWTAccessExpiry)),
			Issuer:    "master-slave-server",
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, err
	}

	// Refresh token
	refreshClaims := JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Type:   "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.JWTRefreshExpiry)),
			Issuer:    "master-slave-server",
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshStr, err := refreshToken.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
	}, nil
}

// parseToken parses and validates a JWT token string.
func (s *AuthService) parseToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
