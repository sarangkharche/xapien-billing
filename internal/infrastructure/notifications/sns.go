package notifications

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/sarangkharche/xapien-billing/internal/domain"
)

type SNSNotificationService struct {
	client   *sns.Client
	topicArn string
}

func NewSNSNotificationService(cfg aws.Config, topicArn string) *SNSNotificationService {
	return &SNSNotificationService{
		client:   sns.NewFromConfig(cfg),
		topicArn: topicArn,
	}
}

func (s *SNSNotificationService) SendApproachingLimitNotification(org domain.Organisation) error {
	subject := fmt.Sprintf("Credit Limit Warning - %s (80%% usage)", org.ID)
	message := fmt.Sprintf(`
Dear Customer Success Team,

Organization "%s" is approaching their monthly credit limit.

Current Status:
- Plan: %s
- Total Credits: %d
- Used Credits: %d
- Remaining Credits: %d
- Usage Percentage: %.1f%%

This is an automated notification to help you proactively reach out to the customer about upgrading their plan.

Best regards,
Xapien Billing System
`, org.ID, org.Plan, org.GetTotalCredits(), org.MonthlyUsageRemaining, org.GetRemainingCredits(),
		float64(org.MonthlyUsageRemaining)/float64(org.GetTotalCredits())*100)

	input := &sns.PublishInput{
		TopicArn: &s.topicArn,
		Subject:  &subject,
		Message:  &message,
	}

	_, err := s.client.Publish(context.TODO(), input)
	if err != nil {
		log.Printf("Failed to send approaching limit notification via SNS: %v\n", err)
		return fmt.Errorf("SNS publish error: %w", err)
	}

	log.Printf("Sent approaching limit notification for org %s via SNS\n", org.ID)
	return nil
}

func (s *SNSNotificationService) SendLimitReachedNotification(org domain.Organisation) error {
	subject := fmt.Sprintf("Credit Limit REACHED - %s", org.ID)
	message := fmt.Sprintf(`
Dear Customer Success Team,

Organization "%s" has REACHED their monthly credit limit and is now blocked from running reports.

Current Status:
- Plan: %s
- Total Credits: %d
- Used Credits: %d
- Remaining Credits: %d

URGENT: Please contact the customer immediately to discuss upgrading their plan or adding top-up credits.

User Activity:
`, org.ID, org.Plan, org.GetTotalCredits(), org.MonthlyUsageRemaining, org.GetRemainingCredits())

	// Add user usage details
	for userID, usage := range org.UserUsage {
		message += fmt.Sprintf("- User %s: %d reports\n", userID, usage)
	}

	message += `
Best regards,
Xapien Billing System
`

	input := &sns.PublishInput{
		TopicArn: &s.topicArn,
		Subject:  &subject,
		Message:  &message,
	}

	_, err := s.client.Publish(context.TODO(), input)
	if err != nil {
		log.Printf("Failed to send limit reached notification via SNS: %v\n", err)
		return fmt.Errorf("SNS publish error: %w", err)
	}

	log.Printf("Sent limit reached notification for org %s via SNS\n", org.ID)
	return nil
}