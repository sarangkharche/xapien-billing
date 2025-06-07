package domain

type NotificationService interface {
	SendApproachingLimitNotification(org Organisation) error
	SendLimitReachedNotification(org Organisation) error
}

type MockNotificationService struct{}

func (m *MockNotificationService) SendApproachingLimitNotification(org Organisation) error {
	return nil
}

func (m *MockNotificationService) SendLimitReachedNotification(org Organisation) error {
	return nil
}