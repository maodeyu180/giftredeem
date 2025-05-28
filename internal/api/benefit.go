package api

import (
	"errors"
	"fmt"
	benefitpkg "giftredeem/internal/benefit"
	"giftredeem/internal/db"
	"giftredeem/internal/models"
	"giftredeem/internal/response"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// BenefitHandler handles benefit-related requests
type BenefitHandler struct {
	benefitService *benefitpkg.BenefitService
}

// NewBenefitHandler creates a new benefit handler
func NewBenefitHandler() *BenefitHandler {
	return &BenefitHandler{
		benefitService: benefitpkg.NewBenefitService(),
	}
}

// CreateBenefit handles the creation of a new benefit
func (h *BenefitHandler) CreateBenefit(c *gin.Context) {
	// Get user from context (set by auth middleware)
	userValue, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusOK, response.Error(response.CodeUnauthorized, "User not authenticated"))
		return
	}

	user := userValue.(*models.User)

	// Parse request body
	var input benefitpkg.CreateBenefitInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusOK, response.Error(response.CodeInvalidInput, "Invalid input: "+err.Error()))
		return
	}

	// Create benefit
	newBenefit, err := h.benefitService.CreateBenefit(user.ID, input)
	if err != nil {
		code := response.CodeBenefitCreationFailed
		if errors.Is(err, benefitpkg.ErrInvalidInput) {
			code = response.CodeInvalidInput
		}

		c.JSON(http.StatusOK, response.Error(code, "Failed to create benefit: "+err.Error()))
		return
	}

	// Generate claim URL
	baseURL := fmt.Sprintf("%s://%s", c.Request.URL.Scheme, c.Request.Host)
	if c.Request.URL.Scheme == "" {
		baseURL = "http://" + c.Request.Host // Default to HTTP if scheme is empty
	}

	claimURL := h.benefitService.GetClaimURL(baseURL, newBenefit.UUID)

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"benefit": map[string]interface{}{
			"id":                newBenefit.ID,
			"uuid":              newBenefit.UUID,
			"title":             newBenefit.Title,
			"description":       newBenefit.Description,
			"total_count":       newBenefit.TotalCount,
			"claimed_count":     newBenefit.ClaimedCount,
			"created_at":        newBenefit.CreatedAt,
			"expires_at":        newBenefit.ExpiresAt,
			"status":            newBenefit.Status,
			"min_account_age":   newBenefit.MinAccountAge,
			"allowed_providers": newBenefit.AllowedProviders,
		},
		"claim_url": claimURL,
	}))
}

// GetUserBenefits retrieves all benefits created by the current user
func (h *BenefitHandler) GetUserBenefits(c *gin.Context) {
	// Get user from context (set by auth middleware)
	userValue, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusOK, response.Error(response.CodeUnauthorized, "User not authenticated"))
		return
	}

	user := userValue.(*models.User)

	// Get benefits
	benefits, err := h.benefitService.GetUserBenefits(user.ID)
	if err != nil {
		c.JSON(http.StatusOK, response.Error(response.CodeServerError, "Failed to retrieve benefits: "+err.Error()))
		return
	}

	// Generate base URL for claim links
	baseURL := fmt.Sprintf("%s://%s", c.Request.URL.Scheme, c.Request.Host)
	if c.Request.URL.Scheme == "" {
		baseURL = "http://" + c.Request.Host // Default to HTTP if scheme is empty
	}

	// Format response
	responseData := make([]map[string]interface{}, len(benefits))
	for i, b := range benefits {
		responseData[i] = map[string]interface{}{
			"id":                b.ID,
			"uuid":              b.UUID,
			"title":             b.Title,
			"description":       b.Description,
			"total_count":       b.TotalCount,
			"claimed_count":     b.ClaimedCount,
			"created_at":        b.CreatedAt,
			"expires_at":        b.ExpiresAt,
			"status":            b.Status,
			"claim_url":         h.benefitService.GetClaimURL(baseURL, b.UUID),
			"allowed_providers": b.AllowedProviders,
			"min_account_age":   b.MinAccountAge,
		}
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"benefits": responseData,
	}))
}

