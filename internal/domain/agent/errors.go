package agent

import "fmt"

// ErrorCode represents a standardized error code for agent failures.
type ErrorCode string

const (
	// Transient errors - can be retried
	ErrCodeLLMTimeout         ErrorCode = "llm_timeout"
	ErrCodeRateLimit          ErrorCode = "rate_limit"
	ErrCodeServiceUnavailable ErrorCode = "service_unavailable"
	ErrCodeNetworkError       ErrorCode = "network_error"

	// Permanent errors - should not be retried
	ErrCodeInvalidInput     ErrorCode = "invalid_input"
	ErrCodeInvalidConfig    ErrorCode = "invalid_config"
	ErrCodeUnsupportedType  ErrorCode = "unsupported_type"
	ErrCodeResourceNotFound ErrorCode = "resource_not_found"

	// Configuration errors
	ErrCodeMissingAPIKey ErrorCode = "missing_api_key"
	ErrCodeInvalidModel  ErrorCode = "invalid_model"

	// Execution errors
	ErrCodeExecutionFailed ErrorCode = "execution_failed"
	ErrCodeStageTimeout    ErrorCode = "stage_timeout"
	ErrCodeCancelled       ErrorCode = "cancelled"

	// Internal errors
	ErrCodeInternalError ErrorCode = "internal_error"
	ErrCodeUnknown       ErrorCode = "unknown"
)

// ErrorCategory classifies errors for handling strategy.
type ErrorCategory string

const (
	ErrorCategoryTransient     ErrorCategory = "transient"     // Can be retried
	ErrorCategoryPermanent     ErrorCategory = "permanent"     // Should not retry
	ErrorCategoryConfiguration ErrorCategory = "configuration" // User action needed
	ErrorCategoryInternal      ErrorCategory = "internal"      // System error
)

// ErrorCodeInfo maps error codes to their categories and suggested actions.
var errorCodeInfo = map[ErrorCode]struct {
	Category   ErrorCategory
	Suggestion string
}{
	ErrCodeLLMTimeout: {
		Category:   ErrorCategoryTransient,
		Suggestion: "The request timed out. Please try again.",
	},
	ErrCodeRateLimit: {
		Category:   ErrorCategoryTransient,
		Suggestion: "Rate limit exceeded. Please wait a moment and try again.",
	},
	ErrCodeServiceUnavailable: {
		Category:   ErrorCategoryTransient,
		Suggestion: "The service is temporarily unavailable. Please try again later.",
	},
	ErrCodeNetworkError: {
		Category:   ErrorCategoryTransient,
		Suggestion: "A network error occurred. Please check your connection and try again.",
	},
	ErrCodeInvalidInput: {
		Category:   ErrorCategoryPermanent,
		Suggestion: "The input provided is invalid. Please check and correct your input.",
	},
	ErrCodeInvalidConfig: {
		Category:   ErrorCategoryConfiguration,
		Suggestion: "Configuration is invalid. Please check your settings.",
	},
	ErrCodeUnsupportedType: {
		Category:   ErrorCategoryPermanent,
		Suggestion: "This type of request is not supported.",
	},
	ErrCodeResourceNotFound: {
		Category:   ErrorCategoryPermanent,
		Suggestion: "The requested resource was not found.",
	},
	ErrCodeMissingAPIKey: {
		Category:   ErrorCategoryConfiguration,
		Suggestion: "API key is missing. Please configure your API key in settings.",
	},
	ErrCodeInvalidModel: {
		Category:   ErrorCategoryConfiguration,
		Suggestion: "The specified model is invalid or not available.",
	},
	ErrCodeExecutionFailed: {
		Category:   ErrorCategoryInternal,
		Suggestion: "Execution failed. Please try again or contact support if the issue persists.",
	},
	ErrCodeStageTimeout: {
		Category:   ErrorCategoryTransient,
		Suggestion: "The stage took too long to complete. Please try again.",
	},
	ErrCodeCancelled: {
		Category:   ErrorCategoryPermanent,
		Suggestion: "The operation was cancelled.",
	},
	ErrCodeInternalError: {
		Category:   ErrorCategoryInternal,
		Suggestion: "An internal error occurred. Please try again or contact support.",
	},
	ErrCodeUnknown: {
		Category:   ErrorCategoryInternal,
		Suggestion: "An unexpected error occurred. Please try again.",
	},
}

