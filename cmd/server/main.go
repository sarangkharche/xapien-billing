package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/sarangkharche/xapien-billing/internal/domain"
	"github.com/sarangkharche/xapien-billing/internal/infrastructure/dao"
	"github.com/sarangkharche/xapien-billing/internal/infrastructure/notifications"
	apphttp "github.com/sarangkharche/xapien-billing/internal/transport/http"
	"github.com/sarangkharche/xapien-billing/internal/usecase"
)

func main() {
	ctx := context.Background()

	// 	load aws configuration
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	// create dynamodb client
	dbClient := dynamodb.NewFromConfig(cfg)

	// confirm dynamodb connectivity
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	log.Println("Testing connection to DynamoDB...")
	_, err = dbClient.ListTables(ctxTimeout, &dynamodb.ListTablesInput{})
	if err != nil {
		log.Fatalf("Failed to connect to DynamoDB: %v", err)
	}
	log.Println("Connected to DynamoDB successfully")

	// initialize notification service
	var notificationService domain.NotificationService
	snsTopicArn := os.Getenv("SNS_TOPIC_ARN")
	if snsTopicArn != "" {
		log.Printf("Using SNS notification service with topic: %s\n", snsTopicArn)
		notificationService = notifications.NewSNSNotificationService(cfg, snsTopicArn)
	} else {
		log.Println("SNS_TOPIC_ARN not set, using mock notification service")
		notificationService = &domain.MockNotificationService{}
	}

	// initialize dao and handlers
	tableName := "organisation_usage"
	orgDAO := dao.NewDynamoOrgDAO(cfg, tableName)

	uc := &usecase.UseCases{
		DAO:                 orgDAO,
		NotificationService: notificationService,
	}
	handler := apphttp.NewHandler(uc)

	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
