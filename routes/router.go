package routes

import (
	"SystemRemoteDevice/handlers"
	"SystemRemoteDevice/handlers/api"
	"SystemRemoteDevice/handlers/api/auth"

	"github.com/gorilla/mux"
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

	// User API
	userRouter := r.PathPrefix("/api/user").Subrouter()
	userRouter.HandleFunc("", api.GetAllUsers).Methods("GET")
	userRouter.HandleFunc("", api.AddUser).Methods("POST")

	// Auth API
	authRouter := r.PathPrefix("/api/auth").Subrouter()
	authRouter.HandleFunc("/login", auth.LoginHandler).Methods("POST")
	return r
}