// GetUserClaims retrieves all benefits claimed by the current user
func (h *BenefitHandler) GetUserClaims(c *gin.Context) {
	// Get user from context (set by auth middleware)
	userValue, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusOK, response.Error(response.CodeUnauthorized, "User not authenticated"))
		return
	}

	user := userValue.(*models.User)

	// Get claims
	claims, err := h.benefitService.GetUserClaims(user.ID)
	if err != nil {
		c.JSON(http.StatusOK, response.Error(response.CodeServerError, "Failed to retrieve claims: "+err.Error()))
		return
	}

	// Format response
	responseData := make([]map[string]interface{}, len(claims))
	for i, claim := range claims {
		responseData[i] = map[string]interface{}{
			"id":             claim.ID,
			"claimed_at":     claim.ClaimedAt,
			"oauth_provider": claim.OAuthProvider,
			"benefit": map[string]interface{}{
				"id":          claim.Benefit.ID,
				"uuid":        claim.Benefit.UUID,
				"title":       claim.Benefit.Title,
				"description": claim.Benefit.Description,
			},
			"code": claim.RedemptionCode.Code,
		}
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"claims": responseData,
	}))
}

// UpdateBenefitStatus updates the status of a benefit
func (h *BenefitHandler) UpdateBenefitStatus(c *gin.Context) {
	// Get user from context (set by auth middleware)
	userValue, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusOK, response.Error(response.CodeUnauthorized, "User not authenticated"))
		return
	}

	user := userValue.(*models.User)

	// Get benefit UUID from path
	benefitUUID := c.Param("uuid")
	if benefitUUID == "" {
		c.JSON(http.StatusOK, response.Error(response.CodeInvalidInput, "Benefit UUID is required"))
		return
	}

	// Parse request body
	var input struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusOK, response.Error(response.CodeInvalidInput, "Invalid input: "+err.Error()))
		return
	}

	// Update benefit status
	err := h.benefitService.UpdateBenefitStatus(user.ID, benefitUUID, input.Status)
	if err != nil {
		code := response.CodeServerError
		if errors.Is(err, benefitpkg.ErrNotFound) {
			code = response.CodeBenefitNotFound
		} else if errors.Is(err, benefitpkg.ErrInvalidInput) {
			code = response.CodeInvalidInput
		}

		c.JSON(http.StatusOK, response.Error(code, "Failed to update benefit status: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

// GetBenefitClaims retrieves all claims for a specific benefit
func (h *BenefitHandler) GetBenefitClaims(c *gin.Context) {
	// Get user from context (set by auth middleware)
	userValue, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusOK, response.Error(response.CodeUnauthorized, "User not authenticated"))
		return
	}

	user := userValue.(*models.User)

	// Get benefit UUID from path
	benefitUUID := c.Param("uuid")
	if benefitUUID == "" {
		c.JSON(http.StatusOK, response.Error(response.CodeInvalidInput, "Benefit UUID is required"))
		return
	}

	// Get claims
	claims, err := h.benefitService.GetBenefitClaims(user.ID, benefitUUID)
	if err != nil {
		code := response.CodeServerError
		if errors.Is(err, benefitpkg.ErrNotFound) {
			code = response.CodeBenefitNotFound
		}

		c.JSON(http.StatusOK, response.Error(code, "Failed to retrieve benefit claims: "+err.Error()))
		return
	}

	// Format response
	responseData := make([]map[string]interface{}, len(claims))
	for i, claim := range claims {
		responseData[i] = map[string]interface{}{
			"id":             claim.ID,
			"claimed_at":     claim.ClaimedAt,
			"oauth_provider": claim.OAuthProvider,
			"user": map[string]interface{}{
				"id":       claim.User.ID,
				"username": claim.User.Username,
			},
			"code": claim.RedemptionCode.Code,
		}
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"claims": responseData,
	}))
}

