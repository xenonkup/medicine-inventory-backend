package domain

// AppError is a domain-level error carrying a stable machine code, an HTTP
// status, and a human-readable message. Handlers translate it into the
// standard error envelope.
type AppError struct {
	Code     string
	HTTPCode int
	Message  string
}

func (e *AppError) Error() string { return e.Message }

// NewAppError builds an AppError.
func NewAppError(code string, httpCode int, message string) *AppError {
	return &AppError{Code: code, HTTPCode: httpCode, Message: message}
}

// Common domain errors shared across features. Business-specific errors
// (INSUFFICIENT_STOCK, RETURN_AFTER_EXPIRY, ...) are added in later phases.
var (
	ErrInvalidCredentials = NewAppError("INVALID_CREDENTIALS", 401, "username or password is incorrect")
	ErrUnauthorized       = NewAppError("UNAUTHORIZED", 401, "authentication required")
	ErrForbidden          = NewAppError("FORBIDDEN", 403, "you do not have permission to perform this action")
	ErrUserInactive       = NewAppError("USER_INACTIVE", 403, "this account has been deactivated")
	ErrUserNotFound       = NewAppError("USER_NOT_FOUND", 404, "user not found")
	ErrDuplicateUsername  = NewAppError("DUPLICATE_USERNAME", 409, "username already exists")
	ErrInvalidToken       = NewAppError("INVALID_TOKEN", 401, "token is invalid or expired")

	// Category
	ErrCategoryNotFound  = NewAppError("CATEGORY_NOT_FOUND", 404, "category not found")
	ErrDuplicateCategory = NewAppError("DUPLICATE_CATEGORY", 409, "category name already exists")
	ErrCategoryInUse     = NewAppError("CATEGORY_IN_USE", 409, "category still has active medicines")

	// Medicine
	ErrMedicineNotFound      = NewAppError("MEDICINE_NOT_FOUND", 404, "medicine not found")
	ErrDuplicateMedicineCode = NewAppError("DUPLICATE_MEDICINE_CODE", 409, "medicine code already exists")
)
