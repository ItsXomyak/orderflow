package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"

	"order-processing/internal/application/workflow"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("Failed to create logger", err)
	}
	defer logger.Sync()

	c, err := client.NewClient(client.Options{
		HostPort: "localhost:7233",
	})
	if err != nil {
		logger.Fatal("Failed to create Temporal client", zap.Error(err))
	}
	defer c.Close()

	w := worker.New(c, workflow.TaskQueueOrderProcessing, worker.Options{})

	w.RegisterWorkflow(workflow.OrderWorkflow)
	w.RegisterWorkflow(workflow.OrderWorkflowWithCancellation)

	w.RegisterActivity(workflow.CreateOrderActivity)
	w.RegisterActivity(workflow.CheckInventoryActivity)
	w.RegisterActivity(workflow.ProcessPaymentActivity)
	w.RegisterActivity(workflow.SendNotificationActivity)
	w.RegisterActivity(workflow.CancelOrderActivity)

	logger.Info("Starting worker", zap.String("taskQueue", workflow.TaskQueueOrderProcessing))

	err = w.Run(worker.InterruptCh())
	if err != nil {
		logger.Fatal("Failed to start worker", zap.Error(err))
	}

	logger.Info("Worker stopped")
}