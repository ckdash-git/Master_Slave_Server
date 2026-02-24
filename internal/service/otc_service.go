package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/cachatto/master-slave-server/internal/config"
	"github.com/cachatto/master-slave-server/internal/models"
	"github.com/cachatto/master-slave-server/internal/repository"
	"github.com/google/uuid"
)

// Common errors returned by OTCService.
var (
	ErrCodeExpired    = errors.New("code expired or already claimed")
	ErrAppNotFound    = errors.New("application not found")
	ErrAppMismatch    = errors.New("code does not match the requested application")
	ErrNoPermission   = errors.New("user does not have permission for this application")
	ErrCodeGeneration = errors.New("failed to generate one-time code")
)

// OTCResult is returned when an OTC is successfully created.
type OTCResult struct {
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
}

// OTCService handles the one-time-code handshake between Master and Slave apps.
type OTCService struct {
	otcRepo     *repository.OTCRepository
	appRepo     *repository.AppRepository
	authService *AuthService
	cfg         *config.Config
}

// NewOTCService creates a new OTCService.
func NewOTCService(
	otcRepo *repository.OTCRepository,
	appRepo *repository.AppRepository,
	authService *AuthService,
	cfg *config.Config,
) *OTCService {
	return &OTCService{
		otcRepo:     otcRepo,
		appRepo:     appRepo,
		authService: authService,
		cfg:         cfg,
	}
}

// ExchangeCode generates a short-lived one-time code for a specific slave app.
func (s *OTCService) ExchangeCode(userID, appID uuid.UUID) (*OTCResult, error) {
	// Verify the app exists
	_, err := s.appRepo.FindByID(appID)
	if err != nil {
		return nil, ErrAppNotFound
	}

	// Verify user has permission for this app
	hasPermission, err := s.appRepo.HasPermission(userID, appID)
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrNoPermission
	}

	// Generate a cryptographically random code (6 bytes → 12 hex chars)
	codeBytes := make([]byte, 6)
	if _, err := rand.Read(codeBytes); err != nil {
		return nil, ErrCodeGeneration
	}
	code := hex.EncodeToString(codeBytes)

	expiresAt := time.Now().Add(s.cfg.OTCExpiry)

	otc := &models.OneTimeCode{
		UserID:    userID,
		AppID:     appID,
		Code:      code,
		ExpiresAt: expiresAt,
		Claimed:   false,
	}

	if err := s.otcRepo.Create(otc); err != nil {
		return nil, err
	}

	return &OTCResult{
		Code:      code,
		ExpiresAt: expiresAt,
	}, nil
}

// ClaimToken validates a one-time code and returns a JWT token pair.
func (s *OTCService) ClaimToken(code, packageID string) (*TokenPair, error) {
	// Find the code
	otc, err := s.otcRepo.FindByCode(code)
	if err != nil {
		return nil, ErrCodeExpired
	}

	// Check if already claimed or expired
	if otc.Claimed || time.Now().After(otc.ExpiresAt) {
		return nil, ErrCodeExpired
	}

	// Verify the package ID matches the app associated with the code
	app, err := s.appRepo.FindByPackageID(packageID)
	if err != nil {
		return nil, ErrAppNotFound
	}

	if app.ID != otc.AppID {
		return nil, ErrAppMismatch
	}

	// Mark the code as claimed
	if err := s.otcRepo.MarkClaimed(otc); err != nil {
		return nil, err
	}

	// Generate a token pair for the user
	return s.authService.GenerateTokenPairForUser(otc.UserID)
}

// CleanExpiredCodes removes all expired or claimed codes (call periodically).
func (s *OTCService) CleanExpiredCodes() error {
	return s.otcRepo.CleanExpired()
}
