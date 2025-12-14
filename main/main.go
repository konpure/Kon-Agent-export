package main

import (
	"fmt"
	"github.com/konpure/Kon-Agent-export/pkg/api"
	"github.com/konpure/Kon-Agent-export/pkg/config"
	"github.com/konpure/Kon-Agent-export/pkg/processor"
	"github.com/konpure/Kon-Agent-export/pkg/storage"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// load config
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Println("Config loaded successfully:", cfg)

	// init data processor
	dataProcessor := processor.NewDefaultProcessor()
	log.Println("Data processor initialized successfully")

	// init data storage
	dataStorage := storage.NewMemoryStorage(
		cfg.Storage.MaxSize,
		cfg.Storage.ExpireTime,
	)
	log.Println("Data storage initialized successfully")

	// init quic server
	InitQuicServer(dataProcessor, dataStorage)
	log.Println("Quic server initialized successfully")

	// start quic server
	quicAddr := fmt.Sprintf(":%d", cfg.Server.QUICPort)
	go func() {
		if err := StartQuicServer(quicAddr); err != nil {
			log.Fatalf("Failed to start quic server: %v", err)
		}
	}()
	log.Printf("Quic server started successfully on %s", quicAddr)

	// start api server
	httpAddr := fmt.Sprintf(":%d", cfg.Server.HTTPPort)
	apiServer := api.NewAPIServer(dataStorage)
	go func() {
		if err := apiServer.Start(
			httpAddr,
			cfg.Server.ReadTimeout,
			cfg.Server.WriteTimeout,
		); err != nil {
			log.Fatalf("Failed to start api server: %v", err)
		}
	}()
	log.Printf("Api server started successfully on %s", httpAddr)

	// wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// TODO: add graceful shutdown
	log.Println("Server shutting down...")
}
