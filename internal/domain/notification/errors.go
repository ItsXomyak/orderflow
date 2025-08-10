// internal/domain/notification/errors.go
package notification

import "fmt"

// ValidationError ошибка валидации уведомления
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("notification validation error: %s", e.Message)
}

func NewValidationError(message string) *ValidationError {
	return &ValidationError{Message: message}
}

// NotFoundError ошибка, когда уведомление не найдено
type NotFoundError struct {
	NotificationID string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("notification not found: %s", e.NotificationID)
}

func NewNotFoundError(notificationID string) *NotFoundError {
	return &NotFoundError{NotificationID: notificationID}
}

// SendError ошибка отправки уведомления
type SendError struct {
	Channel Channel
	Reason  string
}

func (e *SendError) Error() string {
	return fmt.Sprintf("failed to send notification via %s: %s", e.Channel, e.Reason)
}

func NewSendError(channel Channel, reason string) *SendError {
	return &SendError{Channel: channel, Reason: reason}
}

// UnsupportedChannelError ошибка неподдерживаемого канала
type UnsupportedChannelError struct {
	Channel Channel
}

func (e *UnsupportedChannelError) Error() string {
	return fmt.Sprintf("unsupported notification channel: %s", e.Channel)
}

func NewUnsupportedChannelError(channel Channel) *UnsupportedChannelError {
	return &UnsupportedChannelError{Channel: channel}
}

// TemplateError ошибка работы с шаблоном
type TemplateError struct {
	Type    Type
	Message string
}

func (e *TemplateError) Error() string {
	return fmt.Sprintf("template error for %s: %s", e.Type, e.Message)
}

func NewTemplateError(notificationType Type, message string) *TemplateError {
	return &TemplateError{Type: notificationType, Message: message}
}