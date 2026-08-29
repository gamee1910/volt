package handler

import (
	"net/http"

	"github.com/gamee1910/volt/config"
	"github.com/gamee1910/volt/internal/application"
	"github.com/gamee1910/volt/internal/application/command"
	"github.com/gamee1910/volt/internal/application/query"
	"github.com/gamee1910/volt/internal/common/json"
	"github.com/gamee1910/volt/internal/interfaces/api/handler/request"
)

type ElectricityHandler struct {
	app *application.Application
	cfg *config.Configuration
}

func NewElectricityHandler(
	app *application.Application, cfg *config.Configuration,
) *ElectricityHandler {
	return &ElectricityHandler{app: app, cfg: cfg}
}

func (h *ElectricityHandler) SyncFromEVNHandler(w http.ResponseWriter, r *http.Request) {
	var req request.GetUsageRequest

	if err := json.DecodeBody(r, &req); err != nil {
		req.FromDate = r.URL.Query().Get("from_date")
		req.ToDate = r.URL.Query().Get("to_date")
	}

	if req.FromDate == "" || req.ToDate == "" {
		json.WriteError(w, http.StatusBadRequest, "customer_code, from_date, and to_date are required")
		return
	}

	err := h.app.Commands.SyncElectricity.Handle(r.Context(), command.SyncElectricityCommand{
		FromDate: req.FromDate,
		ToDate:   req.ToDate,
	})
	if err != nil {
		json.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	json.WriteOK(w, map[string]string{"message": "Sync successful"})
}

func (h *ElectricityHandler) GetAllHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := h.app.Queries.AllElectricity.Handle(r.Context(), query.AllElectricityQuery{})
	if err != nil {
		json.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	json.WriteOK(w, resp)
}

func (h *ElectricityHandler) GetYesterdayUsageHandler(w http.ResponseWriter, r *http.Request) {
	response, err := h.app.Queries.YesterdayUsage.Handle(r.Context(), query.YesterdayUsageQuery{})
	if err != nil {
		json.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	json.WriteOK(w, response)
}
