package errors

// AppError 应用错误结构体
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	return e.Message
}

// New 创建新的应用错误
func New(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// NewWithError 创建带有原始错误的应用错误
func NewWithError(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// 常用错误码
const (
	// 通用错误码 1000-1999
	ErrCodeSuccess       = 1000
	ErrCodeUnknown       = 1001
	ErrCodeInvalidParams = 1002
	ErrCodeUnauthorized  = 1003
	ErrCodeForbidden     = 1004
	ErrCodeNotFound      = 1005
	ErrCodeConflict      = 1006

	// 用户相关错误码 2000-2999
	ErrCodeUserNotFound      = 2001
	ErrCodeUserAlreadyExists = 2002
	ErrCodeInvalidPassword   = 2003
	ErrCodeAccountBanned     = 2004
	ErrCodeInviteInvalid     = 2005

	// 资源相关错误码 3000-3999
	ErrCodeResourceNotFound  = 3001
	ErrCodeResourceForbidden = 3002
	ErrCodeResourceExpired   = 3003

	// 积分相关错误码 4000-4999
	ErrCodePointsInsufficient = 4001
	ErrCodePointsInvalid      = 4002

	// 权限相关错误码 5000-5999
	ErrCodePermissionDenied = 5001
	ErrCodeUploadNotAllowed = 5002
)
