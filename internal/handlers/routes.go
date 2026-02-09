package handlers

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func SetupRoutes(orderHandler *OrderHandler) *mux.Router {
	router := mux.NewRouter()

	// Middleware
	router.Use(loggingMiddleware)
	router.Use(corsMiddleware)

	// API v1 routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Only POST and GET (list) endpoints
	api.HandleFunc("/orders", orderHandler.CreateOrder).Methods("POST")
	api.HandleFunc("/orders", orderHandler.ListOrders).Methods("GET")

	return router
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
