package workflow

import "fmt"

// ошибка валидации входных данных workflow
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("workflow validation error: %s", e.Message)
}

func NewValidationError(message string) *ValidationError {
	return &ValidationError{Message: message}
}

// ошибка выполнения Activity
type ActivityError struct {
	ActivityName string
	Step         string
	Code         string
	Message      string
	Retryable    bool
}

func (e *ActivityError) Error() string {
	return fmt.Sprintf("activity %s failed at step %s [%s]: %s", 
		e.ActivityName, e.Step, e.Code, e.Message)
}

func (e *ActivityError) IsRetryable() bool {
	return e.Retryable
}

func NewActivityError(activityName, step, code, message string, retryable bool) *ActivityError {
	return &ActivityError{
		ActivityName: activityName,
		Step:         step,
		Code:         code,
		Message:      message,
		Retryable:    retryable,
	}
}

// ошибка отмены workflow
type CancellationError struct {
	OrderID string
	Reason  string
}

func (e *CancellationError) Error() string {
	return fmt.Sprintf("workflow cancelled for order %s: %s", e.OrderID, e.Reason)
}

func NewCancellationError(orderID, reason string) *CancellationError {
	return &CancellationError{OrderID: orderID, Reason: reason}
}

// ошибка таймаута workflow
type TimeoutError struct {
	Step    string
	Timeout string
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("workflow timeout at step %s after %s", e.Step, e.Timeout)
}

func NewTimeoutError(step, timeout string) *TimeoutError {
	return &TimeoutError{Step: step, Timeout: timeout}
}

// ошибка исчерпания попыток повтора
type RetryExhaustedError struct {
	ActivityName string
	MaxAttempts  int
	LastError    string
}

func (e *RetryExhaustedError) Error() string {
	return fmt.Sprintf("retry exhausted for activity %s after %d attempts, last error: %s", 
		e.ActivityName, e.MaxAttempts, e.LastError)
}

func NewRetryExhaustedError(activityName string, maxAttempts int, lastError string) *RetryExhaustedError {
	return &RetryExhaustedError{
		ActivityName: activityName,
		MaxAttempts:  maxAttempts,
		LastError:    lastError,
	}
}