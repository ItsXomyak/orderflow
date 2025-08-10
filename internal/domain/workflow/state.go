package workflow

import (
	"time"

	"orderflow/internal/domain/order"
)

type State struct {
	OrderID       string       `json:"order_id"`
	CustomerID    string       `json:"customer_id"`
	CurrentStep   string       `json:"current_step"`
	Status        order.Status `json:"status"`
	ErrorMessage  string       `json:"error_message,omitempty"`
	ErrorCode     string       `json:"error_code,omitempty"`
	RetryCount    int          `json:"retry_count"`
	IsCancelled   bool         `json:"is_cancelled"`
	PaymentID     string       `json:"payment_id,omitempty"`
	StartedAt     time.Time    `json:"started_at"`
	CompletedAt   *time.Time   `json:"completed_at,omitempty"`
	
	StepHistory []StepExecution `json:"step_history,omitempty"`
}

type StepExecution struct {
	Step        string    `json:"step"`
	Status      string    `json:"status"` // "started", "completed", "failed"
	StartedAt   time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Error       string    `json:"error,omitempty"`
	RetryCount  int       `json:"retry_count"`
}

func NewState(orderID, customerID string) *State {
	return &State{
		OrderID:     orderID,
		CustomerID:  customerID,
		CurrentStep: StepCreateOrder,
		Status:      order.StatusPending,
		StartedAt:   time.Now(),
		StepHistory: make([]StepExecution, 0),
	}
}

func (s *State) IsCompleted() bool {
	return s.Status == order.StatusCompleted
}

func (s *State) IsFailed() bool {
	return s.Status == order.StatusFailed
}

func (s *State) UpdateStep(step string) {
	s.CurrentStep = step
	s.startStepExecution(step)
}

func (s *State) UpdateStatus(status order.Status) {
	s.Status = status
	if status == order.StatusCompleted || status == order.StatusFailed || status == order.StatusCancelled {
		now := time.Now()
		s.CompletedAt = &now
		s.completeCurrentStep(status == order.StatusCompleted, "")
	}
}

func (s *State) SetError(code, message string) {
	s.ErrorCode = code
	s.ErrorMessage = message
	s.Status = order.StatusFailed
	now := time.Now()
	s.CompletedAt = &now
	s.completeCurrentStep(false, message)
}

func (s *State) Cancel() {
	s.IsCancelled = true
	s.Status = order.StatusCancelled
	now := time.Now()
	s.CompletedAt = &now
	s.completeCurrentStep(false, "Workflow cancelled")
}

func (s *State) IncrementRetry() {
	s.RetryCount++
	if len(s.StepHistory) > 0 {
		s.StepHistory[len(s.StepHistory)-1].RetryCount++
	}
}

func (s *State) GetCurrentStepExecution() *StepExecution {
	if len(s.StepHistory) == 0 {
		return nil
	}
	return &s.StepHistory[len(s.StepHistory)-1]
}

func (s *State) startStepExecution(step string) {
	execution := StepExecution{
		Step:      step,
		Status:    "started",
		StartedAt: time.Now(),
	}
	s.StepHistory = append(s.StepHistory, execution)
}

func (s *State) completeCurrentStep(success bool, errorMsg string) {
	if len(s.StepHistory) == 0 {
		return
	}
	
	currentStep := &s.StepHistory[len(s.StepHistory)-1]
	now := time.Now()
	currentStep.CompletedAt = &now
	
	if success {
		currentStep.Status = "completed"
	} else {
		currentStep.Status = "failed"
		currentStep.Error = errorMsg
	}
}

func (s *State) GetDuration() time.Duration {
	if s.CompletedAt != nil {
		return s.CompletedAt.Sub(s.StartedAt)
	}
	return time.Since(s.StartedAt)
}

func (s *State) GetStepDuration(step string) *time.Duration {
	for _, execution := range s.StepHistory {
		if execution.Step == step && execution.CompletedAt != nil {
			duration := execution.CompletedAt.Sub(execution.StartedAt)
			return &duration
		}
	}
	return nil
}