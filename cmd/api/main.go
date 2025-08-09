package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.temporal.io/sdk/client"
	"go.uber.org/zap"

	"order-processing/internal/application/workflow"
	"order-processing/internal/domain/order"
)

type OrderResponse struct {
	OrderID       string `json:"order_id"`
	WorkflowID    string `json:"workflow_id"`
	RunID         string `json:"run_id"`
	Status        string `json:"status"`
	Message       string `json:"message"`
}

type OrderStatusResponse struct {
	OrderID       string           `json:"order_id"`
	Status        order.OrderStatus `json:"status"`
	PaymentID     string           `json:"payment_id,omitempty"`
	TransactionID string           `json:"transaction_id,omitempty"`
	Error         string           `json:"error,omitempty"`
	WorkflowID    string           `json:"workflow_id"`
	RunID         string           `json:"run_id"`
}

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("Failed to create logger", err)
	}
	defer logger.Sync()

	temporalClient, err := client.NewClient(client.Options{
		HostPort: "localhost:7233",
	})
	if err != nil {
		logger.Fatal("Failed to create Temporal client", zap.Error(err))
	}
	defer temporalClient.Close()

	r := gin.Default()

	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Order Processing System",
		})
	})

	r.POST("/api/orders", func(c *gin.Context) {
		var request order.OrderRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		orderID := fmt.Sprintf("ORDER-%d", time.Now().Unix())

		totalAmount := 0.0
		for i := range request.Items {
			request.Items[i].TotalPrice = float64(request.Items[i].Quantity) * request.Items[i].UnitPrice
			totalAmount += request.Items[i].TotalPrice
		}

		workflowInput := workflow.OrderWorkflowInput{
			OrderID:     orderID,
			CustomerID:  request.CustomerID,
			Items:       request.Items,
			TotalAmount: totalAmount,
		}

		workflowOptions := client.StartWorkflowOptions{
			ID:        fmt.Sprintf("%s-%s", workflow.WorkflowIDPrefix, orderID),
			TaskQueue: workflow.TaskQueueOrderProcessing,
		}

		workflowRun, err := temporalClient.ExecuteWorkflow(context.Background(), workflowOptions, workflow.OrderWorkflow, workflowInput)
		if err != nil {
			logger.Error("Failed to start workflow", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		logger.Info("Workflow started", 
			zap.String("workflowID", workflowRun.GetID()),
			zap.String("runID", workflowRun.GetRunID()),
			zap.String("orderID", orderID),
		)

		c.JSON(http.StatusOK, OrderResponse{
			OrderID:    orderID,
			WorkflowID: workflowRun.GetID(),
			RunID:      workflowRun.GetRunID(),
			Status:     "started",
			Message:    "Order processing started",
		})
	})

	r.GET("/api/orders/:orderID/status", func(c *gin.Context) {
		orderID := c.Param("orderID")
		workflowID := fmt.Sprintf("%s-%s", workflow.WorkflowIDPrefix, orderID)

		workflowRun := temporalClient.GetWorkflow(context.Background(), workflowID, "")
		
		var result workflow.OrderWorkflowOutput
		err := workflowRun.Get(context.Background(), &result)
		if err != nil {
			logger.Error("Failed to get workflow result", zap.Error(err))
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}

		c.JSON(http.StatusOK, OrderStatusResponse{
			OrderID:       result.OrderID,
			Status:        result.Status,
			PaymentID:     result.PaymentID,
			TransactionID: result.TransactionID,
			Error:         result.Error,
			WorkflowID:    workflowID,
			RunID:         workflowRun.GetRunID(),
		})
	})

	r.POST("/api/orders/:orderID/cancel", func(c *gin.Context) {
		orderID := c.Param("orderID")
		workflowID := fmt.Sprintf("%s-%s", workflow.WorkflowIDPrefix, orderID)

		err := temporalClient.SignalWorkflow(context.Background(), workflowID, "", "cancel-order", nil)
		if err != nil {
			logger.Error("Failed to signal workflow cancellation", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		logger.Info("Cancellation signal sent", 
			zap.String("workflowID", workflowID),
			zap.String("orderID", orderID),
		)

		c.JSON(http.StatusOK, gin.H{
			"message": "Cancellation signal sent",
			"orderID": orderID,
		})
	})

	logger.Info("Starting API server on :8080")
	if err := r.Run(":8080"); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}