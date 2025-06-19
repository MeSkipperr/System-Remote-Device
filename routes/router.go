package routes

import (
	"github.com/gorilla/mux"
	"SystemRemoteDevice/handlers"
	"SystemRemoteDevice/handlers/api"
)

func RegisterRoutes() *mux.Router {
	r := mux.NewRouter()

	// Web UI
	r.HandleFunc("/", handlers.HomeHandler)

	// Project / Monitoring routes
	project := r.PathPrefix("/project/monitoring-network").Subrouter()
	project.HandleFunc("/start", handlers.StartMonitoringHandler)
	project.HandleFunc("/status", handlers.StatusMonitoringHandler)
	project.HandleFunc("/stop", handlers.StopMonitoringHandler)

	// API
	apiRouter := r.PathPrefix("/api/device").Subrouter()
	apiRouter.HandleFunc("/id/{id}", api.GetDeviceByID).Methods("GET")
	apiRouter.HandleFunc("/category/{category}", api.GetDevicesByCategory).Methods("GET")
	apiRouter.HandleFunc("", api.AddDevice).Methods("POST")

	return r
}
