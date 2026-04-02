package main

import (
	"flag"
	"log"
	"net/http"

	"corelog/backend"
	"corelog/backend/internal/api"
	"corelog/backend/internal/auth"
	"corelog/backend/internal/db"
	"corelog/backend/internal/ingest"
)

func main() {
	var logPath string
	var dbPath string
	var useDocker bool
	flag.StringVar(&logPath, "log-path", "data/application.log", "Caminho do arquivo de log a ser monitorado pelo servidor")
	flag.StringVar(&dbPath, "db-path", "data/corelog.db", "Caminho do arquivo SQLite interno")
	flag.BoolVar(&useDocker, "docker", false, "Coletar logs de todos os containers Docker automaticamente")
	flag.Parse()

	err := db.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Fatal: failed to initialize database: %v", err)
	}
	defer db.DB.Close()

	// Cria usuário admin padrão apenas se não existir
	adminHash, _ := auth.HashPassword("admin")
	db.DB.Exec("INSERT OR IGNORE INTO users (username, password_hash, requires_password_change) VALUES (?, ?, 1)", "admin", adminHash)

	// Iniciar pipeline de Ingestão de Logs
	ingest.StartPipeline(logPath, useDocker)

	// Setup multiplexer
	mux := http.NewServeMux()
	
	// Public routes
	mux.HandleFunc("POST /api/login", api.LoginHandler)

	// Protected routes
	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("GET /api/logs", api.GetLogsHandler)
	protectedMux.HandleFunc("GET /api/services", api.GetServicesHandler)
	protectedMux.HandleFunc("GET /api/ws", api.WsHandler)
	protectedMux.HandleFunc("POST /api/change-password", api.ChangePasswordHandler)
	
	// Combine and bind AuthMiddleware
	mux.Handle("/api/", auth.AuthMiddleware(protectedMux))

	// Global CORS middleware
	corsMux := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		mux.ServeHTTP(w, r)
	})

	// Inject Built UI Static Files on root /
	uiFS := backend.LoadUIFileSystem()
	fileServer := http.FileServer(uiFS)
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		fileServer.ServeHTTP(w, r)
	}))

	log.Println("CoreLog HTTP Server booting on port :9091 - v1.1.0 - minimal footprint mode active")
	log.Printf("Monitorando Logs em: %s\n", logPath)
	if err := http.ListenAndServe(":9091", corsMux); err != nil {
		log.Fatal(err)
	}
}
