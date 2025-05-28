package response

// StandardResponse represents the standard API response format
type StandardResponse struct {
	Code int         `json:"code"` // 0 means success, other values indicate specific errors
	Msg  string      `json:"msg"`  // Success message or error description
	Data interface{} `json:"data"` // Response data payload, can be null
}

// Common error codes
const (
	// Success code
	CodeSuccess = 0

	// General error codes (1-999)
	CodeUnknownError = 1   // Unknown/general error
	CodeInvalidInput = 400 // Invalid input/request
	CodeUnauthorized = 401 // Unauthorized access
	CodeForbidden    = 403 // Forbidden access
	CodeNotFound     = 404 // Resource not found
	CodeServerError  = 500 // Internal server error

	// Authentication error codes (1000-1999)
	CodeAuthProviderNotFound = 1000 // OAuth provider not found or disabled
	CodeAuthCodeMissing      = 1001 // OAuth authorization code missing
	CodeAuthFailed           = 1002 // Authentication failed
	CodeAuthUserBanned       = 1003 // User account is banned or deleted
	CodeAuthInvalidToken     = 1004 // Invalid token
	CodeAuthExpiredToken     = 1005 // Expired token

	// Benefit error codes (2000-2999)
	CodeBenefitCreationFailed = 2000 // Failed to create benefit
	CodeBenefitNotFound       = 2001 // Benefit not found
	CodeBenefitExpired        = 2002 // Benefit expired
	CodeBenefitDepleted       = 2003 // No more codes available
	CodeBenefitNotActive      = 2004 // Benefit not active
	CodeBenefitAlreadyClaimed = 2005 // User already claimed this benefit
	CodeBenefitIneligible     = 2006 // User ineligible for this benefit
)

// Success creates a success response with data
func Success(data interface{}) StandardResponse {
	return StandardResponse{
		Code: CodeSuccess,
		Msg:  "success",
		Data: data,
	}
}

// Error creates an error response with code and message
func Error(code int, message string) StandardResponse {
	return StandardResponse{
		Code: code,
		Msg:  message,
		Data: nil,
	}
}
