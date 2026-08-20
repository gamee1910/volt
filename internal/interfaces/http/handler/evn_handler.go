package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gamee1910/volt/internal/client"
	"github.com/gamee1910/volt/internal/interfaces/http/handler/request"
	"github.com/gamee1910/volt/pkg/logger"
)

func LoginHandler(evnClient client.EvnhcmcClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := evnClient.Login(r.Context()); err != nil {
			logger.Error("EVNHCMC login failed", "error", err)
			http.Error(w, "EVNHCMC login failed", http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte(`{"message":"login success"}`))
	}
}

func GetDailyPowerUsageHandler(evnClient client.EvnhcmcClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			logger.Error("Parse multipart form failed", "error", err)
			http.Error(w, "Invalid multipart form", http.StatusBadRequest)
			return
		}

		req := request.DailyPowerUsageRequest{
			CustomerCode: r.FormValue("input_makh"),
			FromDate:     r.FormValue("input_tungay"),
			ToDate:       r.FormValue("input_denngay"),
			Token:        r.FormValue("token"),
		}

		if req.CustomerCode == "" || req.FromDate == "" || req.ToDate == "" {
			http.Error(w, "Thiếu thông tin: input_makh, input_tungay hoặc input_denngay", http.StatusBadRequest)
			return
		}

		result, err := evnClient.GetDailyPowerUsageData(r.Context(), req)
		if err != nil {
			logger.Error("EVNHCMC tra cứu điện năng thất bại", "error", err)
			http.Error(w, "Tra cứu điện năng thất bại", http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(result); err != nil {
			logger.Error("Encode response failed", "error", err)
		}
	}
}
