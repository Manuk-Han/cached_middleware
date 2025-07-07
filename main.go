package main

import (
	"cache/config"
	"cache/core"
	"cache/handler"
	"cache/logger"
	"fmt"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 1. Init logger
	logger.Init()
	defer func(Log *zap.SugaredLogger) {
		err := Log.Sync()
		if err != nil {
			// do nothing
		}
	}(logger.Logger)

	// 2. Load config
	conf, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("❌ config load failed: %v", err)
	}

	// 3. Initialize components
	cacheAdapter, err := core.NewCacheAdapter(conf.Cache)
	if err != nil {
		log.Fatalf("❌ cache adapter init failed: %v", err)
	}

	eventBroker, err := core.NewEventBroker(conf.EventBroker)
	if err != nil {
		log.Fatalf("❌ event broker init failed: %v", err)
	}

	strategy, err := core.NewInvalidationStrategy(conf.Invalidation)
	if err != nil {
		log.Fatalf("❌ strategy init failed: %v", err)
	}

	// 4. Setup services
	cacheService := core.NewCacheService(cacheAdapter, strategy)
	eventListener := core.NewEventListener(eventBroker, cacheService)

	// 5. Setup router
	mux := handler.NewRouter(cacheService, eventBroker)

	// 6. Start listener async
	go eventListener.Start()
	fmt.Println("✅ Event listener started.")

	// 7. Start HTTP server
	go func() {
		fmt.Println("🚀 HTTP server running on :8000")
		if err := http.ListenAndServe(":8000", mux); err != nil {
			log.Fatalf("❌ HTTP server error: %v", err)
		}
	}()

	// 8. Wait for termination
	waitForExit()
}

func waitForExit() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	fmt.Println("\n🛑 Goro shutting down.")
}
