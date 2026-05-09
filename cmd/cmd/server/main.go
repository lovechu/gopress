package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourorg/gopress/internal/bootstrap"
)

func main() {
	// 1. 加载配置
	cfg, err := bootstrap.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// 2. 初始化应用
	app, err := bootstrap.NewApp(cfg)
	if err != nil {
		log.Fatalf("failed to bootstrap app: %v", err)
	}

	// 3. 启动 HTTP 服务器
	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      app.Router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("[GoPress] server started on :%s (env=%s)", cfg.App.Port, cfg.App.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// 4. 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[GoPress] shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("[GoPress] bye.")
}
