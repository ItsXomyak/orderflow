package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"orderflow/internal/adapter/repository"
	"orderflow/internal/domain/workflow"
	"orderflow/internal/httpserver"
	activ "orderflow/internal/usecase/activity"
	"orderflow/internal/usecase/service"
	usecaseWorkflow "orderflow/internal/usecase/workflow"
	"orderflow/pkg/logger"
)

func main() {
	appEnv := getEnv("APP_ENV", "development")
	postgresHost := getEnv("POSTGRES_HOST", "localhost")
	postgresPort := getEnv("POSTGRES_PORT", "5432")
	postgresDB := getEnv("POSTGRES_DB", "orderflow")
	postgresUser := getEnv("POSTGRES_USER", "postgres")
	postgresPassword := getEnv("POSTGRES_PASSWORD", "password")

	logger.Init(appEnv)

	logger.Info("Starting OrderFlow application...")

	postgresURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		postgresUser, postgresPassword, postgresHost, postgresPort, postgresDB)

	pool, err := pgxpool.New(context.Background(), postgresURL)
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

	createOrderActivity := activ.NewCreateOrderActivity(orderService)
	checkInventoryActivity := activ.NewCheckInventoryActivity(inventoryService, orderService)
	processPaymentActivity := activ.NewProcessPaymentActivity(paymentService, orderService, inventoryService)
	sendNotificationActivity := activ.NewSendNotificationActivity(notificationService, orderService)
	cancelOrderActivity := activ.NewCancelOrderActivity(orderService, paymentService, inventoryService)

	temporalClient, err := newTemporalClient()
	if err != nil {
	logger.Error("Failed to create Temporal client", "error", err)
	os.Exit(1)
}
defer temporalClient.Close()

w := worker.New(temporalClient, workflow.OrderProcessingTaskQueue, worker.Options{})

w.RegisterActivityWithOptions(createOrderActivity.Execute, activity.RegisterOptions{
    Name: "CreateOrderActivity",
})
w.RegisterActivityWithOptions(checkInventoryActivity.Execute, activity.RegisterOptions{
    Name: "CheckInventoryActivity",
})
w.RegisterActivityWithOptions(processPaymentActivity.Execute, activity.RegisterOptions{
    Name: "ProcessPaymentActivity",
})
w.RegisterActivityWithOptions(sendNotificationActivity.Execute, activity.RegisterOptions{
    Name: "SendNotificationActivity",
})
w.RegisterActivityWithOptions(cancelOrderActivity.Execute, activity.RegisterOptions{
    Name: "CancelOrderActivity",
})

w.RegisterWorkflow(usecaseWorkflow.OrderProcessingWorkflow)

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

func newTemporalClient() (client.Client, error) {
	addr := getEnv("TEMPORAL_ADDRESS", "TEMPORAL_HOST:TEMPORAL_PORT")
	if addr == "" {
		host := getEnv("TEMPORAL_HOST", "temporal") // имя сервиса из docker-compose
		port := getEnv("TEMPORAL_PORT", "7233")
		addr = fmt.Sprintf("%s:%s", host, port)
	}

	var c client.Client
	var err error
	for attempt := 0; attempt < 8; attempt++ {
		c, err = client.Dial(client.Options{
			HostPort:  addr,
			Namespace: getEnv("TEMPORAL_NAMESPACE", "default"),
		})
		if err == nil {
			return c, nil
		}
		time.Sleep(time.Second * time.Duration(1<<attempt)) // 1s,2s,4s,...
	}
	return nil, err
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
