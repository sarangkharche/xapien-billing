package usecase

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/sarangkharche/xapien-billing/internal/domain"
	"github.com/sarangkharche/xapien-billing/internal/infrastructure/dao"
)

type UseCases struct {
	DAO                 dao.OrgDAO
	NotificationService domain.NotificationService
}

func (uc *UseCases) HandleSetPlan(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OrgID string `json:"org_id"`
		Plan  string `json:"plan"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	org := domain.Organisation{
		ID:                   req.OrgID,
		Plan:                 domain.PlanType(req.Plan),
		UsedCredits:          0,
		MonthlyUsageRemaining: 0,
		UserUsage:            make(map[string]int),
		TopUpCredits:         0,
		PerUserLimit:         0,
		LastResetDate:        time.Now(),
	}
	org.UpdateTotalCredits()

	if err := uc.DAO.SaveOrg(org); err != nil {
		log.Printf("Failed to save org: %v\n", err)
		http.Error(w, "Failed to save org", http.StatusInternalServerError)
		return
	}

	log.Printf("Org %s saved with plan %s\n", org.ID, org.Plan)
	w.WriteHeader(http.StatusOK)
}

func (uc *UseCases) HandleSetUserLimit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OrgID string `json:"org_id"`
		Limit int    `json:"limit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	org, err := uc.DAO.GetOrg(req.OrgID)
	if err != nil {
		log.Printf("Org %s not found: %v\n", req.OrgID, err)
		http.Error(w, "Organisation not found", http.StatusNotFound)
		return
	}

	org.PerUserLimit = req.Limit

	if err := uc.DAO.SaveOrg(org); err != nil {
		log.Printf("Failed to update user limit: %v\n", err)
		http.Error(w, "Failed to update user limit", http.StatusInternalServerError)
		return
	}

	log.Printf("Set user limit for org %s to %d\n", org.ID, org.PerUserLimit)
	w.WriteHeader(http.StatusOK)
}

func (uc *UseCases) HandleGetOrg(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("org_id")
	if orgID == "" {
		http.Error(w, "org_id parameter is required", http.StatusBadRequest)
		return
	}

	org, err := uc.DAO.GetOrg(orgID)
	if err != nil {
		log.Printf("Org %s not found: %v\n", orgID, err)
		http.Error(w, "Organisation not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(org); err != nil {
		log.Printf("Failed to encode response: %v\n", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	log.Printf("Retrieved org %s\n", org.ID)
}

func (uc *UseCases) HandleUseCredit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OrgID     string `json:"org_id"`
		UserID    string `json:"user_id"`
		EnquiryID string `json:"enquiry_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	org, err := uc.DAO.GetOrg(req.OrgID)
	if err != nil {
		log.Printf("Org %s not found: %v\n", req.OrgID, err)
		http.Error(w, "Organisation not found", http.StatusNotFound)
		return
	}

	if org.ShouldResetForNewMonth() {
		org.ResetMonthlyUsage()
		log.Printf("Reset monthly usage for org %s\n", org.ID)
	}

	if !org.CanUserRunReport(req.UserID) {
		log.Printf("User %s in org %s cannot run report - limit reached\n", req.UserID, req.OrgID)

		if !org.NotificationSent {
			if err := uc.NotificationService.SendLimitReachedNotification(org); err != nil {
				log.Printf("Failed to send limit reached notification: %v\n", err)
			}
			org.NotificationSent = true
			uc.DAO.SaveOrg(org) // Save notification sent status
		}

		http.Error(w, "Credit limit reached", http.StatusForbidden)
		return
	}

	org.UseCredit(req.UserID)

	if org.IsApproachingLimit() && !org.NotificationSent {
		log.Printf("Organization %s has reached 80%% of their credit limit (Used: %d/%d credits)\n",
			org.ID, org.UsedCredits, org.TotalCredits)
		if err := uc.NotificationService.SendApproachingLimitNotification(org); err != nil {
			log.Printf("Failed to send approaching limit notification: %v\n", err)
		}
		org.NotificationSent = true
	}

	if err := uc.DAO.SaveOrg(org); err != nil {
		log.Printf("Failed to update usage: %v\n", err)
		http.Error(w, "Failed to update usage", http.StatusInternalServerError)
		return
	}

	log.Printf("Credit used for enquiry %s by user %s in org %s. Remaining credits: %d\n",
		req.EnquiryID, req.UserID, req.OrgID, org.GetRemainingCredits())

	response := map[string]interface{}{
		"enquiry_id":        req.EnquiryID,
		"remaining_credits": org.GetRemainingCredits(),
		"user_usage":        org.GetUserUsage(req.UserID),
		"total_credits":     org.TotalCredits,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (uc *UseCases) HandleTopUp(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OrgID   string `json:"org_id"`
		Credits int    `json:"credits"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if req.Credits <= 0 {
		http.Error(w, "Credits must be positive", http.StatusBadRequest)
		return
	}

	org, err := uc.DAO.GetOrg(req.OrgID)
	if err != nil {
		log.Printf("Org %s not found: %v\n", req.OrgID, err)
		http.Error(w, "Organisation not found", http.StatusNotFound)
		return
	}

	if org.ShouldResetForNewMonth() {
		org.ResetMonthlyUsage()
		log.Printf("Reset monthly usage for org %s\n", org.ID)
	}

	org.TopUpCredits += req.Credits
	org.UpdateTotalCredits()

	if err := uc.DAO.SaveOrg(org); err != nil {
		log.Printf("Failed to add top-up credits: %v\n", err)
		http.Error(w, "Failed to add top-up credits", http.StatusInternalServerError)
		return
	}

	log.Printf("Added %d top-up credits to org %s. Total credits: %d\n",
		req.Credits, req.OrgID, org.TotalCredits)

	response := map[string]interface{}{
		"total_credits":     org.TotalCredits,
		"remaining_credits": org.GetRemainingCredits(),
		"top_up_credits":    org.TopUpCredits,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
