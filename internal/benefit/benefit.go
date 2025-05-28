package benefit

import (
	"errors"
	"giftredeem/internal/db"
	"giftredeem/internal/models"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	// ErrInvalidInput indicates invalid input data
	ErrInvalidInput = errors.New("invalid input data")

	// ErrNotFound indicates a benefit was not found
	ErrNotFound = errors.New("benefit not found")

	// ErrAlreadyClaimed indicates the user has already claimed this benefit
	ErrAlreadyClaimed = errors.New("benefit already claimed by this user")

	// ErrNoCodeAvailable indicates no more codes are available for this benefit
	ErrNoCodeAvailable = errors.New("no codes available for this benefit")

	// ErrBenefitExpired indicates the benefit has expired
	ErrBenefitExpired = errors.New("benefit has expired")

	// ErrBenefitPaused indicates the benefit is currently paused
	ErrBenefitPaused = errors.New("benefit is currently paused")

	// ErrProviderNotAllowed indicates the user's OAuth provider is not allowed for this benefit
	ErrProviderNotAllowed = errors.New("your account provider is not allowed to claim this benefit")

	// ErrAccountTooNew indicates the user's account is too new to claim this benefit
	ErrAccountTooNew = errors.New("your account is too new to claim this benefit")
)

// BenefitService handles benefit operations
type BenefitService struct{}

// NewBenefitService creates a new benefit service
func NewBenefitService() *BenefitService {
	return &BenefitService{}
}

// CreateBenefitInput represents the input for creating a new benefit
type CreateBenefitInput struct {
	Title            string                 `json:"title" binding:"required"`
	Description      string                 `json:"description"`
	Codes            []string               `json:"codes" binding:"required,min=1"`
	ExpiresAt        *time.Time             `json:"expires_at"`
	AllowedProviders []string               `json:"allowed_providers"`
	MinAccountAge    int                    `json:"min_account_age"`
	ClaimConditions  map[string]interface{} `json:"claim_conditions"`
}

// CreateBenefit creates a new benefit with redemption codes
func (s *BenefitService) CreateBenefit(userID uint, input CreateBenefitInput) (*models.Benefit, error) {
	// Validate input
	if input.Title == "" {
		return nil, ErrInvalidInput
	}

	// Clean and validate codes
	cleanedCodes := []string{}
	for _, code := range input.Codes {
		code = strings.TrimSpace(code)
		if code != "" {
			cleanedCodes = append(cleanedCodes, code)
		}
	}

	// Ensure we have at least one valid code
	if len(cleanedCodes) == 0 {
		return nil, ErrInvalidInput
	}

	// Remove duplicates from codes
	uniqueCodes := make(map[string]bool)
	for _, code := range cleanedCodes {
		uniqueCodes[code] = true
	}

	finalCodes := []string{}
	for code := range uniqueCodes {
		finalCodes = append(finalCodes, code)
	}

	// Generate a UUID for the benefit
	benefitUUID := uuid.New().String()

	// Set default expiration time if not provided
	expiresAt := time.Now().AddDate(1, 0, 0) // Default: 1 year from now
	if input.ExpiresAt != nil {
		expiresAt = *input.ExpiresAt
	}

	// Start a transaction
	tx := db.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create the benefit
	benefit := models.Benefit{
		UUID:             benefitUUID,
		Title:            input.Title,
		Description:      input.Description,
		CreatorID:        userID,
		TotalCount:       len(finalCodes),
		ClaimedCount:     0,
		CreatedAt:        time.Now(),
		ExpiresAt:        expiresAt,
		Status:           "active",
		AllowedProviders: input.AllowedProviders,
		MinAccountAge:    input.MinAccountAge,
		ClaimConditions:  input.ClaimConditions,
	}

	if err := tx.Create(&benefit).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Create redemption codes
	for _, code := range finalCodes {
		redemptionCode := models.RedemptionCode{
			BenefitID: benefit.ID,
			Code:      code,
			Status:    "available",
			CreatedAt: time.Now(),
			ClaimedAt: nil, // 显式设置为 nil，表示 NULL
		}

		if err := tx.Create(&redemptionCode).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &benefit, nil
}

// GetBenefitByUUID retrieves a benefit by its UUID
func (s *BenefitService) GetBenefitByUUID(uuid string) (*models.Benefit, error) {
	var benefit models.Benefit
	if err := db.DB.Where("uuid = ?", uuid).First(&benefit).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &benefit, nil
}

// ClaimBenefit allows a user to claim a benefit
func (s *BenefitService) ClaimBenefit(userID uint, benefitUUID string, provider string, ipAddress, userAgent string) (*models.RedemptionCode, error) {
	// Start a transaction
	tx := db.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get the benefit
	var benefit models.Benefit
	if err := tx.Where("uuid = ?", benefitUUID).First(&benefit).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Check if benefit is active
	if benefit.Status != "active" {
		tx.Rollback()
		if benefit.Status == "paused" {
			return nil, ErrBenefitPaused
		} else if benefit.Status == "expired" || time.Now().After(benefit.ExpiresAt) {
			return nil, ErrBenefitExpired
		}
		return nil, ErrNotFound
	}

	// Check provider restrictions
	if len(benefit.AllowedProviders) > 0 {
		allowed := false
		for _, p := range benefit.AllowedProviders {
			if p == provider {
				allowed = true
				break
			}
		}

		if !allowed {
			tx.Rollback()
			return nil, ErrProviderNotAllowed
		}
	}

	// Check account age restriction
	if benefit.MinAccountAge > 0 {
		var user models.User
		if err := tx.First(&user, userID).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		accountAge := int(time.Since(user.CreatedAt).Hours() / 24)
		if accountAge < benefit.MinAccountAge {
			tx.Rollback()
			return nil, ErrAccountTooNew
		}
	}

	// Check if user has already claimed this benefit
	var existingClaim models.Claim
	err := tx.Where("user_id = ? AND benefit_id = ?", userID, benefit.ID).First(&existingClaim).Error
	if err == nil {
		tx.Rollback()
		return nil, ErrAlreadyClaimed
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return nil, err
	}

	// Get an available redemption code
	var code models.RedemptionCode
	if err := tx.Where("benefit_id = ? AND status = ?", benefit.ID, "available").First(&code).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNoCodeAvailable
		}
		return nil, err
	}

	// Update the code status
	code.Status = "claimed"
	code.ClaimedBy = &userID
	now := time.Now()
	code.ClaimedAt = &now
	if err := tx.Save(&code).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Create claim record
	claim := models.Claim{
		UserID:        userID,
		BenefitID:     benefit.ID,
		CodeID:        code.ID,
		OAuthProvider: provider,
		ClaimedAt:     time.Now(),
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
	}

	if err := tx.Create(&claim).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Update claimed count
	benefit.ClaimedCount++
	if err := tx.Save(&benefit).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &code, nil
}

