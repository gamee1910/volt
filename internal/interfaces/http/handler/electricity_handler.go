package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gamee1910/volt/internal/config"
	"github.com/gamee1910/volt/internal/domain/service"
	"github.com/gamee1910/volt/internal/interfaces/http/handler/request"
	"github.com/gamee1910/volt/pkg/evnhcm"
)

type ElectricityHandler struct {
	electricityService service.ElectricityService
	cfg                *config.Configuration
}

func NewElectricityHandler(
	electrictiService service.ElectricityService,
	cfg *config.Configuration,
) *ElectricityHandler {
	return &ElectricityHandler{
		electricityService: electrictiService,
		cfg:                cfg,
	}
}

func (h *ElectricityHandler) Login(w http.ResponseWriter, r *http.Request) {
	err := h.electricityService.LoginEVN(
		r.Context(),
		h.cfg.EnvConfig.Username,
		h.cfg.EnvConfig.Password,
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	response := map[string]interface{}{
		"message": "Login successful",
	}
	responseJSON, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	_, _ = w.Write(responseJSON)
}

func (h *ElectricityHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	var req request.GetUsageRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.FromDate = r.URL.Query().Get("from_date")
		req.ToDate = r.URL.Query().Get("to_date")
	}

	if req.FromDate == "" || req.ToDate == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("customer_code, from_date, and to_date are required"))
		return
	}

	evnReq := evnhcm.DailyPowerUsageRequest{
		Token:        "",
		CustomerCode: h.cfg.EnvConfig.CustomerCode,
		FromDate:     req.FromDate,
		ToDate:       req.ToDate,
	}

	result, err := h.electricityService.FetchAndSyncMonthlyUsage(r.Context(), evnReq)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}
