package workflow

import "time"

// Temporal Workflow и Activity имена
const (
	// Workflow
	OrderProcessingWorkflow = "OrderProcessingWorkflow"
	
	// Activities
	CreateOrderActivity      = "CreateOrderActivity"
	CheckInventoryActivity   = "CheckInventoryActivity"
	ProcessPaymentActivity   = "ProcessPaymentActivity"
	SendNotificationActivity = "SendNotificationActivity"
	CancelOrderActivity      = "CancelOrderActivity"
	
	// Task Queue
	OrderProcessingTaskQueue = "order-processing"
)

// Temporal Signals
const (
	CancelOrderSignal = "cancel-order"
)

// Temporal Queries
const (
	OrderStatusQuery   = "order-status"
	WorkflowStateQuery = "workflow-state"
)

// Temporal настройки retry policy
const (
	DefaultMaximumAttempts    = 3
	DefaultInitialInterval    = 1 * time.Second
	DefaultMaximumInterval    = 60 * time.Second
	DefaultBackoffCoefficient = 2.0
)

// Временные интервалы для симуляции
const (
	OrderCreationDuration     = 1 * time.Second
	InventoryCheckDuration    = 2 * time.Second
	PaymentProcessDuration    = 3 * time.Second
	NotificationDuration      = 1 * time.Second
)

// WorkflowSteps представляет этапы workflow
const (
	StepCreateOrder      = "create_order"
	StepCheckInventory   = "check_inventory"
	StepProcessPayment   = "process_payment"
	StepSendNotification = "send_notification"
	StepComplete         = "complete"
	StepFailed           = "failed"
	StepCancelled        = "cancelled"
)

// Коды ошибок workflow
const (
	ErrorCodeValidation           = "VALIDATION_ERROR"
	ErrorCodeInventoryUnavailable = "INVENTORY_UNAVAILABLE"
	ErrorCodePaymentFailed        = "PAYMENT_FAILED"
	ErrorCodeNotificationFailed   = "NOTIFICATION_FAILED"
	ErrorCodeOrderCancelled       = "ORDER_CANCELLED"
	ErrorCodeOrderNotFound        = "ORDER_NOT_FOUND"
	ErrorCodeInternalError        = "INTERNAL_ERROR"
)