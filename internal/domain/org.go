package domain

import (
	"log"
	"time"
)

type PlanType string

const (
	PlanUltimate   PlanType = "Ultimate"
	PlanEnterprise PlanType = "Enterprise"
	PlanBasic      PlanType = "Basic"
	PlanLite       PlanType = "Lite"
	PlanTrial      PlanType = "Trial"
)

// plan limits for each plan
var PlanLimits = map[PlanType]int{
	PlanUltimate:   1000,
	PlanEnterprise: 500,
	PlanBasic:      100,
	PlanLite:       20,
	PlanTrial:      10,
}

type Organisation struct {
	ID                   string            `json:"id" dynamodbav:"org_id"`
	Plan                 PlanType          `json:"plan"`
	UsedCredits          int               `json:"used_credits" dynamodbav:"used_credits"`
	MonthlyUsageRemaining int               `json:"monthly_usage_remaining" dynamodbav:"monthly_usage_remaining"`
	UserUsage            map[string]int    `json:"user_usage" dynamodbav:"user_usage"`
	TopUpCredits         int               `json:"top_up_credits" dynamodbav:"top_up_credits"`
	PerUserLimit         int               `json:"per_user_limit" dynamodbav:"per_user_limit"`
	LastResetDate        time.Time         `json:"last_reset_date" dynamodbav:"last_reset_date"`
	NotificationSent     bool              `json:"notification_sent" dynamodbav:"notification_sent"`
	TotalCredits         int               `json:"total_credits" dynamodbav:"total_credits"`
}

// get total credits available for the month
func (o *Organisation) GetTotalCredits() int {
	planLimit := PlanLimits[o.Plan]
	return planLimit + o.TopUpCredits
}

// update total credits field
func (o *Organisation) UpdateTotalCredits() {
	oldTotal := o.TotalCredits
	o.TotalCredits = o.GetTotalCredits()
	o.MonthlyUsageRemaining = o.TotalCredits - o.UsedCredits
	log.Printf("Updated TotalCredits for org %s: %d -> %d (Plan: %s, TopUpCredits: %d, UsedCredits: %d, Remaining: %d)\n",
		o.ID, oldTotal, o.TotalCredits, o.Plan, o.TopUpCredits, o.UsedCredits, o.MonthlyUsageRemaining)
}

// get remaining credits for the month
func (o *Organisation) GetRemainingCredits() int {
	return o.MonthlyUsageRemaining
}

// get usage for a specific user
func (o *Organisation) GetUserUsage(userID string) int {
	if o.UserUsage == nil {
		return 0
	}
	return o.UserUsage[userID]
}

// check if a user can run a report
func (o *Organisation) CanUserRunReport(userID string) bool {
	// check org-level credits
	if o.GetRemainingCredits() <= 0 {
		return false
	}

	// check per-user limit if set
	if o.PerUserLimit > 0 {
		userUsage := o.GetUserUsage(userID)
		if userUsage >= o.PerUserLimit {
			return false
		}
	}

	return true
}

// use credit increments usage counters
func (o *Organisation) UseCredit(userID string) {
	o.UsedCredits++
	o.MonthlyUsageRemaining = o.TotalCredits - o.UsedCredits
	if o.UserUsage == nil {
		o.UserUsage = make(map[string]int)
	}
	o.UserUsage[userID]++
}

// reset usage counters for a new month
func (o *Organisation) ResetMonthlyUsage() {
	o.UsedCredits = 0
	o.MonthlyUsageRemaining = o.TotalCredits
	o.UserUsage = make(map[string]int)
	o.TopUpCredits = 0
	o.NotificationSent = false
	o.LastResetDate = time.Now()
	o.UpdateTotalCredits()
}

// check if we should reset for a new calendar month
func (o *Organisation) ShouldResetForNewMonth() bool {
	now := time.Now()
	return now.Month() != o.LastResetDate.Month() || now.Year() != o.LastResetDate.Year()
}

// check if usage is at 80% or more of total credits
func (o *Organisation) IsApproachingLimit() bool {
	usage := float64(o.UsedCredits)
	total := float64(o.GetTotalCredits())
	return usage/total >= 0.8
}
