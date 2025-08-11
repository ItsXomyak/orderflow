package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"go.temporal.io/sdk/client"

	"orderflow/internal/domain/order"
	"orderflow/internal/domain/workflow"
	"orderflow/pkg/logger"
)

type OrderHandler struct {
	temporalClient client.Client
}

func NewOrderHandler(temporalClient client.Client) *OrderHandler {
	return &OrderHandler{
		temporalClient: temporalClient,
	}
}

type CreateOrderRequest struct {
	CustomerID string      `json:"customer_id"`
	Items      []order.Item `json:"items"`
}

type CreateOrderResponse struct {
	WorkflowID string `json:"workflow_id"`
	Message    string `json:"message"`
}

type OrderStatusResponse struct {
	WorkflowID string      `json:"workflow_id"`
	Status     order.Status `json:"status"`
	Message    string      `json:"message"`
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.CustomerID == "" {
		http.Error(w, "customer_id is required", http.StatusBadRequest)
		return
	}
	if len(req.Items) == 0 {
		http.Error(w, "items are required", http.StatusBadRequest)
		return
	}

	input := &workflow.OrderProcessingInput{
		CustomerID: req.CustomerID,
		Items:      req.Items,
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        "order-processing-" + req.CustomerID + "-" + strconv.FormatInt(time.Now().Unix(), 10),
		TaskQueue: workflow.OrderProcessingTaskQueue,
	}

	workflowRun, err := h.temporalClient.ExecuteWorkflow(r.Context(), workflowOptions, workflow.OrderProcessingWorkflow, input)
	if err != nil {
		logger.Error("Failed to start workflow", "error", err)
		http.Error(w, "Failed to start order processing", http.StatusInternalServerError)
		return
	}

	response := CreateOrderResponse{
		WorkflowID: workflowRun.GetID(),
		Message:    "Order processing started successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *OrderHandler) GetOrderStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	workflowID := r.URL.Query().Get("workflow_id")
	if workflowID == "" {
		http.Error(w, "workflow_id is required", http.StatusBadRequest)
		return
	}

	var status order.Status
	_, err := h.temporalClient.QueryWorkflow(r.Context(), workflowID, "", workflow.OrderStatusQuery, &status)
	if err != nil {
		logger.Error("Failed to query workflow status", "error", err, "workflow_id", workflowID)
		http.Error(w, "Failed to get order status", http.StatusInternalServerError)
		return
	}

	response := OrderStatusResponse{
		WorkflowID: workflowID,
		Status:     status,
		Message:    "Order status retrieved successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	workflowID := r.URL.Query().Get("workflow_id")
	if workflowID == "" {
		http.Error(w, "workflow_id is required", http.StatusBadRequest)
		return
	}

	err := h.temporalClient.SignalWorkflow(r.Context(), workflowID, "", workflow.CancelOrderSignal, "cancel")
	if err != nil {
		logger.Error("Failed to send cancel signal", "error", err, "workflow_id", workflowID)
		http.Error(w, "Failed to cancel order", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"workflow_id": workflowID,
		"message":     "Order cancellation signal sent successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *OrderHandler) GetWorkflowState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	workflowID := r.URL.Query().Get("workflow_id")
	if workflowID == "" {
		http.Error(w, "workflow_id is required", http.StatusBadRequest)
		return
	}

	var state workflow.State
	_, err := h.temporalClient.QueryWorkflow(r.Context(), workflowID, "", workflow.WorkflowStateQuery, &state)
	if err != nil {
		logger.Error("Failed to query workflow state", "error", err, "workflow_id", workflowID)
		http.Error(w, "Failed to get workflow state", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state)
}
