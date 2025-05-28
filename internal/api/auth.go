package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"giftredeem/internal/auth"
	"giftredeem/internal/db"
	"giftredeem/internal/models"
	"giftredeem/internal/response"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthHandler handles authentication related requests
type AuthHandler struct {
	oauthHandler *auth.OAuthHandler
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		oauthHandler: auth.NewOAuthHandler(),
	}
}

// GetProviders returns all enabled OAuth providers
func (h *AuthHandler) GetProviders(c *gin.Context) {
	providers, err := auth.GetEnabledProviders()
	if err != nil {
		c.JSON(http.StatusOK, response.Error(response.CodeServerError, "Failed to get providers: "+err.Error()))
		return
	}

	// Map to simple response format
	responseData := make([]map[string]interface{}, len(providers))
	for i, p := range providers {
		responseData[i] = map[string]interface{}{
			"name":         p.Name,
			"display_name": p.DisplayName,
		}
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"providers": responseData,
	}))
}

// Login initiates the OAuth flow for a provider
func (h *AuthHandler) Login(c *gin.Context) {
	providerName := c.Param("provider")
	if providerName == "" {
		c.JSON(http.StatusOK, response.Error(response.CodeInvalidInput, "Provider name is required"))
		return
	}

	// Generate auth URL
	authURL, err := h.oauthHandler.GetAuthURL(c, providerName)
	if err != nil {
		code := response.CodeServerError
		if err == auth.ErrInvalidProvider {
			code = response.CodeAuthProviderNotFound
		}

		c.JSON(http.StatusOK, response.Error(code, err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"auth_url": authURL,
	}))
}

// Callback handles the OAuth callback
func (h *AuthHandler) Callback(c *gin.Context) {
	providerName := c.Param("provider")
	if providerName == "" {
		c.JSON(http.StatusOK, response.Error(response.CodeInvalidInput, "Provider name is required"))
		return
	}

	// 获取前端应用根URL，默认为本地开发环境
	frontendBaseURL := c.Query("redirect_uri")
	if frontendBaseURL == "" {
		// 根据环境变量或配置获取前端URL
		frontendBaseURL = os.Getenv("FRONTEND_URL")
		if frontendBaseURL == "" {
			frontendBaseURL = "http://localhost:3000" // 前端服务器端口
		}
	}

	// 前端回调路径
	callbackPath := "/auth/callback/" + providerName
	errorPath := "/login" // 登录失败默认路径

	// 处理OAuth回调
	user, token, err := h.oauthHandler.HandleCallback(c, providerName)

	// 处理API调用模式 - 返回JSON而不是重定向
	if c.GetHeader("Accept") == "application/json" || c.Query("response_type") == "json" {
		if err != nil {
			code := response.CodeAuthFailed
			if err == auth.ErrInvalidProvider {
				code = response.CodeAuthProviderNotFound
			} else if err == auth.ErrUserBanned {
				code = response.CodeAuthUserBanned
			}
			c.JSON(http.StatusOK, response.Error(code, "Authentication failed: "+err.Error()))
		} else {
			c.JSON(http.StatusOK, response.Success(map[string]interface{}{
				"token": token,
				"user": map[string]interface{}{
					"id":         user.ID,
					"username":   user.Username,
					"avatar_url": user.AvatarURL,
				},
			}))
		}
		return
	}

	// 浏览器模式 - 总是重定向到前端
	var redirectURL string

	if err != nil {
		// 登录失败，重定向到错误页面
		code := response.CodeAuthFailed
		if err == auth.ErrInvalidProvider {
			code = response.CodeAuthProviderNotFound
		} else if err == auth.ErrUserBanned {
			code = response.CodeAuthUserBanned
		}
		errorMsg := url.QueryEscape(fmt.Sprintf("%d:%s", code, err.Error()))
		redirectURL = fmt.Sprintf("%s%s?error=%s", frontendBaseURL, errorPath, errorMsg)
	} else {
		// 登录成功，重定向到成功页面，带上token和用户信息
		redirectURL = fmt.Sprintf("%s%s?token=%s&user_id=%d&username=%s&success=true",
			frontendBaseURL,
			callbackPath,
			url.QueryEscape(token),
			user.ID,
			url.QueryEscape(user.Username))

		// 如果用户有头像，也添加到URL
		if user.AvatarURL != "" {
			redirectURL += "&avatar_url=" + url.QueryEscape(user.AvatarURL)
		}
	}

	// 调试信息
	fmt.Printf("Redirecting to: %s\n", redirectURL)

	// 执行重定向
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// GetUserProfile retrieves the profile of the current logged-in user
func (h *AuthHandler) GetUserProfile(c *gin.Context) {
	// Get user from context (set by auth middleware)
	userValue, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusOK, response.Error(response.CodeUnauthorized, "User not authenticated"))
		return
	}

	user := userValue.(*models.User)

	// Get user's OAuth accounts
	accounts, err := auth.GetUserOAuthAccounts(user.ID)
	if err != nil {
		c.JSON(http.StatusOK, response.Error(response.CodeServerError, "Failed to get user accounts: "+err.Error()))
		return
	}

	// Format OAuth accounts for response
	accountsResponse := make([]map[string]interface{}, len(accounts))
	for i, acc := range accounts {
		accountsResponse[i] = map[string]interface{}{
			"provider":          acc.Provider,
			"provider_username": acc.ProviderUsername,
			"provider_email":    acc.ProviderEmail,
			"provider_avatar":   acc.ProviderAvatar,
			"created_at":        acc.CreatedAt,
			"last_used_at":      acc.LastUsedAt,
		}
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"user": map[string]interface{}{
			"id":         user.ID,
			"username":   user.Username,
			"avatar_url": user.AvatarURL,
			"created_at": user.CreatedAt,
			"accounts":   accountsResponse,
		},
	}))
}

