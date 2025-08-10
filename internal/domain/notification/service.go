package notification

import "context"

type Service interface {
	Send(ctx context.Context, req *Request) error
	
	GetByID(ctx context.Context, id string) (*Notification, error)
	
	GetByOrderID(ctx context.Context, orderID string) ([]*Notification, error)
	
	Retry(ctx context.Context, id string) error
}

type Sender interface {
	Send(ctx context.Context, notification *Notification) error
	
	SupportedChannels() []Channel
}

type Template interface {
	Render(ctx context.Context, notificationType Type, data map[string]interface{}) (string, string, error)
	
	GetSubject(ctx context.Context, notificationType Type, data map[string]interface{}) (string, error)
	
	GetMessage(ctx context.Context, notificationType Type, data map[string]interface{}) (string, error)
}