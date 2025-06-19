package api

import (
	"SystemRemoteDevice/models"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "modernc.org/sqlite"
	"github.com/go-playground/validator/v10"

)
var validate = validator.New()

// DeviceRepository fetches devices from DB by custom WHERE clause
func fetchDevicesByQuery(query string, arg interface{}) ([]models.DeviceType, error) {
	db, err := sql.Open("sqlite", "file:./resource/app.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(query, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []models.DeviceType
	for rows.Next() {
		var d models.DeviceType
		err := rows.Scan(
			&d.ID,
			&d.Name,
			&d.IPAddress,
			&d.Device,
			&d.Error,
			&d.Description,
			&d.DownTime,
			&d.Type,
		)
		if err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return devices, nil
}

// writeJSON is a helper function to send JSON responses
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

//* GET HANDLERS
// GetDeviceByID returns single device data by ID
func GetDeviceByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	devices, err := fetchDevicesByQuery(`SELECT * FROM devices WHERE id = ?`, id)
	if err != nil {
		log.Println("Error fetching device by ID:", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	if len(devices) == 0 {
		http.Error(w, `{"error":"Device not found"}`, http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, devices[0]) // Kirim hanya 1 device
}

// GetDevicesByType returns list of devices by type
func GetDevicesByCategory(w http.ResponseWriter, r *http.Request) {
	category := mux.Vars(r)["category"]

	devices, err := fetchDevicesByQuery(`SELECT * FROM devices WHERE type = ?`, category)
	if err != nil {
		log.Println("Error fetching devices by type:", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}

	if len(devices) == 0 {
		http.Error(w, `{"error":"No devices found for this type"}`, http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, devices)
}

//*Post Handlers
// AddDevice adds a new device to the database

func AddDevice(w http.ResponseWriter, r *http.Request) {
	var device models.DeviceType
	if err := json.NewDecoder(r.Body).Decode(&device); err != nil {
		log.Println("Error decoding request body:", err)
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := validate.Struct(device); err != nil {
		log.Println("Validation error:", err)
		http.Error(w, `{"error":"Validation failed"}`, http.StatusBadRequest)
		return
	}

	db, err := sql.Open("sqlite", "file:./resource/app.db")
	if err != nil {
		log.Println("Error connecting to database:", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := `INSERT INTO devices (name, ip_address, device, error, description, down_time, type) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err = db.Exec(query, device.Name, device.IPAddress, device.Device, device.Error, device.Description, device.DownTime, device.Type)
	if err != nil {
		log.Println("Error inserting device:", err)
		http.Error(w, `{"error":"Failed to add device"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"message": "Device added successfully"})
}