package interfaces

import (
	"database/sql"
	"net/http"

	"github.com/gamee1910/volt/internal/infrastructure/evnhcmc"
	"github.com/gamee1910/volt/internal/interfaces/http/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func SetupRouter(db *sql.DB, evnClient *evnhcmc.EVNClient) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.ClientIPFromRemoteAddr)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := db.PingContext(r.Context()); err != nil {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	router.Post("/login", handler.LoginHandler(evnClient))
	router.Post("/daily", handler.GetDailyPowerUsageHandler(evnClient))
	return router
}
