package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"orderflow/internal/domain/notification"
)

type NotificationPG struct {
	pool *pgxpool.Pool
}

func NewNotificationPG(pool *pgxpool.Pool) *NotificationPG {
	return &NotificationPG{pool: pool}
}

func (r *NotificationPG) CreateNotification(ctx context.Context, notificationEntity *notification.Notification) error {
	const q = `
		INSERT INTO notifications (id, customer_id, order_id, type, channel, status, subject, message, metadata, sent_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.pool.Exec(ctx, q,
		notificationEntity.ID, notificationEntity.CustomerID, notificationEntity.OrderID,
		string(notificationEntity.Type), string(notificationEntity.Channel), string(notificationEntity.Status),
		notificationEntity.Subject, notificationEntity.Message, notificationEntity.Metadata,
		notificationEntity.SentAt, notificationEntity.CreatedAt, notificationEntity.UpdatedAt,
	)
	return err
}

func (r *NotificationPG) GetNotification(ctx context.Context, id string) (*notification.Notification, error) {
	const q = `
		SELECT id, customer_id, order_id, type, channel, status, subject, message, metadata, sent_at, created_at, updated_at
		FROM notifications WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, q, id)

	var notificationEntity notification.Notification
	var notificationType, channel, status string
	err := row.Scan(
		&notificationEntity.ID, &notificationEntity.CustomerID, &notificationEntity.OrderID,
		&notificationType, &channel, &status, &notificationEntity.Subject, &notificationEntity.Message,
		&notificationEntity.Metadata, &notificationEntity.SentAt, &notificationEntity.CreatedAt, &notificationEntity.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, notification.NewNotFoundError(id)
	}
	if err != nil {
		return nil, err
	}

	notificationEntity.Type = notification.Type(notificationType)
	notificationEntity.Channel = notification.Channel(channel)
	notificationEntity.Status = notification.Status(status)
	return &notificationEntity, nil
}

func (r *NotificationPG) GetNotificationsByOrderID(ctx context.Context, orderID string) ([]*notification.Notification, error) {
	const q = `
		SELECT id, customer_id, order_id, type, channel, status, subject, message, metadata, sent_at, created_at, updated_at
		FROM notifications WHERE order_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, q, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*notification.Notification
	for rows.Next() {
		var notificationEntity notification.Notification
		var notificationType, channel, status string
		err := rows.Scan(
			&notificationEntity.ID, &notificationEntity.CustomerID, &notificationEntity.OrderID,
			&notificationType, &channel, &status, &notificationEntity.Subject, &notificationEntity.Message,
			&notificationEntity.Metadata, &notificationEntity.SentAt, &notificationEntity.CreatedAt, &notificationEntity.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		notificationEntity.Type = notification.Type(notificationType)
		notificationEntity.Channel = notification.Channel(channel)
		notificationEntity.Status = notification.Status(status)
		notifications = append(notifications, &notificationEntity)
	}

	return notifications, rows.Err()
}

func (r *NotificationPG) UpdateNotification(ctx context.Context, notificationEntity *notification.Notification) error {
	const q = `
		UPDATE notifications
		SET customer_id = $2, order_id = $3, type = $4, channel = $5, status = $6,
		    subject = $7, message = $8, metadata = $9, sent_at = $10, updated_at = $11
		WHERE id = $1
	`
	ct, err := r.pool.Exec(ctx, q,
		notificationEntity.ID, notificationEntity.CustomerID, notificationEntity.OrderID,
		string(notificationEntity.Type), string(notificationEntity.Channel), string(notificationEntity.Status),
		notificationEntity.Subject, notificationEntity.Message, notificationEntity.Metadata,
		notificationEntity.SentAt, notificationEntity.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return notification.NewNotFoundError(notificationEntity.ID)
	}
	return nil
}

func (r *NotificationPG) GetNotifications(ctx context.Context) ([]*notification.Notification, error) {
	const q = `
		SELECT id, customer_id, order_id, type, channel, status, subject, message, metadata, sent_at, created_at, updated_at
		FROM notifications ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*notification.Notification
	for rows.Next() {
		var notificationEntity notification.Notification
		var notificationType, channel, status string
		err := rows.Scan(
			&notificationEntity.ID, &notificationEntity.CustomerID, &notificationEntity.OrderID,
			&notificationType, &channel, &status, &notificationEntity.Subject, &notificationEntity.Message,
			&notificationEntity.Metadata, &notificationEntity.SentAt, &notificationEntity.CreatedAt, &notificationEntity.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		notificationEntity.Type = notification.Type(notificationType)
		notificationEntity.Channel = notification.Channel(channel)
		notificationEntity.Status = notification.Status(status)
		notifications = append(notifications, &notificationEntity)
	}

	return notifications, rows.Err()
}

func (r *NotificationPG) GetFailedNotifications(ctx context.Context) ([]*notification.Notification, error) {
	const q = `
		SELECT id, customer_id, order_id, type, channel, status, subject, message, metadata, sent_at, created_at, updated_at
		FROM notifications WHERE status = 'failed' ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*notification.Notification
	for rows.Next() {
		var notificationEntity notification.Notification
		var notificationType, channel, status string
		err := rows.Scan(
			&notificationEntity.ID, &notificationEntity.CustomerID, &notificationEntity.OrderID,
			&notificationType, &channel, &status, &notificationEntity.Subject, &notificationEntity.Message,
			&notificationEntity.Metadata, &notificationEntity.SentAt, &notificationEntity.CreatedAt, &notificationEntity.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		notificationEntity.Type = notification.Type(notificationType)
		notificationEntity.Channel = notification.Channel(channel)
		notificationEntity.Status = notification.Status(status)
		notifications = append(notifications, &notificationEntity)
	}

	return notifications, rows.Err()
}