// GetUserBenefits retrieves benefits created by a user
func (s *BenefitService) GetUserBenefits(userID uint) ([]models.Benefit, error) {
	var benefits []models.Benefit
	err := db.DB.Where("creator_id = ?", userID).Order("created_at DESC").Find(&benefits).Error
	return benefits, err
}

// GetUserClaims retrieves benefits claimed by a user
func (s *BenefitService) GetUserClaims(userID uint) ([]models.Claim, error) {
	var claims []models.Claim
	err := db.DB.Where("user_id = ?", userID).
		Preload("Benefit").
		Preload("RedemptionCode").
		Order("claimed_at DESC").
		Find(&claims).Error
	return claims, err
}

// UpdateBenefitStatus updates the status of a benefit
func (s *BenefitService) UpdateBenefitStatus(userID uint, benefitUUID, status string) error {
	if status != "active" && status != "paused" && status != "expired" && status != "deleted" {
		return ErrInvalidInput
	}

	// Get the benefit
	var benefit models.Benefit
	if err := db.DB.Where("uuid = ? AND creator_id = ?", benefitUUID, userID).First(&benefit).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}

	// Update status
	benefit.Status = status
	return db.DB.Save(&benefit).Error
}

// GetBenefitClaims retrieves claims for a specific benefit
func (s *BenefitService) GetBenefitClaims(userID uint, benefitUUID string) ([]models.Claim, error) {
	// Get the benefit
	var benefit models.Benefit
	if err := db.DB.Where("uuid = ? AND creator_id = ?", benefitUUID, userID).First(&benefit).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Get claims
	var claims []models.Claim
	err := db.DB.Where("benefit_id = ?", benefit.ID).
		Preload("User").
		Order("claimed_at DESC").
		Find(&claims).Error

	return claims, err
}

// GetClaimURL generates the claim URL for a benefit
func (s *BenefitService) GetClaimURL(baseURL, benefitUUID string) string {
	return baseURL + "/claim/" + benefitUUID
}
