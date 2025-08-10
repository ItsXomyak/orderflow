// internal/domain/notification/service.go
package notification

import "context"

// Service интерфейс для работы с уведомлениями
type Service interface {
	// Send отправляет уведомление
	Send(ctx context.Context, req *Request) error
	
	// GetByID получает уведомление по ID
	GetByID(ctx context.Context, id string) (*Notification, error)
	
	// GetByOrderID получает уведомления по ID заказа
	GetByOrderID(ctx context.Context, orderID string) ([]*Notification, error)
	
	// Retry повторяет отправку неудачного уведомления
	Retry(ctx context.Context, id string) error
}

// Sender интерфейс для отправки уведомлений через различные каналы
type Sender interface {
	// Send отправляет уведомление
	Send(ctx context.Context, notification *Notification) error
	
	// SupportedChannels возвращает поддерживаемые каналы
	SupportedChannels() []Channel
}

// Template интерфейс для работы с шаблонами уведомлений
type Template interface {
	// Render рендерит шаблон уведомления
	Render(ctx context.Context, notificationType Type, data map[string]interface{}) (string, string, error)
	
	// GetSubject возвращает тему уведомления
	GetSubject(ctx context.Context, notificationType Type, data map[string]interface{}) (string, error)
	
	// GetMessage возвращает сообщение уведомления
	GetMessage(ctx context.Context, notificationType Type, data map[string]interface{}) (string, error)
}