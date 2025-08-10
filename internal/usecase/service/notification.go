package service

import (
	"context"
	"errors"
	"sort"
	"sync"

	notif "orderflow/internal/domain/notification"

	"github.com/google/uuid"
)

// In-memory реализация notification.Service поверх доменного слоя
// Заменяется на БД/репозиторий без изменения интерфейсов.

type notificationService struct {
	// Реализация отправки по каналам
	senders map[notif.Channel]notif.Sender
	// Опциональный шаблонизатор
	tmpl notif.Template

	// Хранилище — для демо in-memory
	mu            sync.RWMutex
	notifications map[string]*notif.Notification // id -> notification
	byOrder       map[string][]string            // orderID -> []ids
}

// NewNotificationService конструирует сервис.
// Передайте набор Sender'ов (email/sms/push) и опционально Template.
func NewNotificationService(senders []notif.Sender, tmpl notif.Template) notif.Service {
	idx := make(map[notif.Channel]notif.Sender)
	for _, s := range senders {
		for _, ch := range s.SupportedChannels() {
			idx[ch] = s
		}
	}
	return &notificationService{
		senders:       idx,
		tmpl:          tmpl,
		notifications: make(map[string]*notif.Notification),
		byOrder:       make(map[string][]string),
	}
}

// Send — создать и отправить уведомление. Если есть Template — отрендерить/дополнить Subject/Message.
func (s *notificationService) Send(ctx context.Context, req *notif.Request) error {
	if req == nil {
		return errors.New("nil request")
	}
	// Построить сущность
	n := notif.NewNotification(req)
	n.ID = uuid.NewString()

	// Если есть шаблоны и пустые поля — подставим
	if s.tmpl != nil {
		if n.Subject == "" || n.Message == "" {
			// Для простоты: пробуем Render целиком, без data.
			subj, msg, err := s.tmpl.Render(ctx, n.Type, map[string]interface{}{
				"order_id":    n.OrderID,
				"customer_id": n.CustomerID,
				"type":        n.Type,
			})
			if err != nil {
				return err
			}
			if n.Subject == "" {
				n.Subject = subj
			}
			if n.Message == "" {
				n.Message = msg
			}
		}
	}

	if err := n.Validate(); err != nil {
		return err
	}

	// Найти sender по каналу
	sender, ok := s.senders[n.Channel]
	if !ok {
		return errors.New("no sender for channel: " + string(n.Channel))
	}

	// Отправка
	if err := sender.Send(ctx, n); err != nil {
		s.mu.Lock()
		n.MarkAsFailed()
		s.saveLocked(n)
		s.mu.Unlock()
		return err
	}

	s.mu.Lock()
	n.MarkAsSent()
	s.saveLocked(n)
	s.mu.Unlock()
	return nil
}

func (s *notificationService) GetByID(ctx context.Context, id string) (*notif.Notification, error) {
	if id == "" {
		return nil, errors.New("empty id")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	n, ok := s.notifications[id]
	if !ok {
		return nil, errors.New("notification not found")
	}
	return n, nil
}

func (s *notificationService) GetByOrderID(ctx context.Context, orderID string) ([]*notif.Notification, error) {
	if orderID == "" {
		return nil, errors.New("empty orderID")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := s.byOrder[orderID]
	res := make([]*notif.Notification, 0, len(ids))
	for _, id := range ids {
		if n, ok := s.notifications[id]; ok {
			res = append(res, n)
		}
	}
	// сортируем по времени
	sort.Slice(res, func(i, j int) bool { return res[i].CreatedAt.Before(res[j].CreatedAt) })
	return res, nil
}

func (s *notificationService) Retry(ctx context.Context, id string) error {
	n, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !n.IsFailed() {
		return errors.New("notification is not failed")
	}

	sender, ok := s.senders[n.Channel]
	if !ok {
		return errors.New("no sender for channel: " + string(n.Channel))
	}

	if err := sender.Send(ctx, n); err != nil {
		s.mu.Lock()
		n.MarkAsFailed()
		s.saveLocked(n)
		s.mu.Unlock()
		return err
	}

	s.mu.Lock()
	n.MarkAsSent()
	s.saveLocked(n)
	s.mu.Unlock()
	return nil
}

// saveLocked — сохранить запись и индексацию. Должен вызываться под s.mu.Lock().
func (s *notificationService) saveLocked(n *notif.Notification) {
	if n.ID == "" {
		n.ID = uuid.NewString()
	}
	s.notifications[n.ID] = n
	// Индексация по заказу
	if n.OrderID != "" {
		if _, ok := s.byOrder[n.OrderID]; !ok {
			s.byOrder[n.OrderID] = make([]string, 0, 1)
		}
		// Если id уже есть — не дублируем
		exists := false
		for _, id := range s.byOrder[n.OrderID] {
			if id == n.ID {
				exists = true
				break
			}
		}
		if !exists {
			s.byOrder[n.OrderID] = append(s.byOrder[n.OrderID], n.ID)
		}
	}
}
