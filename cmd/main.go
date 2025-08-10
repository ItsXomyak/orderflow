package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"orderflow/internal/adapter/repository"
	"orderflow/internal/domain/workflow"
	"orderflow/internal/httpserver"
	"orderflow/internal/usecase/activity"
	"orderflow/internal/usecase/service"
	"orderflow/pkg/logger"
)

func main() {
	logger.Init("development")

	logger.Info("Starting OrderFlow application...")

	pool, err := pgxpool.New(context.Background(), "postgres://postgres:password@localhost:5432/orderflow?sslmode=disable")
	if err != nil {
		logger.Error("Failed to connect to PostgreSQL", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	orderRepo := repository.NewOrderPG(pool)
	inventoryRepo := repository.NewInventoryPG(pool)
	paymentRepo := repository.NewPaymentPG(pool)
	notificationRepo := repository.NewNotificationPG(pool)

	orderService := service.NewOrderService(orderRepo)
	inventoryService := service.NewInventoryService(inventoryRepo)
	paymentService := service.NewPaymentService(paymentRepo)
	notificationService := service.NewNotificationService(notificationRepo)

	createOrderActivity := activity.NewCreateOrderActivity(orderService)
	checkInventoryActivity := activity.NewCheckInventoryActivity(inventoryService, orderService)
	processPaymentActivity := activity.NewProcessPaymentActivity(paymentService, orderService, inventoryService)
	sendNotificationActivity := activity.NewSendNotificationActivity(notificationService, orderService)
	cancelOrderActivity := activity.NewCancelOrderActivity(orderService, paymentService, inventoryService)

	temporalClient, err := client.NewClient(client.Options{
		HostPort: "localhost:7233", // Temporal Server адрес
	})
	if err != nil {
		logger.Error("Failed to create Temporal client", "error", err)
		os.Exit(1)
	}
	defer temporalClient.Close()

	w := worker.New(temporalClient, workflow.OrderProcessingTaskQueue, worker.Options{})

	w.RegisterActivity(createOrderActivity.Execute)
	w.RegisterActivity(checkInventoryActivity.Execute)
	w.RegisterActivity(processPaymentActivity.Execute)
	w.RegisterActivity(sendNotificationActivity.Execute)
	w.RegisterActivity(cancelOrderActivity.Execute)

	w.RegisterWorkflow(workflow.OrderProcessingWorkflow)

	httpServer := httpserver.NewServer(8080, temporalClient)

	go func() {
		logger.Info("Starting Temporal Worker...")
		if err := w.Run(worker.InterruptCh()); err != nil {
			logger.Error("Failed to start worker", "error", err)
			os.Exit(1)
		}
	}()

	go func() {
		if err := httpServer.Start(); err != nil && err != http.ErrServerClosed {
			logger.Error("Failed to start HTTP server", "error", err)
			os.Exit(1)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down OrderFlow application...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("Failed to shutdown HTTP server gracefully", "error", err)
	}

	logger.Info("OrderFlow application stopped")
}

func startOrderWorkflow(temporalClient client.Client, input *workflow.OrderProcessingInput) (client.WorkflowRun, error) {
	workflowOptions := client.StartWorkflowOptions{
		ID:        "order-processing-" + input.CustomerID,
		TaskQueue: workflow.OrderProcessingTaskQueue,
	}

	return temporalClient.ExecuteWorkflow(context.Background(), workflowOptions, workflow.OrderProcessingWorkflow, input)
}

func cancelOrderWorkflow(temporalClient client.Client, workflowID string) error {
	return temporalClient.CancelWorkflow(context.Background(), workflowID, "")
}

func getWorkflowStatus(temporalClient client.Client, workflowID string) (interface{}, error) {
	var result interface{}
	_, err := temporalClient.QueryWorkflow(context.Background(), workflowID, "", workflow.OrderStatusQuery, &result)
	return result, err
}