// NewErrorDetail creates a standardized ErrorDetail with category and suggestion.
func NewErrorDetail(code ErrorCode, message string, retryable bool) *ErrorDetail {
	info, ok := errorCodeInfo[code]
	if !ok {
		info = errorCodeInfo[ErrCodeUnknown]
	}

	// Override retryable based on category if not explicitly set
	if !retryable && info.Category == ErrorCategoryTransient {
		retryable = true
	}

	return &ErrorDetail{
		Message:    message,
		Code:       string(code),
		Retryable:  retryable,
		Stage:      "",
		Category:   string(info.Category),
		Suggestion: info.Suggestion,
	}
}

// WrapAgentError wraps an error with stage context and error code.
func WrapAgentError(err error, stage StageName, code ErrorCode) *ErrorDetail {
	if err == nil {
		return nil
	}

	detail := NewErrorDetail(code, err.Error(), false)
	detail.Stage = stage
	return detail
}

// ErrorDetailWithStage creates an ErrorDetail with stage context.
func ErrorDetailWithStage(detail *ErrorDetail, stage StageName) *ErrorDetail {
	if detail == nil {
		return nil
	}
	cloned := *detail
	cloned.Stage = stage
	return &cloned
}

// ClassifyError attempts to classify an error into an appropriate ErrorCode.
func ClassifyError(err error) ErrorCode {
	if err == nil {
		return ErrCodeUnknown
	}

	errStr := err.Error()

	// Check for common error patterns
	switch {
	case containsAny(errStr, []string{"timeout", "deadline exceeded", "context deadline"}):
		return ErrCodeLLMTimeout
	case containsAny(errStr, []string{"rate limit", "too many requests", "429"}):
		return ErrCodeRateLimit
	case containsAny(errStr, []string{"unavailable", "502", "503", "504", "bad gateway", "service unavailable", "retryable status"}):
		return ErrCodeServiceUnavailable
	case containsAny(errStr, []string{"network", "connection refused", "no such host"}):
		return ErrCodeNetworkError
	case containsAny(errStr, []string{"invalid input", "bad request", "invalid parameter"}):
		return ErrCodeInvalidInput
	case containsAny(errStr, []string{"api key", "unauthorized", "401", "forbidden", "403"}):
		return ErrCodeMissingAPIKey
	case containsAny(errStr, []string{"not found", "404"}):
		return ErrCodeResourceNotFound
	case containsAny(errStr, []string{"cancelled", "canceled", "context canceled"}):
		return ErrCodeCancelled
	default:
		return ErrCodeUnknown
	}
}

// containsAny checks if s contains any of the substrings.
func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if contains(s, substr) {
			return true
		}
	}
	return false
}

// contains checks if s contains substr (case-insensitive).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && containsIgnoreCase(s, substr)))
}

func containsIgnoreCase(s, substr string) bool {
	// Simple case-insensitive contains
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			sc := s[i+j]
			subc := substr[j]
			if sc >= 'A' && sc <= 'Z' {
				sc += 32
			}
			if subc >= 'A' && subc <= 'Z' {
				subc += 32
			}
			if sc != subc {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// Error implements the error interface for ErrorDetail.
func (e *ErrorDetail) Error() string {
	if e == nil {
		return ""
	}
	if e.Stage != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Stage, e.Message)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// IsRetryable returns true if the error is retryable.
func (e *ErrorDetail) IsRetryable() bool {
	if e == nil {
		return false
	}
	return e.Retryable
}

// IsTransient returns true if the error category is transient.
func (e *ErrorDetail) IsTransient() bool {
	if e == nil {
		return false
	}
	return e.Category == string(ErrorCategoryTransient)
}

// IsConfiguration returns true if the error requires configuration action.
func (e *ErrorDetail) IsConfiguration() bool {
	if e == nil {
		return false
	}
	return e.Category == string(ErrorCategoryConfiguration)
}