// VerifyCode verifies an OAuth authorization code and exchanges it for a token
func (h *AuthHandler) VerifyCode(c *gin.Context) {
	providerName := c.Param("provider")
	if providerName == "" {
		c.JSON(http.StatusOK, response.Error(response.CodeInvalidInput, "Provider name is required"))
		return
	}

	// 从请求体中获取授权码
	var request struct {
		Code string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusOK, response.Error(response.CodeInvalidInput, "Invalid request: "+err.Error()))
		return
	}

	// 获取提供商配置
	var provider models.OAuthProvider
	if err := db.DB.Where("name = ? AND enabled = ?", providerName, true).First(&provider).Error; err != nil {
		code := response.CodeServerError
		if errors.Is(err, gorm.ErrRecordNotFound) {
			code = response.CodeAuthProviderNotFound
		}
		c.JSON(http.StatusOK, response.Error(code, "Provider not found or disabled"))
		return
	}

	// 构建回调URL（用于交换令牌）
	redirectURI := fmt.Sprintf("%s://%s/auth/callback/%s",
		c.Request.URL.Scheme,
		c.Request.Host,
		providerName)

	if c.Request.URL.Scheme == "" {
		redirectURI = fmt.Sprintf("http://%s/auth/callback/%s",
			c.Request.Host,
			providerName)
	}

	// 准备请求数据
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", request.Code)
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", provider.ClientID)

	// 处理 ClientSecret 可能为空的情况
	if provider.ClientSecret != nil {
		data.Set("client_secret", *provider.ClientSecret)
	}

	// 请求访问令牌
	req, err := http.NewRequest("POST", provider.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		c.JSON(http.StatusOK, response.Error(response.CodeServerError, "Failed to create token request: "+err.Error()))
		return
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")

	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusOK, response.Error(response.CodeServerError, "Failed to exchange code for token: "+err.Error()))
		return
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusOK, response.Error(response.CodeServerError, "Failed to read token response: "+err.Error()))
		return
	}

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusOK, response.Error(response.CodeAuthFailed, fmt.Sprintf("Token request failed with status %d: %s", resp.StatusCode, string(body))))
		return
	}

	// 解析JSON响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		c.JSON(http.StatusOK, response.Error(response.CodeServerError, "Failed to parse token response: "+err.Error()))
		return
	}

	// 提取访问令牌
	accessToken, ok := result["access_token"].(string)
	if !ok || accessToken == "" {
		c.JSON(http.StatusOK, response.Error(response.CodeAuthFailed, "Invalid token response: missing access_token"))
		return
	}

	// 获取用户信息
	userInfoReq, err := http.NewRequest("GET", provider.UserInfoURL, nil)
	if err != nil {
		c.JSON(http.StatusOK, response.Error(response.CodeServerError, "Failed to create user info request: "+err.Error()))
		return
	}

	userInfoReq.Header.Add("Authorization", "Bearer "+accessToken)
	userInfoReq.Header.Add("Accept", "application/json")

	userInfoResp, err := client.Do(userInfoReq)
	if err != nil {
		c.JSON(http.StatusOK, response.Error(response.CodeServerError, "Failed to get user info: "+err.Error()))
		return
	}
	defer userInfoResp.Body.Close()

	userInfoBody, err := io.ReadAll(userInfoResp.Body)
	if err != nil {
		c.JSON(http.StatusOK, response.Error(response.CodeServerError, "Failed to read user info response: "+err.Error()))
		return
	}

	if userInfoResp.StatusCode != http.StatusOK {
		c.JSON(http.StatusOK, response.Error(response.CodeAuthFailed, fmt.Sprintf("User info request failed with status %d: %s", userInfoResp.StatusCode, string(userInfoBody))))
		return
	}

	// 解析用户信息
	var userInfo map[string]interface{}
	if err := json.Unmarshal(userInfoBody, &userInfo); err != nil {
		c.JSON(http.StatusOK, response.Error(response.CodeServerError, "Failed to parse user info: "+err.Error()))
		return
	}

	// 处理用户信息，创建或更新用户
	// 这部分可以复用 OAuthHandler 中的代码
	tokenData := map[string]string{
		"access_token": accessToken,
	}
	if refreshToken, ok := result["refresh_token"].(string); ok {
		tokenData["refresh_token"] = refreshToken
	}
	if expiresIn, ok := result["expires_in"].(string); ok {
		tokenData["expires_in"] = expiresIn
	}

	// 查找或创建用户
	user, err := h.oauthHandler.FindOrCreateUserFromInfo(providerName, userInfo, tokenData)
	if err != nil {
		code := response.CodeServerError
		if err == auth.ErrUserBanned {
			code = response.CodeAuthUserBanned
		}
		c.JSON(http.StatusOK, response.Error(code, "Failed to process user: "+err.Error()))
		return
	}

	// 生成JWT令牌
	token, err := h.oauthHandler.GenerateJWT(user)
	if err != nil {
		c.JSON(http.StatusOK, response.Error(response.CodeServerError, "Failed to generate token: "+err.Error()))
		return
	}

	// 返回JWT令牌和用户信息
	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":         user.ID,
			"username":   user.Username,
			"avatar_url": user.AvatarURL,
		},
	}))
}
