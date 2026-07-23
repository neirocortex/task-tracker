package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"

	_ "github.com/jackc/pgx/v5/stdlib"

	deliveryGrpc "taskTracker/internal/delivery/grpc"
	taskv1 "taskTracker/internal/delivery/grpc/v1"
	deliveryHttp "taskTracker/internal/delivery/http"
	repositoryKafka "taskTracker/internal/repository/kafka"
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
		os.Exit(1)
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

	//kafka
	kafkaAddr := os.Getenv("KAFKA_BROKERS")
	if kafkaAddr == "" {
		kafkaAddr = "kafka:9092"
	}
	kafkaProducer := repositoryKafka.NewTaskNotyfier(kafkaAddr)
	kafkaConsumer := repositoryKafka.NewTaskConsumer(kafkaAddr, kafkaProducer)

	// clean architecture layers dependency injection
	// repository
	taskRepository := repositoryPostgres.NewTaskRepository(db)
	tagRepository := repositoryPostgres.NewTagRepository(db)

	// usecase
	taskCreateCmd := usecase.NewCreateTaskCommand(taskRepository, tagRepository, kafkaProducer)
	taskUpdateCmd := usecase.NewUpdateTaskCommand(taskRepository, tagRepository)
	taskDeleteCmd := usecase.NewDeleteTaskCommand(taskRepository)
	taskGetQ := usecase.NewGetTaskByIDQuery(taskRepository, tagRepository)
	taskListQ := usecase.NewListTasksQuery(taskRepository, tagRepository, taskRepository)
	recordExecCmd := usecase.NewRecordExecutionCommand(taskRepository)

	tagCreateCmd := usecase.NewCreateTagCommand(tagRepository)
	tagUpdateCmd := usecase.NewUpdateTagCommand(tagRepository)
	tagDeleteCmd := usecase.NewDeleteTagCommand(tagRepository)
	tagListQ := usecase.NewGetTagsQuery(tagRepository)

	serverErrors := make(chan error, 2)

	// delivery http
	mux := http.NewServeMux()
	taskHandlerHttp := deliveryHttp.NewTaskHandler(taskCreateCmd, taskUpdateCmd, taskDeleteCmd, taskGetQ, taskListQ, recordExecCmd)
	taskHandlerHttp.RegisterRoutes(mux)
	tagHandlerHttp := deliveryHttp.NewTagHandler(tagCreateCmd, tagUpdateCmd, tagDeleteCmd, tagListQ)
	tagHandlerHttp.RegisterRoutes(mux)

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
	go func() {
		logger.Info("http server listening", "addr", serverAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	// delivery grpc
	taskHandlerGrpc := deliveryGrpc.NewTaskHandler(taskCreateCmd, taskUpdateCmd, taskDeleteCmd, taskGetQ, taskListQ, recordExecCmd)
	grpcServer := grpc.NewServer()
	taskv1.RegisterTaskServiceServer(grpcServer, taskHandlerGrpc)

	go func() {
		logger.Info("gRPC server is starting on :50051")
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			serverErrors <- fmt.Errorf("failed to listen tcp port for grpc: %w", err)
			return
		}

		if err := grpcServer.Serve(lis); err != nil {
			serverErrors <- fmt.Errorf("grpc server failed to serve: %w", err)
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

		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			logger.Info("stopping HTTP server gracefully")
			if err := srv.Shutdown(shutdownCtx); err != nil {
				logger.Error("forced HTTP shutdown due to timeout", "error", err)
				if err := srv.Close(); err != nil {
					logger.Error("failed to close HTTP server listeners", "error", err)
				}
			}
			logger.Info("HTTP server stopped cleanly")
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			logger.Info("stopping gRPC server gracefully")
			grpcServer.GracefulStop()
			logger.Info("gRPC server stopped cleanly")
		}()

		if kafkaProducer != nil {
			wg.Add(1)
			go func() {
				defer wg.Done()
				logger.Info("stopping kafka producer gracefully")
				kafkaProducer.Close()
				logger.Info("kafka producer stopped cleanly")
			}()
		}

		if kafkaConsumer != nil {
			wg.Add(1)
			go func() {
				defer wg.Done()
				logger.Info("stopping kafka consumer gracefully")
				kafkaConsumer.Close()
				logger.Info("kafka consumer stopped cleanly")
			}()
		}

		allServersStopped := make(chan struct{})
		go func() {
			wg.Wait()
			close(allServersStopped)
		}()

		select {
		case <-allServersStopped:
			logger.Info("all servers stopped successfully within timeout")
		case <-shutdownCtx.Done():
			logger.Error("shutdown timed out, forcing application exit")
			grpcServer.Stop()
		}
	}

	logger.Info("application stopped cleanly")
}
