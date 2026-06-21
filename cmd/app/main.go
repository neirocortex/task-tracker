package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	deliveryHttp "taskTracker/internal/delivery/http"
	repositoryPostgres "taskTracker/internal/repository/postgres"
	usecase "taskTracker/internal/usecase"
)

func main() {
	// log init
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	logger.Info("Starting medical task tracker application")

	// db con init
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/med_mis?sslmode=disable"
	}
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		//fail-fast
		logger.Error("failed to open database connection", "error", err)
		return
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("failed to close database connection cleanly", "error", err)
		}
	}()
	// db settings for prod
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// clean architecture layers dependency injection
	// repository
	taskRepository := repositoryPostgres.NewTaskRepository(db)
	tagRepository := repositoryPostgres.NewTagRepository(db)

	// usecase
	taskCreateCmd := usecase.NewCreateTaskCommand(taskRepository, tagRepository)
	taskUpdateCmd := usecase.NewUpdateTaskCommand(taskRepository, tagRepository)
	taskDeleteCmd := usecase.NewDeleteTaskCommand(taskRepository)
	taskGetQ := usecase.NewGetTaskByIDQuery(taskRepository)
	taskListQ := usecase.NewListTasksQuery(taskRepository, tagRepository)

	tagCreateCmd := usecase.NewCreateTagCommand(tagRepository)
	tagUpdateCmd := usecase.NewUpdateTagCommand(tagRepository)
	tagDeleteCmd := usecase.NewDeleteTagCommand(tagRepository)
	tagListQ := usecase.NewGetTagsQuery(tagRepository)

	// delivery
	mux := http.NewServeMux()

	taskHandler := deliveryHttp.NewTaskHandler(taskCreateCmd, taskUpdateCmd, taskDeleteCmd, taskGetQ, taskListQ)
	taskHandler.RegisterRoutes(mux)

	tagHandler := deliveryHttp.NewTagHandler(tagCreateCmd, tagUpdateCmd, tagDeleteCmd, tagListQ)
	tagHandler.RegisterRoutes(mux)

	// start server
	serverAddr := os.Getenv("SERVER_ADDRESS")
	if serverAddr == "" {
		serverAddr = ":8080"
	}
	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("http server listening", "addr", serverAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	// graceful shutdown

	shutdownSignalled := make(chan os.Signal, 1)
	signal.Notify(shutdownSignalled, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Error("server failed to start", "error", err)
		os.Exit(1)

	case sig := <-shutdownSignalled:
		logger.Info("shutdown signal received", "signal", sig.String())

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("forced shutdown due to timeout", "error", err)
			if err := srv.Close(); err != nil {
				logger.Error("failed to close server listeners", "error", err)
			}
		}
	}

	logger.Info("application stopped cleanly")
}
