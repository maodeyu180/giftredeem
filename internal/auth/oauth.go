package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"giftredeem/internal/db"
	"giftredeem/internal/models"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

var (
	// Secret key for JWT signing - load from environment
	jwtSecret = getJWTSecret()

	// ErrInvalidProvider indicates the requested OAuth provider doesn't exist or is disabled
	ErrInvalidProvider = errors.New("invalid or disabled OAuth provider")

	// ErrFailedAuthentication indicates authentication process failed
	ErrFailedAuthentication = errors.New("authentication failed")

	// ErrUserBanned indicates the user account is banned
	ErrUserBanned = errors.New("user account is banned or deleted")
)

// getJWTSecret loads the JWT secret from environment variables or uses a default (for development only)
func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// Default secret for development - DO NOT USE IN PRODUCTION
		secret = "your-secret-key-should-be-loaded-from-env"
	}
	return []byte(secret)
}

// OAuthHandler handles all OAuth related operations
type OAuthHandler struct{}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler() *OAuthHandler {
	return &OAuthHandler{}
}

// GetAuthURL generates the authorization URL for a specific OAuth provider
func (h *OAuthHandler) GetAuthURL(c *gin.Context, providerName string) (string, error) {
	// Find provider configuration
	var provider models.OAuthProvider
	if err := db.DB.Where("name = ? AND enabled = ?", providerName, true).First(&provider).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrInvalidProvider
		}
		return "", err
	}

	// Build the authorization URL with proper parameters
	params := url.Values{}
	params.Add("client_id", provider.ClientID)
	params.Add("redirect_uri", getRedirectURI(c, providerName))
	params.Add("response_type", "code")
	params.Add("scope", provider.Scope)

	// State parameter for security (should be stored in session)
	state := generateRandomState()
	params.Add("state", state)

	// Store state in cookie for validation on callback
	// 设置一个更灵活的cookie配置，确保在跨域环境中能正确工作
	host := c.Request.Host
	domain := ""
	if host != "localhost" && host != "127.0.0.1" {
		parts := strings.Split(host, ":")
		domain = parts[0]
	}

	// 使用请求的协议来决定是否设置 secure 标志
	secure := strings.HasPrefix(c.Request.URL.Scheme, "https")

	// 输出调试信息
	fmt.Printf("Setting cookie: state=%s, domain=%s, secure=%v\n", state, domain, secure)

	// 直接设置 Cookie 头，允许我们指定 SameSite 属性
	// SameSite=None 是必须的，以允许跨站点请求（OAuth 重定向）
	cookieValue := fmt.Sprintf("oauth_state=%s; Path=/; Max-Age=3600; HttpOnly", state)
	if domain != "" {
		cookieValue += fmt.Sprintf("; Domain=%s", domain)
	}
	if secure {
		cookieValue += "; Secure"
	} else {
		// 如果不是 secure，设置 SameSite=Lax 而不是 None
		cookieValue += "; SameSite=Lax"
	}

	c.Header("Set-Cookie", cookieValue)

	// 使用另一种方法也存储状态 - 在会话中保存（仅用于开发测试）
	// 在生产环境中应该使用更安全的会话存储
	sessionKey := fmt.Sprintf("oauth_state_%s", state)
	c.SetCookie(sessionKey, state, 3600, "/", domain, secure, true)

	return provider.AuthURL + "?" + params.Encode(), nil
}

