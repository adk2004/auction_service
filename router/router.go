package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func SetupRouter() http.Handler {
	r := chi.NewRouter()
	
	//  auction routes
	// r.Post("/auction", )
	// r.Post("/auction/bid", )
	// r.Get("/auction/{id}", )

	// test route
	r.Get("/test", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "Test Success"}`))
	})

	return r
}