// GetBenefitByUUID retrieves a benefit by its UUID
func (h *BenefitHandler) GetBenefitByUUID(c *gin.Context) {
	// Get benefit UUID from path
	benefitUUID := c.Param("uuid")
	if benefitUUID == "" {
		c.JSON(http.StatusOK, response.Error(response.CodeInvalidInput, "Benefit UUID is required"))
		return
	}

	// Get user from context (if authenticated)
	var userID uint
	userValue, exists := c.Get("user")
	if exists {
		user := userValue.(*models.User)
		userID = user.ID
	}

	// Get benefit
	benefit, err := h.benefitService.GetBenefitByUUID(benefitUUID)
	if err != nil {
		code := response.CodeServerError
		if errors.Is(err, benefitpkg.ErrNotFound) {
			code = response.CodeBenefitNotFound
		}

		c.JSON(http.StatusOK, response.Error(code, "Failed to retrieve benefit: "+err.Error()))
		return
	}

	// Check if benefit is active
	var claimStatus string = "available"
	if userID > 0 {
		// Check if user has already claimed
		var existingClaim models.Claim
		err := db.DB.Where("user_id = ? AND benefit_id = ?", userID, benefit.ID).First(&existingClaim).Error
		if err == nil {
			claimStatus = "claimed"
		}
	}

	if benefit.Status != "active" {
		claimStatus = "unavailable"
		errorMsg := "This benefit is no longer available"
		if benefit.Status == "paused" {
			errorMsg = "This benefit is temporarily paused"
			claimStatus = "paused"
		} else if benefit.Status == "expired" || benefit.ExpiresAt.Before(time.Now()) {
			errorMsg = "This benefit has expired"
			claimStatus = "expired"
		}

		if claimStatus == "unavailable" {
			c.JSON(http.StatusOK, response.Error(response.CodeBenefitNotActive, errorMsg))
			return
		}
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"benefit": map[string]interface{}{
			"uuid":          benefit.UUID,
			"title":         benefit.Title,
			"description":   benefit.Description,
			"total_count":   benefit.TotalCount,
			"claimed_count": benefit.ClaimedCount,
			"created_at":    benefit.CreatedAt,
			"expires_at":    benefit.ExpiresAt,
			"creator": map[string]interface{}{
				"username": benefit.Creator.Username,
				"id":       benefit.CreatorID,
			},
			"allowed_providers": benefit.AllowedProviders,
			"min_account_age":   benefit.MinAccountAge,
		},
		"claim_status": claimStatus,
	}))
}

// ClaimBenefit allows a user to claim a benefit
func (h *BenefitHandler) ClaimBenefit(c *gin.Context) {
	// Get user from context (set by auth middleware)
	userValue, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusOK, response.Error(response.CodeUnauthorized, "User not authenticated"))
		return
	}

	user := userValue.(*models.User)

	// Get benefit UUID from path
	benefitUUID := c.Param("uuid")
	if benefitUUID == "" {
		c.JSON(http.StatusOK, response.Error(response.CodeInvalidInput, "Benefit UUID is required"))
		return
	}

	// Get OAuth provider from the user's token
	claimsValue, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusOK, response.Error(response.CodeUnauthorized, "Invalid authentication"))
		return
	}

	// Default provider to "unknown" if not found in token
	provider := "unknown"
	if claims, ok := claimsValue.(map[string]interface{}); ok {
		if p, exists := claims["provider"]; exists && p != nil {
			provider = p.(string)
		}
	}

	// Get the benefit first to include in the response
	benefit, err := h.benefitService.GetBenefitByUUID(benefitUUID)
	if err != nil {
		code := response.CodeServerError
		if errors.Is(err, benefitpkg.ErrNotFound) {
			code = response.CodeBenefitNotFound
		}
		c.JSON(http.StatusOK, response.Error(code, "Failed to retrieve benefit: "+err.Error()))
		return
	}

	// Claim the benefit
	redemptionCode, err := h.benefitService.ClaimBenefit(
		user.ID,
		benefitUUID,
		provider,
		c.ClientIP(),
		c.Request.UserAgent(),
	)

	if err != nil {
		code := response.CodeServerError

		if errors.Is(err, benefitpkg.ErrNotFound) {
			code = response.CodeBenefitNotFound
		} else if errors.Is(err, benefitpkg.ErrBenefitExpired) {
			code = response.CodeBenefitExpired
		} else if errors.Is(err, benefitpkg.ErrBenefitPaused) {
			code = response.CodeBenefitNotActive
		} else if errors.Is(err, benefitpkg.ErrNoCodeAvailable) {
			code = response.CodeBenefitDepleted
		} else if errors.Is(err, benefitpkg.ErrAlreadyClaimed) {
			code = response.CodeBenefitAlreadyClaimed
		} else if errors.Is(err, benefitpkg.ErrProviderNotAllowed) || errors.Is(err, benefitpkg.ErrAccountTooNew) {
			code = response.CodeBenefitIneligible
		}

		c.JSON(http.StatusOK, response.Error(code, "Failed to claim benefit: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"claim": map[string]interface{}{
			"claimed_at":     time.Now(),
			"oauth_provider": provider,
			"benefit": map[string]interface{}{
				"id":          benefit.ID,
				"uuid":        benefit.UUID,
				"title":       benefit.Title,
				"description": benefit.Description,
			},
			"code": redemptionCode.Code,
		},
	}))
}