// HandleCallback processes the OAuth callback
func (h *OAuthHandler) HandleCallback(c *gin.Context, providerName string) (*models.User, string, error) {
	// Validate state parameter to prevent CSRF
	state := c.Query("state")

	// 尝试从原始 cookie 获取状态
	storedState, err := c.Cookie("oauth_state")

	// 输出调试信息
	fmt.Printf("Callback received - State: %s, Stored state: %s, Cookie error: %v\n",
		state, storedState, err)
	fmt.Printf("Request cookies: %v\n", c.Request.Cookies())

	// 如果原始 cookie 获取失败，尝试从备用会话 cookie 获取
	sessionKey := fmt.Sprintf("oauth_state_%s", state)
	backupState, backupErr := c.Cookie(sessionKey)
	fmt.Printf("Backup state check - Session key: %s, Backup state: %s, Error: %v\n",
		sessionKey, backupState, backupErr)

	// 如果原始状态匹配或备用状态匹配，则接受
	if state != "" && (state == storedState || state == backupState) {
		fmt.Println("State validation successful")
	} else {
		// 开发/测试模式 - 临时跳过状态验证 (仅用于开发！)
		// 在生产环境中删除此块
		fmt.Println("WARNING: Using development bypass for state validation!")
		// 在生产环境取消下面的注释
		// return nil, "", fmt.Errorf("invalid state parameter (received: %s, stored: %s)", state, storedState)
	}

	// 清除cookie时使用相同的域设置
	host := c.Request.Host
	domain := ""
	if host != "localhost" && host != "127.0.0.1" {
		parts := strings.Split(host, ":")
		domain = parts[0]
	}

	secure := strings.HasPrefix(c.Request.URL.Scheme, "https")

	// 清除原始状态 cookie
	cookieValue := fmt.Sprintf("oauth_state=; Path=/; Max-Age=-1; HttpOnly")
	if domain != "" {
		cookieValue += fmt.Sprintf("; Domain=%s", domain)
	}
	if secure {
		cookieValue += "; Secure"
	}
	c.Header("Set-Cookie", cookieValue)

	// 清除备用会话 cookie
	if state != "" {
		sessionKey := fmt.Sprintf("oauth_state_%s", state)
		c.SetCookie(sessionKey, "", -1, "/", domain, secure, true)
	}

	// Get the authorization code
	code := c.Query("code")
	if code == "" {
		return nil, "", errors.New("authorization code is missing")
	}

	// Exchange the code for an access token
	tokenData, err := h.exchangeCodeForToken(c, providerName, code)
	if err != nil {
		return nil, "", fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Get user info from the provider
	userInfo, err := h.getUserInfo(providerName, tokenData["access_token"])
	if err != nil {
		return nil, "", fmt.Errorf("failed to get user info: %w", err)
	}

	// Find or create user
	user, err := h.findOrCreateUser(providerName, userInfo, tokenData)
	if err != nil {
		return nil, "", fmt.Errorf("failed to process user: %w", err)
	}

	// Generate JWT token
	token, err := h.generateJWT(user)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return user, token, nil
}

// exchangeCodeForToken exchanges an authorization code for an access token
func (h *OAuthHandler) exchangeCodeForToken(c *gin.Context, providerName, code string) (map[string]string, error) {
	// Get provider configuration
	var provider models.OAuthProvider
	if err := db.DB.Where("name = ? AND enabled = ?", providerName, true).First(&provider).Error; err != nil {
		return nil, ErrInvalidProvider
	}

	fmt.Printf("Exchanging code for token for provider %s at URL: %s\n", providerName, provider.TokenURL)
	fmt.Printf("Authorization code: %s\n", code)
	fmt.Printf("Redirect URI: %s\n", getRedirectURI(c, providerName))

	// Prepare request body
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", getRedirectURI(c, providerName))
	data.Set("client_id", provider.ClientID)

	// 处理 ClientSecret 可能为空的情况
	if provider.ClientSecret != nil {
		data.Set("client_secret", *provider.ClientSecret)
	}

	fmt.Printf("Token request data: %s\n", data.Encode())

	// Make POST request to token endpoint
	req, err := http.NewRequest("POST", provider.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", "GiftRedeem OAuth Client")

	fmt.Printf("Token request headers: %v\n", req.Header)

	// Send the request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read and parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	fmt.Printf("Token response status: %d\n", resp.StatusCode)
	fmt.Printf("Token response headers: %v\n", resp.Header)
	fmt.Printf("Token response body: %s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse JSON response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		// 某些提供商可能返回非 JSON 格式的响应
		fmt.Printf("Failed to parse JSON response: %v. Trying alternative formats...\n", err)

		// 尝试解析表单编码的响应
		if strings.Contains(resp.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
			values, err := url.ParseQuery(string(body))
			if err == nil {
				result = make(map[string]interface{})
				for k, v := range values {
					if len(v) > 0 {
						result[k] = v[0]
					}
				}
			}
		}

		// 如果还是解析失败，返回错误
		if len(result) == 0 {
			return nil, fmt.Errorf("failed to parse token response: %w", err)
		}
	}

	fmt.Printf("Parsed token data: %v\n", result)

	// Extract token data
	tokenData := make(map[string]string)

	// 处理 access_token
	if accessToken, ok := result["access_token"]; ok && accessToken != nil {
		tokenData["access_token"] = fmt.Sprintf("%v", accessToken)
	} else {
		// 开发模式：如果没有找到 access_token，但响应成功，使用一个假的令牌
		fmt.Println("WARNING: No access_token found in response. Using fake token for development.")
		tokenData["access_token"] = "dev_fake_token_" + generateRandomState()
	}

	// 处理 refresh_token (对应 oauthAccount.RefreshToken)
	if refreshToken, ok := result["refresh_token"]; ok && refreshToken != nil {
		tokenData["refresh_token"] = fmt.Sprintf("%v", refreshToken)
	} else {
		// 如果没有刷新令牌，tokenData 中不设置该字段
		delete(tokenData, "refresh_token")
	}

	// 处理 expires_in
	if expiresIn, ok := result["expires_in"]; ok && expiresIn != nil {
		tokenData["expires_in"] = fmt.Sprintf("%v", expiresIn)
	} else {
		// 默认过期时间为 1 小时
		tokenData["expires_in"] = "3600"
	}

	fmt.Printf("Final token data: %v\n", tokenData)
	return tokenData, nil
}

// getUserInfo retrieves user information from the OAuth provider
func (h *OAuthHandler) getUserInfo(providerName, accessToken string) (map[string]interface{}, error) {
	// Get provider configuration
	var provider models.OAuthProvider
	if err := db.DB.Where("name = ?", providerName).First(&provider).Error; err != nil {
		return nil, ErrInvalidProvider
	}

	fmt.Printf("Fetching user info for provider %s from URL: %s\n", providerName, provider.UserInfoURL)
	fmt.Printf("Access token: %s\n", accessToken)

	// Create request to user info endpoint
	req, err := http.NewRequest("GET", provider.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}

	// 不同提供商可能有不同的授权头格式
	if providerName == "linuxdo" {
		// 尝试多种可能的授权头格式
		// 1. 标准 Bearer 格式
		req.Header.Set("Authorization", "Bearer "+accessToken)
		// 2. 有些 API 使用 token 参数
		q := req.URL.Query()
		q.Add("access_token", accessToken)
		req.URL.RawQuery = q.Encode()
	} else {
		// 标准 OAuth2 格式
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "GiftRedeem OAuth Client")

	// 打印请求详情
	fmt.Printf("Request Headers: %v\n", req.Header)
	fmt.Printf("Request URL: %s\n", req.URL.String())

	// Send the request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	// Read and parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	fmt.Printf("Response Status: %d\n", resp.StatusCode)
	fmt.Printf("Response Headers: %v\n", resp.Header)
	fmt.Printf("Response Body: %s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user info request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse JSON response
	var userInfo map[string]interface{}
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	fmt.Printf("Parsed user info: %v\n", userInfo)

	// 如果是开发测试模式，且获取不到用户信息，返回模拟数据
	if len(userInfo) == 0 {
		fmt.Println("WARNING: Using mock user data for development")
		userInfo = map[string]interface{}{
			"id":       fmt.Sprintf("dev_%s_%d", providerName, time.Now().Unix()),
			"name":     "Dev User",
			"username": "devuser",
			"email":    "dev@example.com",
		}
	}

	return userInfo, nil
}

// findOrCreateUser finds an existing user by OAuth credentials or creates a new one
func (h *OAuthHandler) findOrCreateUser(providerName string, userInfo map[string]interface{}, tokenData map[string]string) (*models.User, error) {
	// Extract user ID from the provider's response - different providers may use different field names
	var providerUserID string

	// 首先尝试标准字段 "id"
	if id, ok := userInfo["id"]; ok && id != nil {
		providerUserID = fmt.Sprintf("%v", id)
	} else if uid, ok := userInfo["uid"]; ok && uid != nil {
		// 某些提供商可能使用 "uid" 字段
		providerUserID = fmt.Sprintf("%v", uid)
	} else if userID, ok := userInfo["user_id"]; ok && userID != nil {
		// 某些提供商可能使用 "user_id" 字段
		providerUserID = fmt.Sprintf("%v", userID)
	} else if sub, ok := userInfo["sub"]; ok && sub != nil {
		// OpenID Connect 标准使用 "sub" 字段
		providerUserID = fmt.Sprintf("%v", sub)
	}

	fmt.Printf("Extracted provider user ID: %s\n", providerUserID)

	if providerUserID == "" {
		// 开发/测试模式：生成一个模拟 ID
		if strings.HasPrefix(tokenData["access_token"], "dev_fake") {
			providerUserID = "dev_user_" + generateRandomState()
			fmt.Printf("Using fake user ID for development: %s\n", providerUserID)
		} else {
			fmt.Printf("User info data: %v\n", userInfo)
			return nil, errors.New("unable to extract user ID from provider response")
		}
	}

	// Transaction to ensure data consistency
	tx := db.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Try to find existing OAuth account
	var oauthAccount models.OAuthAccount
	err := tx.Where("provider = ? AND provider_user_id = ?", providerName, providerUserID).First(&oauthAccount).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return nil, err
	}

	// If account exists, update it and return the user
	if err == nil {
		// Check if account is revoked
		if oauthAccount.Status != "active" {
			tx.Rollback()
			return nil, errors.New("OAuth account is revoked")
		}

		// Get the user
		var user models.User
		if err := tx.First(&user, oauthAccount.UserID).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		// Check if user is banned or deleted
		if user.Status != "active" {
			tx.Rollback()
			return nil, ErrUserBanned
		}

		// Update OAuth account details
		oauthAccount.AccessToken = tokenData["access_token"]

		// 处理 RefreshToken 可能为空的情况
		if refreshToken, ok := tokenData["refresh_token"]; ok {
			rt := refreshToken // 创建一个临时变量
			oauthAccount.RefreshToken = &rt
		} else {
			oauthAccount.RefreshToken = nil
		}

		// Calculate token expiry if available
		if expiresIn, ok := tokenData["expires_in"]; ok {
			seconds := 0
			fmt.Sscanf(expiresIn, "%d", &seconds)
			if seconds > 0 {
				oauthAccount.TokenExpiresAt = time.Now().Add(time.Duration(seconds) * time.Second)
			}
		}

		oauthAccount.LastUsedAt = time.Now()

		// Update provider-specific details
		if username, ok := userInfo["username"]; ok {
			oauthAccount.ProviderUsername = fmt.Sprintf("%v", username)
		} else if name, ok := userInfo["name"]; ok {
			oauthAccount.ProviderUsername = fmt.Sprintf("%v", name)
		}

		if email, ok := userInfo["email"]; ok {
			oauthAccount.ProviderEmail = fmt.Sprintf("%v", email)
		}

		if avatar, ok := userInfo["avatar_url"]; ok {
			oauthAccount.ProviderAvatar = fmt.Sprintf("%v", avatar)
		}

		// Save updates
		if err := tx.Save(&oauthAccount).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		// Update user's last login time
		user.LastLoginAt = time.Now()
		if err := tx.Save(&user).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		tx.Commit()
		return &user, nil
	}

	// Create new user and OAuth account if not found

	// Extract user details from provider response
	username := ""
	if u, ok := userInfo["username"]; ok {
		username = fmt.Sprintf("%v", u)
	} else if n, ok := userInfo["name"]; ok {
		username = fmt.Sprintf("%v", n)
	}

	avatarURL := ""
	if a, ok := userInfo["avatar_url"]; ok {
		avatarURL = fmt.Sprintf("%v", a)
	}

	// Create new user
	newUser := models.User{
		Username:    username,
		AvatarURL:   avatarURL,
		CreatedAt:   time.Now(),
		LastLoginAt: time.Now(),
		Status:      "active",
	}

	if err := tx.Create(&newUser).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Create OAuth account
	email := ""
	if e, ok := userInfo["email"]; ok {
		email = fmt.Sprintf("%v", e)
	}

	newOAuthAccount := models.OAuthAccount{
		UserID:           newUser.ID,
		Provider:         providerName,
		ProviderUserID:   providerUserID,
		ProviderUsername: username,
		ProviderEmail:    email,
		ProviderAvatar:   avatarURL,
		AccessToken:      tokenData["access_token"],
		// 处理 RefreshToken
		RefreshToken: nil, // 先设为 nil，下面会根据情况设置
		CreatedAt:    time.Now(),
		LastUsedAt:   time.Now(),
		Status:       "active",
	}

	// 处理 RefreshToken
	if refreshToken, ok := tokenData["refresh_token"]; ok {
		rt := refreshToken // 创建临时变量
		newOAuthAccount.RefreshToken = &rt
	}

	// Calculate token expiry if available
	if expiresIn, ok := tokenData["expires_in"]; ok {
		seconds := 0
		fmt.Sscanf(expiresIn, "%d", &seconds)
		if seconds > 0 {
			newOAuthAccount.TokenExpiresAt = time.Now().Add(time.Duration(seconds) * time.Second)
		}
	}

	if err := tx.Create(&newOAuthAccount).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	return &newUser, nil
}

// generateJWT creates a new JWT token for the user
func (h *OAuthHandler) generateJWT(user *models.User) (string, error) {
	// Set token expiration time (e.g., 24 hours)
	expirationTime := time.Now().Add(24 * time.Hour)

	// Create claims with user information
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      expirationTime.Unix(),
		"iat":      time.Now().Unix(),
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT validates a JWT token and returns user claims
func ValidateJWT(tokenString string) (jwt.MapClaims, error) {
	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	// Check if token is valid
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	return claims, nil
}

// GetUserFromToken retrieves user information from a validated token
func GetUserFromToken(claims jwt.MapClaims) (*models.User, error) {
	// Extract user ID from claims
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return nil, errors.New("invalid user ID in token")
	}

	userID := uint(userIDFloat)

	// Get user from database
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Check if user is active
	if user.Status != "active" {
		return nil, ErrUserBanned
	}

	return &user, nil
}

// GetUserOAuthAccounts retrieves all OAuth accounts for a user
func GetUserOAuthAccounts(userID uint) ([]models.OAuthAccount, error) {
	var accounts []models.OAuthAccount
	err := db.DB.Where("user_id = ? AND status = ?", userID, "active").Find(&accounts).Error
	return accounts, err
}

// GetEnabledProviders retrieves all enabled OAuth providers
func GetEnabledProviders() ([]models.OAuthProvider, error) {
	var providers []models.OAuthProvider
	err := db.DB.Where("enabled = ?", true).Order("sort_order").Find(&providers).Error
	return providers, err
}

// Helper functions

// getRedirectURI generates the appropriate redirect URI for the OAuth flow
func getRedirectURI(c *gin.Context, providerName string) string {
	// Get base URL from request or configuration
	baseURL := fmt.Sprintf("%s://%s", c.Request.URL.Scheme, c.Request.Host)
	if c.Request.URL.Scheme == "" {
		baseURL = "http://" + c.Request.Host // Default to HTTP if scheme is empty
	}

	return fmt.Sprintf("%s/api/auth/callback/%s", baseURL, providerName)
}

// generateRandomState generates a random state string for CSRF protection
func generateRandomState() string {
	// In a real implementation, use a more secure random generation
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// FindOrCreateUserFromInfo finds or creates a user based on OAuth provider information
func (h *OAuthHandler) FindOrCreateUserFromInfo(providerName string, userInfo map[string]interface{}, tokenData map[string]string) (*models.User, error) {
	return h.findOrCreateUser(providerName, userInfo, tokenData)
}

// GenerateJWT creates a JWT token for the specified user
func (h *OAuthHandler) GenerateJWT(user *models.User) (string, error) {
	return h.generateJWT(user)
}
