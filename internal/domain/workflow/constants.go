package workflow

import "time"

const (
	OrderProcessingWorkflow = "OrderProcessingWorkflow"
	
	CreateOrderActivity      = "CreateOrderActivity"
	CheckInventoryActivity   = "CheckInventoryActivity"
	ProcessPaymentActivity   = "ProcessPaymentActivity"
	SendNotificationActivity = "SendNotificationActivity"
	CancelOrderActivity      = "CancelOrderActivity"
	
	OrderProcessingTaskQueue = "order-processing"
)

const (
	CancelOrderSignal = "cancel-order"
)

const (
	OrderStatusQuery   = "order-status"
	WorkflowStateQuery = "workflow-state"
)

const (
	DefaultMaximumAttempts    = 3
	DefaultInitialInterval    = 1 * time.Second
	DefaultMaximumInterval    = 60 * time.Second
	DefaultBackoffCoefficient = 2.0
)

const (
	OrderCreationDuration     = 1 * time.Second
	InventoryCheckDuration    = 2 * time.Second
	PaymentProcessDuration    = 3 * time.Second
	NotificationDuration      = 1 * time.Second
)

const (
	StepCreateOrder      = "create_order"
	StepCheckInventory   = "check_inventory"
	StepProcessPayment   = "process_payment"
	StepSendNotification = "send_notification"
	StepComplete         = "complete"
	StepFailed           = "failed"
	StepCancelled        = "cancelled"
)

const (
	ErrorCodeValidation           = "VALIDATION_ERROR"
	ErrorCodeInventoryUnavailable = "INVENTORY_UNAVAILABLE"
	ErrorCodePaymentFailed        = "PAYMENT_FAILED"
	ErrorCodeNotificationFailed   = "NOTIFICATION_FAILED"
	ErrorCodeOrderCancelled       = "ORDER_CANCELLED"
	ErrorCodeOrderNotFound        = "ORDER_NOT_FOUND"
	ErrorCodeInternalError        = "INTERNAL_ERROR"
)