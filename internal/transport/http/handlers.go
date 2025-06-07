package http

import (
	"net/http"

	"github.com/sarangkharche/xapien-billing/internal/usecase"
)

func NewHandler(uc *usecase.UseCases) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/set-plan", uc.HandleSetPlan)
	mux.HandleFunc("/set-user-limit", uc.HandleSetUserLimit)
	mux.HandleFunc("/get-org", uc.HandleGetOrg)
	mux.HandleFunc("/use-credit", uc.HandleUseCredit)
	mux.HandleFunc("/top-up", uc.HandleTopUp)
	return mux
}
