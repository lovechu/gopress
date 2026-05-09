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

// @title       GoPress API
// @version     1.0
// @description GoPress 是一个用 Go 语言开发的现代化 CMS 框架，支持多站点、多语言、多租户。

// @contact.name  API Support
// @contact.url   http://www.example.com/support
// @contact.email support@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host     localhost:8080
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in                         header
// @description                输入 Bearer {token}  (注意 Bearer 和 token 之间有一个空格)

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

	// 3. 添加 Swagger 路由 (可选，如果已生成 docs 包)
	// 使用 swag init 命令生成 docs 包后，取消下面的注释
	// url := ginSwagger.URL("/swagger/doc.json")
	// app.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	// 4. 启动 HTTP 服务器
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

	// 5. 优雅关闭
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
