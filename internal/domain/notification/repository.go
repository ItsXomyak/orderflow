package notification

import (
	"context"
)

type Repository interface {
	CreateNotification(ctx context.Context, notification *Notification) error
	GetNotification(ctx context.Context, id string) (*Notification, error)
	GetNotificationsByOrderID(ctx context.Context, orderID string) ([]*Notification, error)
	UpdateNotification(ctx context.Context, notification *Notification) error
	GetNotifications(ctx context.Context) ([]*Notification, error)
	GetFailedNotifications(ctx context.Context) ([]*Notification, error)
}
