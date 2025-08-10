package workflow

import (
	"time"

	"go.temporal.io/sdk/workflow"

	workflowDomain "orderflow/internal/domain/workflow"
)
type SignalHandlers struct{}

func NewSignalHandlers() *SignalHandlers {
	return &SignalHandlers{}
}

func (s *SignalHandlers) SetupCancelSignal(ctx workflow.Context, state *workflowDomain.State) workflow.ReceiveChannel {
	logger := workflow.GetLogger(ctx)
	
	cancelChannel := workflow.GetSignalChannel(ctx, workflowDomain.CancelOrderSignal)
	
	workflow.Go(ctx, func(ctx workflow.Context) {
		for {
			var cancelRequest CancelSignalRequest
			if !cancelChannel.ReceiveAsync(&cancelRequest) {
				workflow.Sleep(ctx, 100*time.Millisecond)
				continue
			}
			
			logger.Info("Received cancel signal", 
				"order_id", state.OrderID,
				"reason", cancelRequest.Reason,
				"requested_by", cancelRequest.RequestedBy)
			
			if s.canCancelAtCurrentStep(state.CurrentStep) {
				state.Cancel()
				logger.Info("Order cancellation initiated", "order_id", state.OrderID)
			} else {
				logger.Warn("Cannot cancel order at current step", 
					"current_step", state.CurrentStep,
					"order_id", state.OrderID)
			}
		}
	})
	
	return cancelChannel
}

type CancelSignalRequest struct {
	Reason      string `json:"reason"`
	RequestedBy string `json:"requested_by"`
}

func (s *SignalHandlers) canCancelAtCurrentStep(currentStep string) bool {
	switch currentStep {
	case workflowDomain.StepCreateOrder:
		return true
	case workflowDomain.StepCheckInventory:
		return true
	case workflowDomain.StepProcessPayment:
		return true
	case workflowDomain.StepSendNotification:
		return false
	case workflowDomain.StepComplete:
		return false
	case workflowDomain.StepFailed:
		return false
	case workflowDomain.StepCancelled:
		return false
	default:
		return false
	}
}

func (s *SignalHandlers) SetupWorkflowQueries(ctx workflow.Context, state *workflowDomain.State) error {
	err := workflow.SetQueryHandler(ctx, workflowDomain.OrderStatusQuery, func() (string, error) {
		return string(state.Status), nil
	})
	if err != nil {
		return err
	}
	
	err = workflow.SetQueryHandler(ctx, workflowDomain.WorkflowStateQuery, func() (*workflowDomain.State, error) {
		return state, nil
	})
	if err != nil {
		return err
	}
	
	err = workflow.SetQueryHandler(ctx, "step-history", func() ([]workflowDomain.StepExecution, error) {
		return state.StepHistory, nil
	})
	if err != nil {
		return err
	}
	
	err = workflow.SetQueryHandler(ctx, "progress", func() (map[string]interface{}, error) {
		totalSteps := 5 // create, check, payment, notification, complete
		currentStepIndex := s.getStepIndex(state.CurrentStep)
		
		progress := map[string]interface{}{
			"current_step":       state.CurrentStep,
			"total_steps":        totalSteps,
			"current_step_index": currentStepIndex,
			"progress_percent":   float64(currentStepIndex) / float64(totalSteps) * 100,
			"is_completed":       state.IsCompleted(),
			"is_failed":          state.IsFailed(),
			"is_cancelled":       state.IsCancelled,
			"duration_seconds":   state.GetDuration().Seconds(),
		}
		
		return progress, nil
	})
	
	return err
}

func (s *SignalHandlers) getStepIndex(step string) int {
	switch step {
	case workflowDomain.StepCreateOrder:
		return 1
	case workflowDomain.StepCheckInventory:
		return 2
	case workflowDomain.StepProcessPayment:
		return 3
	case workflowDomain.StepSendNotification:
		return 4
	case workflowDomain.StepComplete:
		return 5
	default:
		return 0
	}
}