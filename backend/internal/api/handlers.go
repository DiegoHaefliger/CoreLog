package api

import (
	"encoding/json"
	"net/http"
	"database/sql"

	"corelog/backend/internal/auth"
	"corelog/backend/internal/db"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LogEntry struct {
	ID        int    `json:"id"`
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	SourceApp string `json:"source_app"`
	Message   string `json:"message"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var hash string
	var reqChange bool
	err := db.DB.QueryRow("SELECT password_hash, requires_password_change FROM users WHERE username = ?", req.Username).Scan(&hash, &reqChange)
	
	// Default to early exit if user missing
	if err != nil || !auth.CheckPasswordHash(req.Password, hash) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateJWT(req.Username)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	auth.SetAuthCookie(w, token)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success", 
		"token": token,
		"requires_password_change": reqChange,
	})
}

func GetLogsHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	service := r.URL.Query().Get("service")
	level := r.URL.Query().Get("level")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	
	var rows *sql.Rows
	var err error

	// Build dynamic query
	baseQuery := "SELECT id, timestamp, level, source_app, message FROM logs"
	var args []interface{}

	if query != "" {
		baseQuery += " WHERE id IN (SELECT rowid FROM logs_fts WHERE message MATCH ?)"
		args = append(args, query)
	} else {
		baseQuery += " WHERE 1=1" // placeholder to easily append AND
	}

	if service != "" {
		baseQuery += " AND source_app = ?"
		args = append(args, service)
	}

	if level != "" {
		baseQuery += " AND UPPER(level) = UPPER(?)"
		args = append(args, level)
	}

	if from != "" {
		// Normaliza timestamps ISO (2026-04-01T12:00:05Z → 2026-04-01 12:00:05) para comparação correta
		baseQuery += " AND REPLACE(REPLACE(timestamp, 'T', ' '), 'Z', '') >= ?"
		args = append(args, from)
	}

	if to != "" {
		baseQuery += " AND REPLACE(REPLACE(timestamp, 'T', ' '), 'Z', '') <= ?"
		args = append(args, to)
	}

	baseQuery += " ORDER BY id DESC LIMIT 100"
	
	rows, err = db.DB.Query(baseQuery, args...)

	if err != nil {
		http.Error(w, "Error fetching logs: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var logs []LogEntry
	for rows.Next() {
		var logEntry LogEntry
		if err := rows.Scan(&logEntry.ID, &logEntry.Timestamp, &logEntry.Level, &logEntry.SourceApp, &logEntry.Message); err != nil {
			continue
		}
		logs = append(logs, logEntry)
	}

	w.Header().Set("Content-Type", "application/json")
	if len(logs) == 0 {
		logs = make([]LogEntry, 0)
	}
	json.NewEncoder(w).Encode(logs)
}

func GetServicesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query("SELECT DISTINCT source_app FROM logs WHERE source_app != '' AND source_app != 'unknown' ORDER BY source_app")
	if err != nil {
		http.Error(w, "Error fetching services: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var services []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err == nil {
			services = append(services, s)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if len(services) == 0 {
		services = make([]string, 0)
	}
	json.NewEncoder(w).Encode(services)
}

type ChangePasswordRequest struct {
	NewPassword string `json:"new_password"`
}

func ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	username, ok := r.Context().Value("username").(string)
	if !ok || username == "" {
		http.Error(w, "Unauthorized context", http.StatusUnauthorized)
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.NewPassword == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	newHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	_, err = db.DB.Exec("UPDATE users SET password_hash = ?, requires_password_change = 0 WHERE username = ?", newHash, username)
	if err != nil {
		http.Error(w, "Error updating password", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"success"}`))
}
