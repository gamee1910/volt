package routes

import (
	"database/sql"
	"net/http"

	"github.com/gamee1910/volt/internal/infrastructure/di"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func SetupRouter(db *sql.DB, container *di.Container) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.ClientIPFromRemoteAddr)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := db.PingContext(r.Context()); err != nil {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Post("/login", container.ElectricityHandler().Login)
			r.Post("/sync", container.ElectricityHandler().SyncFromEVN)
			r.Get("/", container.ElectricityHandler().GetAll)
			r.Get("/yesterday", container.ElectricityHandler().GetYesterdayUsage)
		})
	})

	return r
}
