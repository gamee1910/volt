package handler

import (
	"net/http"

	"github.com/gamee1910/volt/config"
	"github.com/gamee1910/volt/internal/application/dto"
	"github.com/gamee1910/volt/internal/domain/service"
	"github.com/gamee1910/volt/internal/interfaces/api/handler/request"
	pkgjson "github.com/gamee1910/volt/pkg/json"
)

type ElectricityHandler struct {
	electricityService service.ElectricityService
	cfg                *config.Configuration
}

func NewElectricityHandler(
	electrictiService service.ElectricityService, cfg *config.Configuration,
) *ElectricityHandler {
	return &ElectricityHandler{
		electricityService: electrictiService, cfg: cfg,
	}
}

func (h *ElectricityHandler) SyncFromEVNHandler(w http.ResponseWriter, r *http.Request) {
	var req request.GetUsageRequest

	if err := pkgjson.DecodeBody(r, &req); err != nil {
		req.FromDate = r.URL.Query().Get("from_date")
		req.ToDate = r.URL.Query().Get("to_date")
	}

	if req.FromDate == "" || req.ToDate == "" {
		pkgjson.WriteError(w, http.StatusBadRequest, "customer_code, from_date, and to_date are required")
		return
	}

	err := h.electricityService.DailyPowerUsage(r.Context(), dto.GetUsageParam{
		FromDate: req.FromDate,
		ToDate:   req.ToDate,
	})
	if err != nil {
		pkgjson.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	pkgjson.WriteOK(w, map[string]string{"message": "Sync successful"})
}

func (h *ElectricityHandler) GetAllHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := h.electricityService.GetAll(r.Context())
	if err != nil {
		pkgjson.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	pkgjson.WriteOK(w, resp)
}

func (h *ElectricityHandler) GetYesterdayUsageHandler(w http.ResponseWriter, r *http.Request) {
	response, err := h.electricityService.GetYesterDayUsage(r.Context())
	if err != nil {
		pkgjson.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	pkgjson.WriteOK(w, response)
}
