package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/hibiken/asynq"
	_ "github.com/lib/pq"
)

type App struct {
	DB         *sql.DB
	TaskClient *asynq.Client
	TaskServer *asynq.Server
	LokiClient *LokiClient
	Normalizer *LogNormalizer
	Correlator *CorrelationEngine
}

type Alert struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
	Severity  string                 `json:"severity"`
	Message   string                 `json:"message"`
	ProjectID string                 `json:"project_id"`
	RawData   map[string]interface{} `json:"raw_data"`
}

type AnalysisResult struct {
	AlertID           string                 `json:"alert_id"`
	ProjectID         string                 `json:"project_id"`
	CorrelatedLogs    []NormalizedLog        `json:"correlated_logs"`
	UserCorrelations  []UserCorrelation      `json:"user_correlations"`
	EnrichmentData    map[string]interface{} `json:"enrichment_data"`
	AnalysisTimestamp time.Time              `json:"analysis_timestamp"`
	ProcessingTimeMs  int64                  `json:"processing_time_ms"`
}

func main() {
	// Initialize database
	db, err := initDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Initialize task queue
	taskClient := asynq.NewClient(asynq.RedisClientOpt{Addr: "localhost:6379"})
	defer taskClient.Close()

	taskServer := asynq.NewServer(
		asynq.RedisClientOpt{Addr: "localhost:6379"},
		asynq.Config{Concurrency: 10},
	)

	// Initialize components
	lokiClient := NewLokiClient("http://localhost:3100")
	normalizer := NewLogNormalizer()
	correlator := NewCorrelationEngine(db)

	app := &App{
		DB:         db,
		TaskClient: taskClient,
		TaskServer: taskServer,
		LokiClient: lokiClient,
		Normalizer: normalizer,
		Correlator: correlator,
	}

	// Setup task handlers
	taskMux := asynq.NewServeMux()
	taskMux.HandleFunc("alert:analyze", app.handleAlertAnalysis)

	// Start task server
	go func() {
		if err := taskServer.Run(taskMux); err != nil {
			log.Fatal("Failed to start task server:", err)
		}
	}()

	// Setup HTTP routes with Chi
	router := chi.NewRouter()

	// Add middleware
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)

	// Routes
	router.Post("/alerts", app.handleAlert)
	router.Get("/analysis/{alert_id}", app.getAnalysisResult)
	router.Get("/health", app.healthCheck)

	// Start mock data generator
	go app.startMockDataGenerator()

	// Start HTTP server
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		log.Println("Starting server on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed:", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	taskServer.Shutdown()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
}

func (app *App) handleAlert(w http.ResponseWriter, r *http.Request) {
	var alert Alert
	if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	alert.ID = generateID()
	alert.Timestamp = time.Now()

	// Queue analysis task
	task := asynq.NewTask("alert:analyze", mustMarshal(alert))
	if _, err := app.TaskClient.Enqueue(task); err != nil {
		http.Error(w, "Failed to queue analysis", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"alert_id": alert.ID,
		"status":   "queued_for_analysis",
	})
}

func (app *App) getAnalysisResult(w http.ResponseWriter, r *http.Request) {
	alertID := chi.URLParam(r, "alert_id")

	result, err := app.getStoredAnalysisResult(alertID)
	if err != nil {
		http.Error(w, "Analysis result not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (app *App) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func initDB() (*sql.DB, error) {
	connStr := "host=localhost port=5432 user=postgres password=password dbname=soc_analysis sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Create tables
	if err := createTables(db); err != nil {
		return nil, err
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS analysis_results (
			id SERIAL PRIMARY KEY,
			alert_id VARCHAR(255) UNIQUE NOT NULL,
			project_id VARCHAR(255) NOT NULL,
			result_data JSONB NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS user_correlations (
			id SERIAL PRIMARY KEY,
			user_identifier VARCHAR(255) NOT NULL,
			ip_address INET NOT NULL,
			first_seen TIMESTAMP NOT NULL,
			last_seen TIMESTAMP NOT NULL,
			confidence_score FLOAT NOT NULL,
			source_systems TEXT[] NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_user_correlations_user ON user_correlations(user_identifier)`,
		`CREATE INDEX IF NOT EXISTS idx_user_correlations_ip ON user_correlations(ip_address)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %v", err)
		}
	}

	return nil
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func mustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
